// Package employment defines the job lifecycle messages: a coordinator's signed
// offer (Assignment), a node's signed refusal (Decline), and a node's signed
// outcome (JobReport). The model is PULL — a node fetches work when it wants it
// and reports the result; the coordinator never calls the node.
package employment

import (
	"crypto/ed25519"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
)

// JobSpec describes the work. It is carried opaquely enough to stay
// workload-agnostic; the executor interprets it. Workload and PrinterKind are
// advisory routing hints, not enforcement — opt-out enforcement is local to the
// node (see Decline / listing.WorkloadOptIn).
type JobSpec struct {
	Workload     string // "compute" | "print" | "storage"
	Image        string
	Args         []string
	PrinterKind  string // set for print workloads; empty otherwise
	GPUAPI       string // advisory hint: "vulkan" | "nnapi" | "cuda" | "metal"; empty = no GPU required
	GPUMinVRAMMB int64  // advisory minimum VRAM for GPU work; 0 when GPUAPI is empty
}

// Assignment is a coordinator's signed offer of a specific job to a specific
// node, with the fee terms attached inline so the node sees the split before it
// commits. In the pull model there is no separate accept message: a node
// accepts by acting (reporting) and refuses via Decline.
type Assignment struct {
	JobID     string
	NodeID    identity.NodeID
	Spec      JobSpec
	Fee       fees.Terms
	OfferedAt time.Time
	Signature []byte // ed25519 by the coordinator
}

// DeclineReason enumerates why a node refuses an assignment. LocalPolicy is the
// opt-out path: the node's local allowlist forbids the workload regardless of
// what its advertised OptIn flags said. This is the enforcement point for
// opt-out — on the node, off the wire.
type DeclineReason string

const (
	DeclineLocalPolicy DeclineReason = "local_policy"
	DeclineCapacity    DeclineReason = "capacity"
	DeclineUnavailable DeclineReason = "unavailable"
)

// Decline is a node's signed refusal of an assignment.
type Decline struct {
	JobID      string
	NodeID     identity.NodeID
	Reason     DeclineReason
	DeclinedAt time.Time
	Signature  []byte // ed25519 by the node
}

// JobReport is a node's signed outcome for a job. Its fields mirror the
// executor completion parse already live in SoHoLINK (exit_code, failure_cause,
// tmpfs_exhausted), so the existing completion path maps onto this type
// directly. This signed report is the fact from which metering and payout
// derive.
type JobReport struct {
	JobID          string
	NodeID         identity.NodeID
	ExitCode       int
	FailureCause   string
	TmpfsExhausted bool
	StartedAt      time.Time
	FinishedAt     time.Time
	Signature      []byte // ed25519 by the node
}

const (
	domainAssignment = "sohocloud/assignment/v0"
	domainDecline    = "sohocloud/decline/v0"
	domainJobReport  = "sohocloud/jobreport/v0"
)

// CanonicalBytes returns the deterministic signing payload for the assignment,
// Signature excluded.
func (a Assignment) CanonicalBytes() []byte {
	b := canon.New(domainAssignment)
	b.String(a.JobID)
	b.String(string(a.NodeID))
	b.String(a.Spec.Workload)
	b.String(a.Spec.Image)
	b.Count(len(a.Spec.Args))
	for _, arg := range a.Spec.Args {
		b.String(arg)
	}
	b.String(a.Spec.PrinterKind)
	b.String(a.Spec.GPUAPI)
	b.Int64(a.Spec.GPUMinVRAMMB)
	b.Int64(int64(a.Fee.ContributorShareBps))
	b.Int64(int64(a.Fee.PlatformFeeBps))
	b.Time(a.OfferedAt)
	return b.Sum()
}

// Sign sets Signature using the coordinator's private key.
func (a *Assignment) Sign(priv ed25519.PrivateKey) {
	a.Signature = ed25519.Sign(priv, a.CanonicalBytes())
}

// Verify reports whether Signature is a valid coordinator signature over the
// assignment.
func (a Assignment) Verify(pub ed25519.PublicKey) bool {
	return canon.VerifySig(pub, a.CanonicalBytes(), a.Signature)
}

// CanonicalBytes returns the deterministic signing payload for the decline,
// Signature excluded.
func (d Decline) CanonicalBytes() []byte {
	b := canon.New(domainDecline)
	b.String(d.JobID)
	b.String(string(d.NodeID))
	b.String(string(d.Reason))
	b.Time(d.DeclinedAt)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (d *Decline) Sign(priv ed25519.PrivateKey) {
	d.Signature = ed25519.Sign(priv, d.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature over the decline.
func (d Decline) Verify(pub ed25519.PublicKey) bool {
	return canon.VerifySig(pub, d.CanonicalBytes(), d.Signature)
}

// CanonicalBytes returns the deterministic signing payload for the job report,
// Signature excluded.
func (r JobReport) CanonicalBytes() []byte {
	b := canon.New(domainJobReport)
	b.String(r.JobID)
	b.String(string(r.NodeID))
	b.Int64(int64(r.ExitCode))
	b.String(r.FailureCause)
	b.Bool(r.TmpfsExhausted)
	b.Time(r.StartedAt)
	b.Time(r.FinishedAt)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (r *JobReport) Sign(priv ed25519.PrivateKey) {
	r.Signature = ed25519.Sign(priv, r.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature over the report.
func (r JobReport) Verify(pub ed25519.PublicKey) bool {
	return canon.VerifySig(pub, r.CanonicalBytes(), r.Signature)
}
