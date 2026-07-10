// Package lease defines the storage employment family. Storage is a LEASE,
// not a job: an ongoing obligation to hold a sealed shard, renewed and
// audited over time, so the one-shot employment messages do not fit it.
//
// The protocol speaks ONLY in opaque shard references and quantized size
// classes — never true sizes, never content. How a shard is encrypted,
// padded, and audited for privacy is a frontend concern above this waist
// (Cloudy internal/storage); what crosses the wire is the coordinator's
// signed lease offer, the node's signed exits, and the node's signed proof
// of possession — the fact from which storage metering and payout derive,
// exactly as JobReport is for compute.
package lease

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
)

// StorageLease is a coordinator's signed offer that a node hold one sealed
// shard for a bounded term, fee terms inline so the node sees the split
// before it commits (same discipline as employment.Assignment). Renewal is a
// new StorageLease for the same LeaseID with a strictly higher Seq; a
// verifier rejects rollback exactly as with listings.
type StorageLease struct {
	LeaseID   string
	NodeID    identity.NodeID
	ShardRef  [32]byte // opaque content address; carries no meaning off-manifest
	SizeClass int64    // quantized payload bytes; the wire never sees true sizes
	Fee       fees.Terms
	IssuedAt  time.Time
	ExpiresAt time.Time
	Seq       uint64 // strictly monotonic per LeaseID; renewal counter
	Signature []byte // ed25519 by the coordinator; excluded from CanonicalBytes
}

// DeclineReason mirrors the employment vocabulary; local_policy remains the
// opt-out path — enforcement on the node, never the wire.
type DeclineReason string

const (
	DeclineLocalPolicy DeclineReason = "local_policy"
	DeclineCapacity    DeclineReason = "capacity"
	DeclineUnavailable DeclineReason = "unavailable"
)

// LeaseDecline is a node's signed refusal of an offered lease.
type LeaseDecline struct {
	LeaseID    string
	NodeID     identity.NodeID
	Reason     DeclineReason
	DeclinedAt time.Time
	Signature  []byte // ed25519 by the node
}

// LeaseRelease is a node's signed early exit from a lease it had accepted.
// Sovereignty includes leaving: a node may always stop holding, and the
// signed release marks the moment its metering stops and re-placement of the
// shard becomes the frontend's problem.
type LeaseRelease struct {
	LeaseID    string
	NodeID     identity.NodeID
	ReleasedAt time.Time
	Signature  []byte // ed25519 by the node
}

// ProofChallenge asks the node holding a lease to prove possession of one
// byte range of the sealed shard, salted by a single-use nonce. It is NOT
// signed: a node fetches challenges by polling the coordinator over the
// authenticated channel (pull model — same as assignments), and the
// challenge commits no one to anything. The signed artifact is the response.
type ProofChallenge struct {
	LeaseID  string
	Offset   int64
	Length   int64
	Nonce    [16]byte
	IssuedAt time.Time
}

// ProofResponse is the node's signed answer — self-contained (it restates
// the challenged range and nonce) so it stands alone as a non-repudiable
// metering fact. A verifier MUST reject a response whose Nonce it has seen
// before for that lease: nonces are single-use, which is what defeats
// replaying a recorded answer.
type ProofResponse struct {
	LeaseID     string
	NodeID      identity.NodeID
	Offset      int64
	Length      int64
	Nonce       [16]byte
	Digest      [32]byte
	RespondedAt time.Time
	Signature   []byte // ed25519 by the node
}

// ProofDigest is the one conformant response computation:
// SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset:Offset+Length]).
// Binding the parameters into the hash means an answer for one range never
// doubles as an answer for another. This is byte-identical to the member-side
// expectation Cloudy precomputes at seal time.
func ProofDigest(nonce [16]byte, offset, length int64, rangeBytes []byte) [32]byte {
	h := sha256.New()
	h.Write(nonce[:])
	var params [16]byte
	binary.BigEndian.PutUint64(params[:8], uint64(offset))
	binary.BigEndian.PutUint64(params[8:], uint64(length))
	h.Write(params[:])
	h.Write(rangeBytes)
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

const (
	domainLease        = "sohocloud/lease/v0"
	domainLeaseDecline = "sohocloud/lease-decline/v0"
	domainLeaseRelease = "sohocloud/lease-release/v0"
	domainProof        = "sohocloud/proof/v0"
)

// CanonicalBytes returns the deterministic signing payload, Signature
// excluded. Byte format documented in SPEC.md §4.7.
func (l StorageLease) CanonicalBytes() []byte {
	b := canon.New(domainLease)
	b.String(l.LeaseID)
	b.String(string(l.NodeID))
	b.Bytes(l.ShardRef[:])
	b.Int64(l.SizeClass)
	b.Int64(int64(l.Fee.ContributorShareBps))
	b.Int64(int64(l.Fee.PlatformFeeBps))
	b.Time(l.IssuedAt)
	b.Time(l.ExpiresAt)
	b.Uint64(l.Seq)
	return b.Sum()
}

// Sign sets Signature using the coordinator's private key.
func (l *StorageLease) Sign(priv ed25519.PrivateKey) {
	l.Signature = ed25519.Sign(priv, l.CanonicalBytes())
}

// Verify reports whether Signature is a valid coordinator signature.
func (l StorageLease) Verify(pub ed25519.PublicKey) bool {
	return len(l.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, l.CanonicalBytes(), l.Signature)
}

// CanonicalBytes for the decline (SPEC.md §4.8).
func (d LeaseDecline) CanonicalBytes() []byte {
	b := canon.New(domainLeaseDecline)
	b.String(d.LeaseID)
	b.String(string(d.NodeID))
	b.String(string(d.Reason))
	b.Time(d.DeclinedAt)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (d *LeaseDecline) Sign(priv ed25519.PrivateKey) {
	d.Signature = ed25519.Sign(priv, d.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature.
func (d LeaseDecline) Verify(pub ed25519.PublicKey) bool {
	return len(d.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, d.CanonicalBytes(), d.Signature)
}

// CanonicalBytes for the release (SPEC.md §4.9).
func (r LeaseRelease) CanonicalBytes() []byte {
	b := canon.New(domainLeaseRelease)
	b.String(r.LeaseID)
	b.String(string(r.NodeID))
	b.Time(r.ReleasedAt)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (r *LeaseRelease) Sign(priv ed25519.PrivateKey) {
	r.Signature = ed25519.Sign(priv, r.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature.
func (r LeaseRelease) Verify(pub ed25519.PublicKey) bool {
	return len(r.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, r.CanonicalBytes(), r.Signature)
}

// CanonicalBytes for the proof response (SPEC.md §4.10).
func (p ProofResponse) CanonicalBytes() []byte {
	b := canon.New(domainProof)
	b.String(p.LeaseID)
	b.String(string(p.NodeID))
	b.Int64(p.Offset)
	b.Int64(p.Length)
	b.Bytes(p.Nonce[:])
	b.Bytes(p.Digest[:])
	b.Time(p.RespondedAt)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (p *ProofResponse) Sign(priv ed25519.PrivateKey) {
	p.Signature = ed25519.Sign(priv, p.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature.
func (p ProofResponse) Verify(pub ed25519.PublicKey) bool {
	return len(p.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, p.CanonicalBytes(), p.Signature)
}
