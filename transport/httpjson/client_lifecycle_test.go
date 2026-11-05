package httpjson

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientClosesResponseBodies(t *testing.T) {
	failureBody := strings.Repeat("x", 8192)
	for _, tc := range []struct {
		name, operation string
		status          int
		body            string
		wantErr         bool
	}{
		{"success/jobs", "jobs", 200, `[]`, false},
		{"success/listing", "listing", 200, `[]`, false},
		{"bad-json/jobs", "jobs", 200, `{`, true},
		{"bad-json/listing", "listing", 200, `{`, false},
		{"failure/jobs", "jobs", 503, failureBody, true},
		{"failure/listing", "listing", 503, failureBody, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := &trackedBody{Reader: strings.NewReader(tc.body)}
			c := &Client{BaseURL: "http://coordinator.invalid", HTTP: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: b}, nil
			})}}
			err := clientCalls(c, context.Background())[tc.operation]()
			if (err != nil) != tc.wantErr {
				t.Fatalf("err %v, want error %v", err, tc.wantErr)
			}
			if !b.closed {
				t.Fatal("body not closed")
			}
			if tc.status == 503 && (b.read > 4096 || !strings.Contains(err.Error(), "status 503")) {
				t.Fatalf("unbounded/missing HTTP error: read %d, err %v", b.read, err)
			}
		})
	}
}

func TestClientCancellation(t *testing.T) {
	srv := httptest.NewServer(Handler(&testCoordinator{}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for name, call := range clientCalls(c, ctx) {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, context.Canceled) {
				t.Fatalf("want cancellation, got %v", err)
			}
		})
	}
	// Cancel after the transport has started too, without sleeps or live sockets.
	entered := make(chan struct{})
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	c.HTTP = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	done := make(chan error, 1)
	go func() { _, err := c.Fees(ctx); done <- err }()
	<-entered
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request did not cancel")
	}
}
