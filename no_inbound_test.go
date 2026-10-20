package sohocloud

// JFA conformance binding for SUB-no-inbound-requirement. Naming follows
// conformance_test.go: TestInvariant_<ID>_<Name>, with "-" written "_".

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/NTARI-RAND/sohocloud-protocol/coordinator"
	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
	"github.com/NTARI-RAND/sohocloud-protocol/transport/httpjson"
)

// TestInvariant_SUB_no_inbound_requirement_NodeDialsOut binds registry ID
// SUB-no-inbound-requirement: a node contributes "without requiring open
// inbound ports or a static address" — it registers, heartbeats, polls for
// work and reports over outbound connections only, behind NAT/CGNAT and
// dynamic addressing.
//
// SPEC §6 makes this the protocol's shape (the pull model: "the coordinator
// never calls the node"). This binding proves the shape holds:
//
//   - SPEC §6 still says so, and every operation it lists runs node → coord;
//   - the coordinator interfaces are pinned, so a new operation — a possible
//     coordinator → node call — fails until reviewed;
//   - no wire message carries a dialable address: no wire package imports a
//     networking package, and no field is named for an address, host, port,
//     endpoint, URL, IP, callback or webhook;
//   - no package in the module opens a listening socket;
//   - over the reference transport, a node with no listener of its own
//     completes the whole lifecycle — listing, heartbeat, poll, decline,
//     report, fees — on a fresh outbound connection for every call, and its
//     address changes mid-life without the coordinator noticing or caring.
//
// Residual, named: this proves the protocol and its reference transport need
// no inbound path. It does not prove a given node agent never opens one —
// Cloudy's cmd/cloudy-agent does not yet run this loop (its agent → coordinator
// relay is a named next step) — and a real NAT/CGNAT path is a multi-node
// test (Phase 2), not a loopback one.
func TestInvariant_SUB_no_inbound_requirement_NodeDialsOut(t *testing.T) {
	t.Run("SPEC states the pull model", testInboundSpec)
	t.Run("the coordinator surface is pinned", testInboundSurface)
	t.Run("no wire message carries an address", testInboundNoAddress)
	t.Run("nothing in the module listens", testInboundNoListener)
	t.Run("a node with no listener completes its lifecycle", testInboundLifecycle)
}

func testInboundSpec(t *testing.T) {
	data, err := os.ReadFile("SPEC.md")
	if err != nil {
		t.Fatalf("read SPEC.md: %v", err)
	}
	spec := string(data)
	start := strings.Index(spec, "## 6. Coordinator operations")
	if start < 0 {
		t.Fatal("SPEC.md has no §6 Coordinator operations")
	}
	sec := spec[start:]
	if end := strings.Index(sec[1:], "\n## "); end >= 0 {
		sec = sec[:end+1]
	}
	if !strings.Contains(sec, "the coordinator never calls the node") {
		t.Error("SPEC §6 no longer states that the coordinator never calls the node")
	}
	rows := 0
	for _, line := range strings.Split(sec, "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 4 || !strings.Contains(cells[1], "`") {
			continue
		}
		rows++
		dir := strings.TrimSpace(cells[2])
		if dir != "node → coord" && dir != "anyone → coord" {
			t.Errorf("SPEC §6 operation %s runs %q: every operation must be initiated toward the coordinator", strings.TrimSpace(cells[1]), dir)
		}
	}
	if rows < 6 {
		t.Errorf("SPEC §6 lists %d operations, want at least 6 — the check is not reading the table", rows)
	}
}

// coordinatorSurface is every method on the coordinator interfaces, each a
// call a node (or anyone) makes to a coordinator. A new method is a possible
// coordinator → node path and fails until reviewed and added here.
var coordinatorSurface = map[string][]string{
	"Coordinator":             {"Decline", "Fees", "Heartbeat", "PollJobs", "ReportJob", "SubmitListing"},
	"StorageCoordinator":      {"DeclineLease", "PollLeases", "ReleaseLease", "SubmitProof"},
	"KeyLifecycleCoordinator": {"RevokeKey", "RotateKey"},
}

func testInboundSurface(t *testing.T) {
	for name, typ := range map[string]reflect.Type{
		"Coordinator":             reflect.TypeOf((*coordinator.Coordinator)(nil)).Elem(),
		"StorageCoordinator":      reflect.TypeOf((*coordinator.StorageCoordinator)(nil)).Elem(),
		"KeyLifecycleCoordinator": reflect.TypeOf((*coordinator.KeyLifecycleCoordinator)(nil)).Elem(),
	} {
		var got []string
		for i := 0; i < typ.NumMethod(); i++ {
			got = append(got, typ.Method(i).Name)
		}
		if want := coordinatorSurface[name]; !reflect.DeepEqual(got, want) {
			t.Errorf("coordinator.%s methods = %v, want %v: a changed operation must be reviewed as a possible call into the node", name, got, want)
		}
	}

	// The module declares no other interface a node would have to serve.
	allowed := map[string]bool{
		"coordinator.Coordinator": true, "coordinator.StorageCoordinator": true,
		"coordinator.KeyLifecycleCoordinator": true,
		"operator.Signer":                     true, "operator.Verifier": true,
	}
	eachGoFile(t, func(path string, f *ast.File) {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, ok := ts.Type.(*ast.InterfaceType); ok {
				if q := f.Name.Name + "." + ts.Name.Name; !allowed[q] {
					t.Errorf("%s declares interface %s: a new interface is a possible node-side surface the coordinator would call; review it and add it to the allowlist", path, q)
				}
			}
			return true
		})
	})
}

// addressWords are the words that, as a whole word in a field name, name
// something a peer could dial.
var addressWords = map[string]bool{
	"addr": true, "address": true, "host": true, "hostname": true, "port": true,
	"endpoint": true, "url": true, "uri": true, "ip": true, "ipv4": true,
	"ipv6": true, "callback": true, "webhook": true, "listen": true,
	"inbound": true, "socket": true, "dial": true,
}

// words splits a Go identifier into lower-case words: "CallbackURL" →
// [callback url], "GPUMinVRAMMB" → [gpu min vrammb], "JobReport" → [job report].
func words(id string) []string {
	var out []string
	r := []rune(id)
	start := 0
	for i := 1; i <= len(r); i++ {
		if i == len(r) || (unicode.IsUpper(r[i]) && (unicode.IsLower(r[i-1]) ||
			(i+1 < len(r) && unicode.IsLower(r[i+1]) && unicode.IsUpper(r[i-1])))) {
			out = append(out, strings.ToLower(string(r[start:i])))
			start = i
		}
	}
	return out
}

var netPackages = map[string]bool{"net": true, "net/url": true, "net/netip": true, "net/http": true, "crypto/tls": true}

func testInboundNoAddress(t *testing.T) {
	structs := 0
	eachGoFile(t, func(path string, f *ast.File) {
		if strings.HasPrefix(filepath.ToSlash(path), "transport/") {
			return // the transport is how a node dials out; it is not the wire
		}
		for _, imp := range f.Imports {
			if p, _ := strconv.Unquote(imp.Path.Value); netPackages[p] || strings.HasPrefix(p, "net/") {
				t.Errorf("%s imports %s: a wire package has no business with network addresses", path, p)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			st, ok := n.(*ast.StructType)
			if !ok {
				return true
			}
			structs++
			for _, fld := range st.Fields.List {
				for _, name := range fld.Names {
					for _, w := range words(name.Name) {
						if addressWords[w] {
							t.Errorf("%s: field %s names a dialable address: the coordinator never calls the node, so no message may carry where to reach one", path, name.Name)
						}
					}
				}
			}
			return true
		})
	})
	if structs < 15 {
		t.Fatalf("checked only %d struct types — the scan is not looking at the wire", structs)
	}
}

// listenCalls are the standard-library calls that open, or serve on, a
// listening socket.
var listenCalls = map[string]bool{
	"Listen": true, "ListenTCP": true, "ListenUDP": true, "ListenIP": true,
	"ListenUnix": true, "ListenUnixgram": true, "ListenPacket": true,
	"ListenMulticastUDP": true, "ListenAndServe": true, "ListenAndServeTLS": true,
	"Serve": true, "ServeTLS": true, "NewServer": true, "NewUnstartedServer": true,
}

func testInboundNoListener(t *testing.T) {
	eachGoFile(t, func(path string, f *ast.File) {
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if ok && listenCalls[sel.Sel.Name] {
				t.Errorf("%s: %s opens a listening socket: nothing in the protocol may require a node — or anyone — to accept inbound connections", path, sel.Sel.Name)
			}
			return true
		})
	})
}

// eachGoFile parses every non-test Go file in the module.
func eachGoFile(t *testing.T, fn func(path string, f *ast.File)) {
	t.Helper()
	n := 0
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if name := info.Name(); name != "." && (strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		n++
		fn(path, f)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if n < 15 {
		t.Fatalf("parsed only %d Go files — the scan is not looking at the module", n)
	}
}

// recordingCoordinator is a coordinator that records what it was sent and
// hands back one assignment. It holds no way to reach a node: nothing it is
// sent says where a node is.
type recordingCoordinator struct {
	mu       sync.Mutex
	calls    []string
	lastSeq  uint64
	nodeSeen identity.NodeID
}

func (c *recordingCoordinator) note(op string, id identity.NodeID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, op)
	if id != "" {
		c.nodeSeen = id
	}
}

func (c *recordingCoordinator) SubmitListing(_ context.Context, l listing.CapabilityListing) error {
	c.note("SubmitListing", l.NodeID)
	return nil
}

func (c *recordingCoordinator) Heartbeat(_ context.Context, h liveness.Heartbeat) error {
	c.mu.Lock()
	if h.Seq <= c.lastSeq {
		c.mu.Unlock()
		return errSeq
	}
	c.lastSeq = h.Seq
	c.mu.Unlock()
	c.note("Heartbeat", h.NodeID)
	return nil
}

func (c *recordingCoordinator) PollJobs(_ context.Context, id identity.NodeID) ([]employment.Assignment, error) {
	c.note("PollJobs", id)
	return []employment.Assignment{
		{JobID: "job-1", NodeID: id, Spec: employment.JobSpec{Workload: "compute"}},
		{JobID: "job-2", NodeID: id, Spec: employment.JobSpec{Workload: "compute"}},
	}, nil
}

func (c *recordingCoordinator) Decline(_ context.Context, d employment.Decline) error {
	c.note("Decline", d.NodeID)
	return nil
}

func (c *recordingCoordinator) ReportJob(_ context.Context, r employment.JobReport) error {
	c.note("ReportJob", r.NodeID)
	return nil
}

func (c *recordingCoordinator) Fees(context.Context) (fees.FeeDeclaration, error) {
	c.note("Fees", "")
	return fees.FeeDeclaration{CoordinatorID: "coord-test", Terms: fees.Terms{ContributorShareBps: 6500, PlatformFeeBps: 3500}}, nil
}

type seqError struct{}

func (seqError) Error() string { return "non-increasing Seq" }

var errSeq = seqError{}

// nodeClient is a node's handle to the coordinator: the reference transport
// over an HTTP client that dials from localIP and never reuses a connection,
// so every call is a fresh outbound connection from a fresh source port — what
// a node behind NAT looks like from the coordinator's side.
func nodeClient(baseURL, localIP string) *httpjson.Client {
	d := &net.Dialer{Timeout: 5 * time.Second}
	if localIP != "" {
		d.LocalAddr = &net.TCPAddr{IP: net.ParseIP(localIP)}
	}
	return &httpjson.Client{
		BaseURL: baseURL,
		HTTP: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{DialContext: d.DialContext, DisableKeepAlives: true},
		},
	}
}

func testInboundLifecycle(t *testing.T) {
	coord := &recordingCoordinator{}
	var mu sync.Mutex
	sources := map[string]bool{}
	hosts := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		sources[r.RemoteAddr] = true
		if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			hosts[h] = true
		}
		mu.Unlock()
		httpjson.Handler(coord).ServeHTTP(w, r)
	}))
	defer srv.Close()

	ctx := context.Background()
	const node = identity.NodeID("node-behind-nat")
	now := time.Unix(1700000000, 0).UTC()

	// The node, from its first address: register, heartbeat, learn fees, poll
	// for work, decline one job and report the other.
	n := nodeClient(srv.URL, "127.0.0.1")
	steps := []struct {
		name string
		call func(*httpjson.Client) error
	}{
		{"SubmitListing", func(c *httpjson.Client) error {
			return c.SubmitListing(ctx, listing.CapabilityListing{NodeID: node, IssuedAt: now, Seq: 1})
		}},
		{"Heartbeat", func(c *httpjson.Client) error {
			return c.Heartbeat(ctx, liveness.Heartbeat{NodeID: node, SentAt: now, Seq: 1})
		}},
		{"Fees", func(c *httpjson.Client) error { _, err := c.Fees(ctx); return err }},
	}
	for _, s := range steps {
		if err := s.call(n); err != nil {
			t.Fatalf("%s from the node's first address: %v", s.name, err)
		}
	}
	jobs, err := n.PollJobs(ctx, node)
	if err != nil || len(jobs) != 2 {
		t.Fatalf("PollJobs: %d jobs, err %v; want 2 jobs pulled by the node", len(jobs), err)
	}
	if err := n.Decline(ctx, employment.Decline{JobID: jobs[0].JobID, NodeID: node, DeclinedAt: now}); err != nil {
		t.Fatalf("Decline: %v", err)
	}

	// The node's address changes (a new DHCP lease, a CGNAT remap). Where the
	// platform supports a second loopback address it moves to 127.0.0.2;
	// either way every call already arrives from a new source port.
	moved := "127.0.0.2"
	if l, err := net.Listen("tcp", moved+":0"); err != nil {
		t.Logf("cannot bind %s here (%v); the address change is a source-port change only", moved, err)
		moved = "127.0.0.1"
	} else {
		l.Close()
	}
	n = nodeClient(srv.URL, moved)
	if err := n.Heartbeat(ctx, liveness.Heartbeat{NodeID: node, SentAt: now.Add(time.Minute), Seq: 2}); err != nil {
		t.Fatalf("Heartbeat after the address change: %v — a node's standing must not depend on a stable address", err)
	}
	if err := n.ReportJob(ctx, employment.JobReport{JobID: jobs[1].JobID, NodeID: node, StartedAt: now, FinishedAt: now.Add(time.Minute)}); err != nil {
		t.Fatalf("ReportJob after the address change: %v", err)
	}

	coord.mu.Lock()
	calls, seen := strings.Join(coord.calls, ","), coord.nodeSeen
	coord.mu.Unlock()
	if want := "SubmitListing,Heartbeat,Fees,PollJobs,Decline,Heartbeat,ReportJob"; calls != want {
		t.Errorf("coordinator saw %s, want %s", calls, want)
	}
	if seen != node {
		t.Errorf("coordinator knows the node as %q, want %q: identity, not address, names a node", seen, node)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(sources) != 7 {
		t.Errorf("the coordinator saw %d distinct source addresses for 7 calls, want 7: every call must be a fresh outbound connection", len(sources))
	}
	if moved != "127.0.0.1" && len(hosts) != 2 {
		t.Errorf("the coordinator saw hosts %v, want two: the node's address change did not take", hosts)
	}
}
