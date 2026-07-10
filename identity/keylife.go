package identity

import (
	"crypto/ed25519"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
)

// KeyRotation replaces a node's verification key. It is signed by the OLD
// key: possession of the outgoing key is what authorizes naming a successor,
// so enrollment can stop being forever. A verifier MUST validate the
// signature against the key it currently holds for the node, MUST enforce
// strictly monotonic Seq per node (rollback of a rotation is an old-key
// resurrection), and after acceptance MUST verify subsequent node messages
// against NewPublicKey only.
type KeyRotation struct {
	NodeID       NodeID
	NewPublicKey []byte // 32-byte ed25519 public key
	Algo         string // "ed25519"; names the successor's algorithm explicitly
	RotatedAt    time.Time
	Seq          uint64 // strictly monotonic per node across rotations
	Signature    []byte // ed25519 by the OLD (current) key; excluded from canon
}

// KeyRevocation declares a node key dead with no successor. It is signed by
// the revoked key itself; a thief holding the key can also produce this, and
// a verifier MUST honor it anyway — a revocation proves someone with the key
// wants it unusable, and the safe reading of that is always "stop trusting
// it". Re-enrollment after revocation is out-of-band (the same path as first
// enrollment), never a wire message a key thief could forge.
type KeyRevocation struct {
	NodeID           NodeID
	RevokedPublicKey []byte // 32-byte ed25519 public key being killed
	RevokedAt        time.Time
	Seq              uint64 // strictly monotonic per node
	Signature        []byte // ed25519 by the revoked key; excluded from canon
}

const (
	domainRotate = "sohocloud/node-rotate/v0"
	domainRevoke = "sohocloud/node-revoke/v0"
)

// CanonicalBytes returns the deterministic signing payload, Signature
// excluded. Byte format documented in SPEC.md §4.11.
func (k KeyRotation) CanonicalBytes() []byte {
	b := canon.New(domainRotate)
	b.String(string(k.NodeID))
	b.Bytes(k.NewPublicKey)
	b.String(k.Algo)
	b.Time(k.RotatedAt)
	b.Uint64(k.Seq)
	return b.Sum()
}

// Sign sets Signature using the node's OLD (current) private key.
func (k *KeyRotation) Sign(priv ed25519.PrivateKey) {
	k.Signature = ed25519.Sign(priv, k.CanonicalBytes())
}

// Verify reports whether Signature is valid under the node's CURRENT key —
// the one being rotated away from.
func (k KeyRotation) Verify(currentPub ed25519.PublicKey) bool {
	return len(k.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(currentPub, k.CanonicalBytes(), k.Signature)
}

// CanonicalBytes returns the deterministic signing payload, Signature
// excluded. Byte format documented in SPEC.md §4.12.
func (k KeyRevocation) CanonicalBytes() []byte {
	b := canon.New(domainRevoke)
	b.String(string(k.NodeID))
	b.Bytes(k.RevokedPublicKey)
	b.Time(k.RevokedAt)
	b.Uint64(k.Seq)
	return b.Sum()
}

// Sign sets Signature using the private half of the key being revoked.
func (k *KeyRevocation) Sign(priv ed25519.PrivateKey) {
	k.Signature = ed25519.Sign(priv, k.CanonicalBytes())
}

// Verify reports whether Signature is valid under the key being revoked.
func (k KeyRevocation) Verify(revokedPub ed25519.PublicKey) bool {
	return len(k.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(revokedPub, k.CanonicalBytes(), k.Signature)
}
