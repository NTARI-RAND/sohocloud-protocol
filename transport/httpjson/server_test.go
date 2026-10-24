package httpjson_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
	"github.com/NTARI-RAND/sohocloud-protocol/transport/httpjson"
)

type recordingCoordinator struct {
	method string
	input  any
	calls  int
	err    error
	jobs   []employment.Assignment
	fees   fees.FeeDeclaration
}

func (c *recordingCoordinator) record(method string, input any) error {
	c.method, c.input = method, input
	c.calls++
	return c.err
}

func (c *recordingCoordinator) SubmitListing(_ context.Context, v listing.CapabilityListing) error {
	return c.record("listing", v)
}

func (c *recordingCoordinator) Heartbeat(_ context.Context, v liveness.Heartbeat) error {
	return c.record("heartbeat", v)
}

func (c *recordingCoordinator) Decline(_ context.Context, v employment.Decline) error {
	return c.record("decline", v)
}

func (c *recordingCoordinator) ReportJob(_ context.Context, v employment.JobReport) error {
	return c.record("report", v)
}

func (c *recordingCoordinator) PollJobs(_ context.Context, id identity.NodeID) ([]employment.Assignment, error) {
	return c.jobs, c.record("jobs", id)
}

func (c *recordingCoordinator) Fees(_ context.Context) (fees.FeeDeclaration, error) {
	return c.fees, c.record("fees", nil)
}

func coordinatorServer(t *testing.T, coordinator *recordingCoordinator) *httptest.Server {
	t.Helper()
	handler := httpjson.Handler(coordinator)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("POST Content-Type = %q", r.Header.Get("Content-Type"))
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func assertCoordinatorCall(t *testing.T, c *recordingCoordinator, method string, input any) {
	t.Helper()
	if c.calls != 1 || c.method != method || !reflect.DeepEqual(c.input, input) {
		t.Fatalf("coordinator received %d calls: %s %#v; want %s %#v", c.calls, c.method, c.input, method, input)
	}
}

func TestClientHandlerPOSTRoundTrip(t *testing.T) {
	at := time.Date(2026, 1, 2, 3, 4, 5, 123, time.UTC)
	id := identity.NodeID("node +/&?=雪")
	advert := listing.CapabilityListing{
		NodeID: id, Class: listing.ClassStandard, Seq: 7, IssuedAt: at,
		Capacity: listing.Capacity{VCPUs: 4, MemMB: 8192},
		OptIn:    listing.WorkloadOptIn{Compute: true}, Signature: []byte{1, 2, 3},
	}
	heartbeat := liveness.Heartbeat{NodeID: id, SentAt: at, Seq: 8, Signature: []byte{4, 5}}
	decline := employment.Decline{JobID: "job-1", NodeID: id, Reason: employment.DeclineLocalPolicy, DeclinedAt: at}
	report := employment.JobReport{JobID: "job-2", NodeID: id, ExitCode: 1, FailureCause: "timeout", StartedAt: at, FinishedAt: at.Add(time.Second)}
	tests := []struct {
		name   string
		input  any
		invoke func(*httpjson.Client) error
	}{
		{"listing", advert, func(c *httpjson.Client) error { return c.SubmitListing(context.Background(), advert) }},
		{"heartbeat", heartbeat, func(c *httpjson.Client) error { return c.Heartbeat(context.Background(), heartbeat) }},
		{"decline", decline, func(c *httpjson.Client) error { return c.Decline(context.Background(), decline) }},
		{"report", report, func(c *httpjson.Client) error { return c.ReportJob(context.Background(), report) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coordinator := &recordingCoordinator{}
			server := coordinatorServer(t, coordinator)
			if err := tt.invoke(&httpjson.Client{BaseURL: server.URL}); err != nil {
				t.Fatal(err)
			}
			assertCoordinatorCall(t, coordinator, tt.name, tt.input)
		})
	}
}

func TestClientHandlerGETRoundTrip(t *testing.T) {
	at := time.Date(2026, 1, 2, 3, 4, 5, 123, time.UTC)
	id := identity.NodeID("node +/&?=雪")
	jobs := []employment.Assignment{{JobID: "job-3", NodeID: id, OfferedAt: at, Signature: []byte{6, 7}}}
	declaration := fees.FeeDeclaration{CoordinatorID: "coordinator-1", Terms: fees.Terms{ContributorShareBps: 6500, PlatformFeeBps: 3500}, EffectiveAt: at, Seq: 9, Signature: []byte{8, 9}}
	tests := []struct {
		name   string
		input  any
		output any
		invoke func(*httpjson.Client) (any, error)
	}{
		{"jobs", id, jobs, func(c *httpjson.Client) (any, error) { return c.PollJobs(context.Background(), id) }},
		{"fees", nil, declaration, func(c *httpjson.Client) (any, error) { return c.Fees(context.Background()) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coordinator := &recordingCoordinator{jobs: jobs, fees: declaration}
			server := coordinatorServer(t, coordinator)
			output, err := tt.invoke(&httpjson.Client{BaseURL: server.URL, HTTP: server.Client()})
			if err != nil {
				t.Fatal(err)
			}
			assertCoordinatorCall(t, coordinator, tt.name, tt.input)
			if !reflect.DeepEqual(output, tt.output) {
				t.Fatalf("output = %#v; want %#v", output, tt.output)
			}
		})
	}
}

func TestHandlerRejectsInvalidRequests(t *testing.T) {
	type requestCase struct {
		name, method, path, body, message string
		status                            int
	}
	tests := []requestCase{
		{"missing node", http.MethodGet, "/v0/jobs", "", "missing node_id", http.StatusBadRequest},
		{"empty node", http.MethodGet, "/v0/jobs?node_id=", "", "missing node_id", http.StatusBadRequest},
		{"unknown route", http.MethodGet, "/v0/unknown", "", "404 page not found", http.StatusNotFound},
	}
	for _, route := range []string{"listing", "heartbeat", "decline", "report", "jobs", "fees"} {
		tests = append(tests, requestCase{route + " method", http.MethodDelete, "/v0/" + route, "{}", "method not allowed", http.StatusMethodNotAllowed})
	}
	for _, route := range []string{"listing", "heartbeat", "decline", "report"} {
		for _, body := range []string{"", "{", "[]", `{"Signature":"!"}`} {
			tests = append(tests, requestCase{route + " body " + body, http.MethodPost, "/v0/" + route, body, "bad json:", http.StatusBadRequest})
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coordinator := &recordingCoordinator{}
			response := httptest.NewRecorder()
			httpjson.Handler(coordinator).ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))
			if response.Code != tt.status || !strings.Contains(response.Body.String(), tt.message) {
				t.Fatalf("response = %d %q; want %d containing %q", response.Code, response.Body.String(), tt.status, tt.message)
			}
			if coordinator.calls != 0 {
				t.Fatalf("invalid request reached coordinator %d times", coordinator.calls)
			}
		})
	}
}

func TestHandlerReturnsCoordinatorErrors(t *testing.T) {
	for _, route := range []string{"listing", "heartbeat", "decline", "report", "jobs", "fees"} {
		t.Run(route, func(t *testing.T) {
			method, path := http.MethodPost, "/v0/"+route
			if route == "jobs" || route == "fees" {
				method = http.MethodGet
			}
			if route == "jobs" {
				path += "?node_id=node-1"
			}
			coordinator := &recordingCoordinator{err: errors.New("coordinator unavailable")}
			response := httptest.NewRecorder()
			httpjson.Handler(coordinator).ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader("{}")))
			if response.Code != http.StatusInternalServerError || response.Body.String() != "coordinator unavailable\n" {
				t.Fatalf("response = %d %q", response.Code, response.Body.String())
			}
			if coordinator.calls != 1 || coordinator.method != route {
				t.Fatalf("coordinator received %d calls to %q", coordinator.calls, coordinator.method)
			}
		})
	}
}
