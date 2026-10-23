package httpjson

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
)

func FuzzHandlerJSON(f *testing.F) {
	now := time.Unix(1700000000, 123).UTC()
	seeds := []any{listing.CapabilityListing{NodeID: "node", IssuedAt: now, Seq: 1<<63 + 1}, liveness.Heartbeat{NodeID: "node", SentAt: now, Seq: 1<<63 + 2}, employment.Decline{NodeID: "node", JobID: "job", DeclinedAt: now}, employment.JobReport{NodeID: "node", JobID: "job", StartedAt: now, FinishedAt: now}}
	for i, v := range seeds {
		b, err := json.Marshal(v)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(uint8(i), string(b))
	}
	for _, s := range []string{"", "{", "[]", "null", `{"SentAt":"2500-01-01T00:00:00Z"}`, `{"Seq":18446744073709551615}`, `{"Seq":18446744073709551616}`, `{"Signature":"!"}`} {
		f.Add(uint8(1), s)
	}
	f.Fuzz(func(t *testing.T, kind uint8, body string) {
		paths := []string{"/v0/listing", "/v0/heartbeat", "/v0/decline", "/v0/report"}
		c := &testCoordinator{}
		w := httptest.NewRecorder()
		Handler(c).ServeHTTP(w, httptest.NewRequest(http.MethodPost, paths[int(kind)%len(paths)], strings.NewReader(body)))
		switch w.Code {
		case 204:
			if len(c.received) != 1 {
				t.Fatal("decoded request was lost")
			}
		case 400:
			if len(c.received) != 0 {
				t.Fatal("bad JSON reached coordinator")
			}
		default:
			t.Fatalf("unexpected status %d", w.Code)
		}
		// The reference handler is not a semantic validator/authenticator. Decode
		// success (including null/unknown fields) does not authorize a message.
	})
}

func FuzzClientJSON(f *testing.F) {
	now := time.Unix(1700000000, 123).UTC()
	for i, v := range []any{[]employment.Assignment{{NodeID: "node", JobID: "job", OfferedAt: now}}, fees.FeeDeclaration{CoordinatorID: "coord", EffectiveAt: now, Seq: 1<<63 + 1}} {
		b, err := json.Marshal(v)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(i == 0, true, string(b))
	}
	for _, s := range []string{"", "{", "[]", "null", `{"Seq":18446744073709551616}`, `{"EffectiveAt":"bad-time"}`, strings.Repeat("x", 8192)} {
		f.Add(false, true, s)
		f.Add(true, false, s)
	}
	f.Fuzz(func(t *testing.T, jobs, success bool, payload string) {
		code := 503
		if success {
			code = 200
		}
		body := &trackedBody{Reader: strings.NewReader(payload)}
		c := &Client{BaseURL: "http://coordinator.invalid", HTTP: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: code, Header: make(http.Header), Body: body}, nil
		})}}
		var err error
		if jobs {
			_, err = c.PollJobs(context.Background(), "node")
		} else {
			_, err = c.Fees(context.Background())
		}
		if !body.closed {
			t.Fatal("response body leaked")
		}
		if !success && (err == nil || body.read > 4096) {
			t.Fatalf("HTTP failure lost or read unbounded: read %d, err %v", body.read, err)
		}

	})
}

func TestClientRejectsMalformedResponses(t *testing.T) {
	for _, operation := range []string{"jobs", "fees"} {
		bad := []string{"", "{", `{"Signature":"!"}`, `{"EffectiveAt":"not a timestamp"}`, `{"Seq":18446744073709551616}`}
		if operation == "jobs" {
			bad = []string{"", "{", `{}`, `[{"Signature":"!"}]`, `[{"OfferedAt":"not a timestamp"}]`}
		} else {
			bad = append(bad, `[]`)
		}
		for _, payload := range bad {
			t.Run(operation+"/"+payload, func(t *testing.T) {
				c := &Client{BaseURL: "http://coordinator.invalid", HTTP: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 200, Header: make(http.Header), Body: &trackedBody{Reader: strings.NewReader(payload)}}, nil
				})}}
				if err := clientCalls(c, context.Background())[operation](); err == nil {
					t.Fatal("malformed response accepted")
				}
			})
		}
	}
}
