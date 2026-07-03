// Package coordinator defines the pluggable coordination role. SoHoLINK
// implements it; a frontend MAY run its own implementation or match directly
// against nodes. Nothing in this interface exposes matching policy — how a
// coordinator ranks listings and chooses which node gets which job is
// deliberately private to the implementation, so a consumer is never welded to
// one coordinator's scoring. This is the seam that keeps the coordinator layer
// leaveable (open problem #7).
//
// The model is PULL: the coordinator never calls the node. A node submits
// listings and heartbeats, polls for assignments when it wants work, and
// reports outcomes. There is therefore no Node interface here.
package coordinator

import (
	"context"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
)

// Coordinator is the coordination surface a node and a frontend speak to.
type Coordinator interface {
	// SubmitListing records a node's signed capability advertisement.
	// Implementations MUST verify the node signature and MUST reject a listing
	// whose Seq does not strictly exceed the last one seen for that node.
	SubmitListing(ctx context.Context, l listing.CapabilityListing) error

	// Heartbeat records a node's signed liveness signal, under the same
	// signature-and-monotonic-Seq discipline as SubmitListing.
	Heartbeat(ctx context.Context, h liveness.Heartbeat) error

	// PollJobs returns the assignments currently offered to the calling node.
	// Implementations MUST bind the caller's SPIFFE identity to id via
	// identity.BindsTo before returning anything: 401 if no identity is present,
	// 403 if it does not match.
	PollJobs(ctx context.Context, id identity.NodeID) ([]employment.Assignment, error)

	// Decline records a node's signed refusal of an assignment.
	Decline(ctx context.Context, d employment.Decline) error

	// ReportJob records a node's signed outcome — the fact from which metering
	// and payout derive.
	ReportJob(ctx context.Context, r employment.JobReport) error

	// Fees returns the coordinator's current signed fee declaration.
	Fees(ctx context.Context) (fees.FeeDeclaration, error)
}
