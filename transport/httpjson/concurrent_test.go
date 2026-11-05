package httpjson

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
)

type requestMarker struct {
	operation string
	node      identity.NodeID
	job       string
}

// The race run checks concurrent access; distinct markers also catch payload
// swaps or shared request state that would survive a request-count assertion.
func TestConcurrentClientCalls(t *testing.T) {
	c := &testCoordinator{}
	client := inProcessClient(Handler(c))
	want := map[requestMarker]int{{operation: "fees"}: 16}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		node := identity.NodeID(fmt.Sprintf("worker-%d", i))
		for _, operation := range []string{"listing", "heartbeat", "jobs"} {
			want[requestMarker{operation: operation, node: node}] = 1
		}
		want[requestMarker{"decline", node, "declined-" + string(node)}] = 1
		want[requestMarker{"report", node, "finished-" + string(node)}] = 1
		wg.Add(1)
		go func(node identity.NodeID) {
			defer wg.Done()
			for name, call := range markedClientCalls(client, node) {
				if err := call(); err != nil {
					t.Errorf("%s/%s: %v", node, name, err)
				}
			}
		}(node)
	}
	wg.Wait()
	got := make(map[requestMarker]int)
	for _, value := range c.received {
		got[receivedMarker(t, value)]++
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("received request markers = %v, want %v", got, want)
	}
}

func markedClientCalls(c *Client, node identity.NodeID) map[string]func() error {
	ctx := context.Background()
	return map[string]func() error{
		"listing":   func() error { return c.SubmitListing(ctx, listing.CapabilityListing{NodeID: node}) },
		"heartbeat": func() error { return c.Heartbeat(ctx, liveness.Heartbeat{NodeID: node}) },
		"decline": func() error {
			return c.Decline(ctx, employment.Decline{NodeID: node, JobID: "declined-" + string(node)})
		},
		"report": func() error {
			return c.ReportJob(ctx, employment.JobReport{NodeID: node, JobID: "finished-" + string(node)})
		},
		"jobs": func() error { _, err := c.PollJobs(ctx, node); return err },
		"fees": func() error { _, err := c.Fees(ctx); return err },
	}
}

func receivedMarker(t *testing.T, value any) requestMarker {
	t.Helper()
	var marker requestMarker
	switch v := value.(type) {
	case listing.CapabilityListing:
		marker = requestMarker{"listing", v.NodeID, ""}
	case liveness.Heartbeat:
		marker = requestMarker{"heartbeat", v.NodeID, ""}
	case employment.Decline:
		marker = requestMarker{"decline", v.NodeID, v.JobID}
	case employment.JobReport:
		marker = requestMarker{"report", v.NodeID, v.JobID}
	case identity.NodeID:
		marker = requestMarker{"jobs", v, ""}
	case string:
		marker = requestMarker{operation: v}
	default:
		t.Fatalf("unexpected coordinator input %T", value)
	}
	return marker
}
