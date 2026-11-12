package httpjson_test

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/transport/httpjson"
)

// Real chunked HTTP must respect budgets without changing signed fee bytes.
func TestClientChunkedSignedResponse(t *testing.T) {
	private := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	declaration := fees.FeeDeclaration{CoordinatorID: "coord", Seq: 9, EffectiveAt: time.Unix(1700000000, 0).UTC()}
	declaration.Sign(private)
	encoded, err := json.Marshal(declaration)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.(http.Flusher).Flush()
		_, _ = w.Write(encoded)
	}))
	defer server.Close()
	client := &httpjson.Client{BaseURL: server.URL, HTTP: server.Client(), MaxResponseBytes: int64(len(encoded))}
	got, err := client.Fees(context.Background())
	if err != nil || !got.Verify(private.Public().(ed25519.PublicKey)) || got.Seq != 9 {
		t.Fatalf("fee=%v err=%v", got, err)
	}
	client.MaxResponseBytes--
	got, err = client.Fees(context.Background())
	if !errors.Is(err, httpjson.ErrResponseTooLarge) || got.Seq != 0 {
		t.Fatalf("fee=%v overflow=%v", got, err)
	}
}

// EOF validation must not make cancellation wait forever for a coordinator.
func TestClientCancellationWhileReadingResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "[]")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	for _, limit := range []int64{0, 64} {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		client := &httpjson.Client{BaseURL: server.URL, HTTP: server.Client(), MaxResponseBytes: limit}
		jobs, err := client.PollJobs(ctx, "node")
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) || jobs != nil {
			t.Fatalf("jobs=%v cancellation=%v", jobs, err)
		}
	}
}
