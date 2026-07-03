// Package listing defines a node's signed advertisement of what compute and
// physical-print work it will accept. A listing is self-issued and node-signed:
// no coordinator speaks for a node's capabilities.
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

// Capacity is a point-in-time resource snapshot the node advertises.
type Capacity struct {
	VCPUs    int
	MemMB    int
	DiskMB   int
	PrintQPS int // print jobs the node will queue; 0 if none
}

// WorkloadOptIn is ADVISORY. The coordinator is NOT a security boundary for
// opt-out. Enforcement lives in the node's local, locally-trusted allowlist;
// these flags only hint the matcher so it does not offer work the node will
// refuse. A node that advertises Print=false here still MUST reject a print
// assignment locally — the wire is never trusted for opt-out enforcement.
type WorkloadOptIn struct {
	Compute bool
	Print   bool
}

// CapabilityListing is a node's signed capability advertisement.
type CapabilityListing struct {
	NodeID    identity.NodeID
	Class     ComputeClass
	Printers  []PrinterCapability
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
	b.Int64(int64(l.Capacity.VCPUs))
	b.Int64(int64(l.Capacity.MemMB))
	b.Int64(int64(l.Capacity.DiskMB))
	b.Int64(int64(l.Capacity.PrintQPS))
	b.Bool(l.OptIn.Compute)
	b.Bool(l.OptIn.Print)
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
