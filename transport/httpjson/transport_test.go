package httpjson

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
)

// Each request has its own capture; a mutex lets the same handler serve the
// concurrent-client test without adding a race in the test double itself.
type testCoordinator struct {
	mu       sync.Mutex
	contexts []context.Context
	received []any
	jobs     []employment.Assignment
	fee      fees.FeeDeclaration
	err      error
}

func (c *testCoordinator) record(v any, contexts ...context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.received = append(c.received, v)
	if len(contexts) > 0 {
		c.contexts = append(c.contexts, contexts[0])
	}
	return c.err
}
func (c *testCoordinator) SubmitListing(ctx context.Context, v listing.CapabilityListing) error {
	return c.record(v, ctx)
}
func (c *testCoordinator) Heartbeat(ctx context.Context, v liveness.Heartbeat) error {
	return c.record(v, ctx)
}
func (c *testCoordinator) Decline(ctx context.Context, v employment.Decline) error {
	return c.record(v, ctx)
}
func (c *testCoordinator) ReportJob(ctx context.Context, v employment.JobReport) error {
	return c.record(v, ctx)
}
func (c *testCoordinator) PollJobs(ctx context.Context, id identity.NodeID) ([]employment.Assignment, error) {
	return c.jobs, c.record(id, ctx)
}
func (c *testCoordinator) Fees(ctx context.Context) (fees.FeeDeclaration, error) {
	return c.fee, c.record("fees", ctx)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Uses real client serialization, routing and handler decoding without sockets.
func inProcessClient(h http.Handler) *Client {
	return &Client{BaseURL: "http://coordinator.invalid", HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Result(), nil
	})}}
}

// Catches wrong endpoint/message wiring, lost fields, timestamp precision loss,
// unsigned/altered responses and incorrect node_id escaping in the real pair.
func TestSignedLifecycle(t *testing.T) {
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, 32))
	pub := priv.Public().(ed25519.PublicKey)
	now := time.Unix(1700000000, 123456789).UTC()
	node := identity.NodeID("node +/?&=日本語")
	l := listing.CapabilityListing{NodeID: node, Class: listing.ClassServer,
		Printers: []listing.PrinterCapability{{Kind: listing.Printer3D, Model: "printer"}},
		GPUs:     []listing.GPUCapability{{API: listing.GPUCUDA, Model: "gpu", VRAMMB: 8192}},
		Capacity: listing.Capacity{VCPUs: 8, MemMB: 16384, DiskMB: 1024, StorageCommitMB: 2048, PrintQPS: 2},
		OptIn:    listing.WorkloadOptIn{Compute: true, Print: true, Storage: true}, IssuedAt: now, Seq: 1<<63 + 3}
	l.Sign(priv)
	h := liveness.Heartbeat{NodeID: node, SentAt: now, Seq: 1<<63 + 4}
	h.Sign(priv)
	d := employment.Decline{NodeID: node, JobID: "declined", Reason: employment.DeclineLocalPolicy, DeclinedAt: now}
	d.Sign(priv)
	r := employment.JobReport{NodeID: node, JobID: "finished", ExitCode: 17, FailureCause: "out of memory", TmpfsExhausted: true, StartedAt: now, FinishedAt: now.Add(time.Second)}
	r.Sign(priv)
	a := employment.Assignment{NodeID: node, JobID: "offered", Spec: employment.JobSpec{Workload: "compute", Image: "image:tag", Args: []string{"a b", "日本語"}, PrinterKind: "threed", GPUAPI: "cuda", GPUMinVRAMMB: 4096}, Fee: fees.Terms{ContributorShareBps: 6500, PlatformFeeBps: 3500}, OfferedAt: now}
	a.Sign(priv)
	fee := fees.FeeDeclaration{CoordinatorID: "coord", Terms: a.Fee, EffectiveAt: now, Seq: 1<<63 + 5}
	fee.Sign(priv)
	c := &testCoordinator{jobs: []employment.Assignment{a}, fee: fee}
	srv := httptest.NewServer(Handler(c))
	defer srv.Close()
	client := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	ctx := context.Background()
	for _, call := range []func() error{func() error { return client.SubmitListing(ctx, l) }, func() error { return client.Heartbeat(ctx, h) }, func() error { return client.Decline(ctx, d) }, func() error { return client.ReportJob(ctx, r) }} {
		if err := call(); err != nil {
			t.Fatal(err)
		}
	}
	jobs, err := client.PollJobs(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	gotFee, err := client.Fees(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(jobs, []employment.Assignment{a}) || !jobs[0].Verify(pub) {
		t.Fatalf("assignment changed in transport: %+v", jobs)
	}
	if !reflect.DeepEqual(gotFee, fee) || !gotFee.Verify(pub) {
		t.Fatalf("fee changed in transport: %+v", gotFee)
	}
	if want := []any{l, h, d, r, node, "fees"}; !reflect.DeepEqual(c.received, want) {
		t.Fatalf("coordinator received %#v, want %#v", c.received, want)
	}
	for _, v := range c.received[:4] {
		if !v.(interface{ Verify(ed25519.PublicKey) bool }).Verify(pub) {
			t.Fatalf("signature lost for %T", v)
		}
	}
}

func TestHandlerRejectsBadRequestsBeforeDispatch(t *testing.T) {
	for _, path := range []string{"/v0/listing", "/v0/heartbeat", "/v0/decline", "/v0/report", "/v0/jobs", "/v0/fees"} {
		t.Run(path, func(t *testing.T) {
			c := &testCoordinator{}
			h := Handler(c)
			method := http.MethodGet
			if path == "/v0/jobs" || path == "/v0/fees" {
				method = http.MethodPost
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader("{}")))
			if w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status %d, want 405", w.Code)
			}
			if len(c.received) != 0 {
				t.Fatal("bad method dispatched")
			}
		})
	}
	for _, path := range []string{"/v0/listing", "/v0/heartbeat", "/v0/decline", "/v0/report"} {
		for _, body := range []string{"", "{", "[]", `{"Signature":"not base64!"}`, `{"NodeID":42}`} {
			t.Run(path+"/"+body, func(t *testing.T) {
				c := &testCoordinator{}
				w := httptest.NewRecorder()
				Handler(c).ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
				if w.Code != http.StatusBadRequest || len(c.received) != 0 {
					t.Fatalf("status %d, dispatches %d; want 400 and none", w.Code, len(c.received))
				}
			})
		}
	}
	c := &testCoordinator{}
	w := httptest.NewRecorder()
	Handler(c).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v0/jobs", nil))
	if w.Code != 400 || len(c.received) != 0 {
		t.Fatalf("missing node_id: status %d, dispatches %d", w.Code, len(c.received))
	}
}

func TestCoordinatorErrorsReachClient(t *testing.T) {
	c := &testCoordinator{err: errors.New("coordinator unavailable")}
	client := inProcessClient(Handler(c))
	for name, call := range clientCalls(client, context.Background()) {
		t.Run(name, func(t *testing.T) {
			err := call()
			if err == nil || !strings.Contains(err.Error(), "status 500") || !strings.Contains(err.Error(), "coordinator unavailable") {
				t.Fatalf("lost coordinator error: %v", err)
			}
		})
	}
}

func clientCalls(c *Client, ctx context.Context) map[string]func() error {
	return map[string]func() error{
		"listing":   func() error { return c.SubmitListing(ctx, listing.CapabilityListing{}) },
		"heartbeat": func() error { return c.Heartbeat(ctx, liveness.Heartbeat{}) },
		"decline":   func() error { return c.Decline(ctx, employment.Decline{}) },
		"report":    func() error { return c.ReportJob(ctx, employment.JobReport{}) },
		"jobs":      func() error { _, err := c.PollJobs(ctx, "node"); return err },
		"fees":      func() error { _, err := c.Fees(ctx); return err },
	}
}

type trackedBody struct {
	io.Reader
	closed bool
	read   int
}

func (b *trackedBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += n
	return n, err
}
func (b *trackedBody) Close() error { b.closed = true; return nil }
