package httpjson_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/transport/httpjson"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

func TestClientResponses(t *testing.T) {
	tests := []struct {
		name, body, message string
		status              int
		post                bool
	}{
		{"jobs success", "[]", "", http.StatusOK, false},
		{"post success", "", "", http.StatusNoContent, true},
		{"bad JSON", "{", "unexpected EOF", http.StatusOK, false},
		{"empty JSON", "", "EOF", http.StatusOK, false},
		{"wrong JSON type", "{}", "cannot unmarshal object", http.StatusOK, false},
		{"get failure", "try later", "sohocloud: GET /v0/jobs: status 503: try later", http.StatusServiceUnavailable, false},
		{"post failure", "rejected", "sohocloud: POST /v0/listing: status 400: rejected", http.StatusBadRequest, true},
		{"large error body", strings.Repeat("x", 5000), "sohocloud: GET /v0/jobs: status 502: " + strings.Repeat("x", 4096), http.StatusBadGateway, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader(tt.body)}
			client := &httpjson.Client{
				BaseURL: "https://coordinator.example.test",
				HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: tt.status, Body: body, Header: make(http.Header)}, nil
				})},
			}
			var err error
			if tt.post {
				err = client.SubmitListing(context.Background(), listing.CapabilityListing{})
			} else {
				_, err = client.PollJobs(context.Background(), "node-1")
			}
			assertResponseError(t, err, tt.message)
			if !body.closed {
				t.Fatal("response body was not closed")
			}
			if tt.name == "large error body" && len(err.Error()) != len(tt.message) {
				t.Fatalf("error body was not limited: error length = %d; want %d", len(err.Error()), len(tt.message))
			}
		})
	}
}

func assertResponseError(t *testing.T, err error, message string) {
	t.Helper()
	if message == "" {
		if err != nil {
			t.Fatal(err)
		}
	} else if err == nil || !strings.Contains(err.Error(), message) {
		t.Fatalf("error = %v; want containing %q", err, message)
	}
}

func invokeRequest(c *httpjson.Client, ctx context.Context, method string) error {
	switch method {
	case "listing":
		return c.SubmitListing(ctx, listing.CapabilityListing{})
	case "jobs":
		_, err := c.PollJobs(ctx, "node-1")
		return err
	default:
		_, err := c.Fees(ctx)
		return err
	}
}

func TestClientRequestFailures(t *testing.T) {
	transportError := errors.New("connection failed")
	for _, method := range []string{"listing", "jobs", "fees"} {
		t.Run(method, func(t *testing.T) {
			client := &httpjson.Client{
				BaseURL: "https://coordinator.example.test",
				HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					return nil, transportError
				})},
			}
			if err := invokeRequest(client, context.Background(), method); !errors.Is(err, transportError) {
				t.Fatalf("transport error = %v; want %v", err, transportError)
			}
			client.BaseURL = "://invalid"
			assertResponseError(t, invokeRequest(client, context.Background(), method), "missing protocol scheme")
		})
	}
}

func TestClientPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &httpjson.Client{
		BaseURL: "https://coordinator.example.test",
		HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, r.Context().Err()
		})},
	}
	for _, method := range []string{"listing", "jobs", "fees"} {
		t.Run(method, func(t *testing.T) {
			if err := invokeRequest(client, ctx, method); !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled request error = %v", err)
			}
		})
	}
}

func TestClientRejectsUnencodableTimestamp(t *testing.T) {
	client := &httpjson.Client{
		BaseURL: "https://coordinator.example.test",
		HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			t.Error("unencodable listing reached transport")
			return nil, errors.New("unexpected request")
		})},
	}
	err := client.SubmitListing(context.Background(), listing.CapabilityListing{
		IssuedAt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	var marshalError *json.MarshalerError
	if !errors.As(err, &marshalError) {
		t.Fatalf("timestamp encoding error = %v", err)
	}
}
