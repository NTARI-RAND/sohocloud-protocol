// Package liveness defines a node's signed liveness signal.
package liveness

import (
	"crypto/ed25519"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
)

// Heartbeat is a node's signed liveness signal. Seq is strictly monotonic per
// node, so a coordinator (or any future witness) detects replay and rollback.
type Heartbeat struct {
	NodeID    identity.NodeID
	SentAt    time.Time
	Seq       uint64
	Signature []byte // ed25519 by the node
}

const domainHeartbeat = "sohocloud/heartbeat/v0"

// CanonicalBytes returns the deterministic signing payload with Signature
// excluded.
func (h Heartbeat) CanonicalBytes() []byte {
	b := canon.New(domainHeartbeat)
	b.String(string(h.NodeID))
	b.Time(h.SentAt)
	b.Uint64(h.Seq)
	return b.Sum()
}

// Sign sets Signature using the node's private key.
func (h *Heartbeat) Sign(priv ed25519.PrivateKey) {
	h.Signature = ed25519.Sign(priv, h.CanonicalBytes())
}

// Verify reports whether Signature is a valid node signature over the
// heartbeat. pub is the node's public key, resolved out-of-band.
func (h Heartbeat) Verify(pub ed25519.PublicKey) bool {
	return len(h.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, h.CanonicalBytes(), h.Signature)
}
