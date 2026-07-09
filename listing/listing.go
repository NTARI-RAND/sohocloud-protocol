// Package listing defines a node's signed advertisement of what compute,
// storage, and physical-print work it will accept. A listing is self-issued
// and node-signed: no coordinator speaks for a node's capabilities.
package listing

import (
	"crypto/ed25519"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
)

// ComputeClass is a coarse capability tier. The protocol enumerates the tiers;
// it does NOT define how a coordinator scores or ranks them. Matching policy is
// coordinator-private and deliberately out of the protocol, so a consumer is
// never welded to one coordinator's scoring.
type ComputeClass string

const (
	ClassMicro    ComputeClass = "micro"    // phones, low-power SBCs
	ClassStandard ComputeClass = "standard" // laptops, desktops
	ClassServer   ComputeClass = "server"   // SOHO servers
)

// PrinterKind distinguishes physical-print capabilities.
type PrinterKind string

const (
	PrinterTraditional PrinterKind = "traditional"
	Printer3D          PrinterKind = "threed"
)

// PrinterCapability describes one physical printer the node offers.
type PrinterCapability struct {
	Kind  PrinterKind
	Model string
}

// GPUAPI identifies the compute API through which a GPU is offered. The
// protocol enumerates APIs so a coordinator can avoid offering GPU work a
// node cannot run; it does NOT define scheduling policy (SPEC §5.2).
type GPUAPI string

const (
	GPUVulkan GPUAPI = "vulkan" // cross-platform, incl. Android / Android TV
	GPUNNAPI  GPUAPI = "nnapi"  // Android Neural Networks API
	GPUCUDA   GPUAPI = "cuda"   // NVIDIA desktop/server
	GPUMetal  GPUAPI = "metal"  // Apple silicon
)

// GPUCapability describes one GPU the node offers. Advertising a GPU is the
// opt-in: a listing with no GPUs receives no GPU work, and a node withdraws
// a GPU by omitting it from its next listing. Like everything in a listing
// this is advisory — local enforcement still governs (SPEC §5.1).
type GPUCapability struct {
	API    GPUAPI
	Model  string
	VRAMMB int
}

// Capacity is a point-in-time resource snapshot the node advertises. DiskMB
// is scratch space available to a running job; StorageCommitMB is long-lived
// storage the node commits to hold for the network (shard hosting). The two
// are deliberately distinct so a node can offer either without the other.
// How stored data is encrypted, sharded, and audited is a frontend/agent
// concern outside this protocol: coordination only — the wire never carries
// stored content.
type Capacity struct {
	VCPUs           int
	MemMB           int
	DiskMB          int
	StorageCommitMB int
	PrintQPS        int // print jobs the node will queue; 0 if none
}

// WorkloadOptIn is ADVISORY. The coordinator is NOT a security boundary for
// opt-out. Enforcement lives in the node's local, locally-trusted allowlist;
// these flags only hint the matcher so it does not offer work the node will
// refuse. A node that advertises Print=false here still MUST reject a print
// assignment locally — the wire is never trusted for opt-out enforcement.
type WorkloadOptIn struct {
	Compute bool
	Print   bool
	Storage bool
}

// CapabilityListing is a node's signed capability advertisement.
type CapabilityListing struct {
	NodeID    identity.NodeID
	Class     ComputeClass
	Printers  []PrinterCapability
	GPUs      []GPUCapability
	Capacity  Capacity
	OptIn     WorkloadOptIn
	IssuedAt  time.Time
	Seq       uint64 // strictly monotonic per node; a verifier rejects rollback
	Signature []byte // ed25519 by the node; excluded from CanonicalBytes
}

const domainListing = "sohocloud/listing/v0"

// CanonicalBytes returns the deterministic signing payload with Signature
// excluded. The byte format is documented in SPEC.md.
func (l CapabilityListing) CanonicalBytes() []byte {
	b := canon.New(domainListing)
	b.String(string(l.NodeID))
	b.String(string(l.Class))
	b.Count(len(l.Printers))
	for _, p := range l.Printers {
		b.String(string(p.Kind))
		b.String(p.Model)
	}
	b.Count(len(l.GPUs))
	for _, g := range l.GPUs {
		b.String(string(g.API))
		b.String(g.Model)
		b.Int64(int64(g.VRAMMB))
	}
	b.Int64(int64(l.Capacity.VCPUs))
	b.Int64(int64(l.Capacity.MemMB))
	b.Int64(int64(l.Capacity.DiskMB))
	b.Int64(int64(l.Capacity.StorageCommitMB))
	b.Int64(int64(l.Capacity.PrintQPS))
	b.Bool(l.OptIn.Compute)
	b.Bool(l.OptIn.Print)
	b.Bool(l.OptIn.Storage)
	b.Time(l.IssuedAt)
	b.Uint64(l.Seq)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (l *CapabilityListing) Sign(priv ed25519.PrivateKey) {
	l.Signature = ed25519.Sign(priv, l.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature over the listing.
// pub is the node's public key, resolved out-of-band by the coordinator; the
// protocol does not distribute keys.
func (l CapabilityListing) Verify(pub ed25519.PublicKey) bool {
	return len(l.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, l.CanonicalBytes(), l.Signature)
}
