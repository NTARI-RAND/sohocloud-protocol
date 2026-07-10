package identity

import (
	"bytes"
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

// Verify reports whether Signature is valid under the node's CURRENT key (the
// one being rotated away from) AND the rotation names a well-formed successor.
// It rejects a rotation whose NewPublicKey is not a 32-byte ed25519 key or
// whose Algo is not "ed25519": otherwise a coordinator that (per SPEC §4.11)
// begins verifying subsequent messages against NewPublicKey would hand a
// malformed key to ed25519.Verify and panic. Validating the successor here is
// the same discipline the operator rotation path already applies.
func (k KeyRotation) Verify(currentPub ed25519.PublicKey) bool {
	return k.Algo == "ed25519" &&
		len(k.NewPublicKey) == ed25519.PublicKeySize &&
		canon.VerifySig(currentPub, k.CanonicalBytes(), k.Signature)
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

// Verify reports whether the revocation is self-signed by the key it names
// AND that key is the one the caller currently trusts for the node. Pass the
// key you currently trust for k.NodeID as trustedPub. Verify returns false
// unless RevokedPublicKey both equals trustedPub and validly signed the
// message.
//
// Binding to trustedPub closes the victim-revocation vector: an attacker can
// self-sign a revocation naming a victim's NodeID with the attacker's own
// key, but RevokedPublicKey is then the attacker's key, not the one the
// coordinator trusts for that node, so bytes.Equal fails and the revocation
// is rejected. A revocation is only ever honored against the key it actually
// kills.
func (k KeyRevocation) Verify(trustedPub ed25519.PublicKey) bool {
	return len(k.RevokedPublicKey) == ed25519.PublicKeySize &&
		bytes.Equal(k.RevokedPublicKey, trustedPub) &&
		canon.VerifySig(k.RevokedPublicKey, k.CanonicalBytes(), k.Signature)
}
