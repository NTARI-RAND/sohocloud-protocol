package httpjson_test

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/NTARI-RAND/sohocloud-protocol/transport/httpjson"
)

func responseClient(body io.ReadCloser, length int64) *httpjson.Client {
	return &httpjson.Client{BaseURL: "https://coordinator.example.test", HTTP: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body, ContentLength: length, Header: make(http.Header)}, nil
	})}}
}

// Omitting the size check must fail at both declared and streamed boundaries.
func TestClientSuccessBudget(t *testing.T) {
	for method, value := range map[string]string{"jobs": "[]", "fees": "{}", "listing": ""} {
		for _, tc := range []struct {
			name, suffix  string
			limit, length int64
			wantError     bool
			maxRead       int
		}{
			{"exact", "  ", int64(len(value) + 2), -1, false, len(value) + 2},
			{"one over", "   ", int64(len(value) + 2), -1, true, len(value) + 3},
			{"declared over", strings.Repeat(" ", 100), 2, 102, true, 0},
			{"streamed over", strings.Repeat(" ", 10000), 8, -1, true, 9},
			{"zero unlimited", strings.Repeat(" ", 10000), 0, -1, false, 20000},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				reader := &countedReader{Reader: strings.NewReader(value + tc.suffix)}
				body := &trackedBody{Reader: reader}
				client := responseClient(body, tc.length)
				client.MaxResponseBytes = tc.limit
				err := invokeRequest(client, context.Background(), method)
				if (err != nil) != tc.wantError {
					t.Fatalf("error = %v; want error = %v", err, tc.wantError)
				}
				if reader.bytes > tc.maxRead {
					t.Fatalf("read %d bytes; maximum %d", reader.bytes, tc.maxRead)
				}
				if !body.closed {
					t.Fatal("response not closed")
				}
			})
		}
	}
}

type countedReader struct {
	io.Reader
	bytes int
}

func (r *countedReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.bytes += n
	return n, err
}

// A valid prefix followed by failure must never escape as a usable result.
func TestClientResponseReadFailure(t *testing.T) {
	failure := errors.New("response interrupted")
	for _, limit := range []int64{0, 2, 64} {
		body := &trackedBody{Reader: io.MultiReader(strings.NewReader("[]"), failedReader{failure})}
		client := responseClient(body, -1)
		client.MaxResponseBytes = limit
		jobs, err := client.PollJobs(context.Background(), "node")
		if !errors.Is(err, failure) || jobs != nil || !body.closed {
			t.Fatalf("jobs=%v err=%v closed=%v", jobs, err, body.closed)
		}
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

// Overflow has priority over malformed JSON and never exposes payload bytes.
func TestClientOverflowBeforeDecoding(t *testing.T) {
	reader := &countedReader{Reader: strings.NewReader("{secret" + strings.Repeat("!", 1<<20))}
	body := &trackedBody{Reader: reader}
	client := responseClient(body, -1)
	client.MaxResponseBytes = 32
	_, err := client.Fees(context.Background())
	if !errors.Is(err, httpjson.ErrResponseTooLarge) || reader.bytes != 33 || !body.closed {
		t.Fatalf("err=%v bytes=%d closed=%v", err, reader.bytes, body.closed)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatal("overflow exposed payload")
	}
}

func TestClientLargeBudgetAndUnknownFields(t *testing.T) {
	body := &trackedBody{Reader: strings.NewReader(`[{"JobID":"job","future":{"a":1}}]`)}
	client := responseClient(body, -1)
	client.MaxResponseBytes = math.MaxInt64
	jobs, err := client.PollJobs(context.Background(), "node")
	if err != nil || len(jobs) != 1 || jobs[0].JobID != "job" {
		t.Fatalf("jobs=%v err=%v", jobs, err)
	}
}
