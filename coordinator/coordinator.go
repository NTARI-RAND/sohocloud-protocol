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
	"github.com/NTARI-RAND/sohocloud-protocol/lease"
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

// StorageCoordinator is the OPTIONAL storage-lease surface. It is a separate
// capability interface, not new methods on Coordinator, so a coordinator can
// adopt the compute/print protocol without implementing storage yet — a
// consumer discovers support by type assertion and treats its absence as
// "this coordinator does not lease storage", never as an error. The same
// pull model and SPIFFE-binding discipline apply throughout.
type StorageCoordinator interface {
	// PollLeases returns the storage leases currently offered to the calling
	// node, and the open proof challenges against its held leases. Identity
	// binding rules are identical to PollJobs.
	PollLeases(ctx context.Context, id identity.NodeID) ([]lease.StorageLease, []lease.ProofChallenge, error)

	// DeclineLease records a node's signed refusal of an offered lease.
	DeclineLease(ctx context.Context, d lease.LeaseDecline) error

	// ReleaseLease records a node's signed early exit from a held lease.
	ReleaseLease(ctx context.Context, r lease.LeaseRelease) error

	// SubmitProof records a node's signed proof of possession — the fact from
	// which storage metering and payout derive. Implementations MUST reject a
	// response whose (LeaseID, Nonce) was seen before: nonces are single-use.
	SubmitProof(ctx context.Context, p lease.ProofResponse) error
}

// KeyLifecycleCoordinator is the OPTIONAL node-key rotation surface, split
// out for the same staged-adoption reason as StorageCoordinator.
type KeyLifecycleCoordinator interface {
	// RotateKey verifies k against the node's CURRENT key, enforces strictly
	// monotonic Seq, and thereafter trusts only k.NewPublicKey for the node.
	RotateKey(ctx context.Context, k identity.KeyRotation) error

	// RevokeKey kills a node key with no successor. Implementations MUST
	// honor a validly signed revocation unconditionally; re-enrollment is
	// out-of-band.
	RevokeKey(ctx context.Context, k identity.KeyRevocation) error
}
