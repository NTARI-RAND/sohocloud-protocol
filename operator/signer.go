// Package operator defines the frontend-to-coordinator OPERATOR identity layer
// ("Layer C" in the operator model): a rotating multi-key credential a frontend
// (e.g. Cloudy) presents to a coordinator, orthogonal to node identity (§2,
// SPEC §11). An operator holds seven Ed25519 keypairs and signs each
// transmission with two of them (a 2-of-7 anti-substitution/rotation discipline,
// NOT threshold security); the network stores only the seven public keys and
// never holds an operator private key.
//
// This package is Layer C ONLY. The six core coordination messages
// (listing/heartbeat/assignment/decline/jobreport/fee) stay ed25519-direct and
// are NOT routed through the Signer/Verifier seam below.
package operator

import "crypto/ed25519"

// Signer produces a signature over canonical message bytes. It is the
// migration seam for the operator layer's signature primitive: the operator
// credential is Ed25519 today (v0), and the DECIDED end state is a drop-in swap
// to Go stdlib ML-DSA-65 as a whole-set atomic algorithm rotation (a new domain
// tag, e.g. sohocloud/operator/v1-mldsa) once crypto/mldsa lands in the standard
// library (targeted Go 1.27). This is why signing goes through an interface
// rather than calling ed25519 directly.
//
// This seam is Layer-C (operator identity) ONLY. The six core coordination
// messages are signed with ed25519 directly and MUST NOT be routed through it.
// No third-party / PQC dependency is added now: the sole implementation is
// stdlib Ed25519 (Ed25519Signer).
type Signer interface {
	// Sign returns the signature over msg. The caller supplies canonical bytes
	// (canon output); Sign never constructs message bytes itself.
	Sign(msg []byte) []byte
}

// Verifier checks a signature over canonical message bytes against a public
// key. It is the verification half of the same Layer-C-only migration seam as
// Signer; the same ML-DSA-at-Go-1.27 swap applies, and it MUST NOT be used for
// the six core coordination messages.
type Verifier interface {
	// Verify reports whether sig is a valid signature by pub over msg.
	Verify(pub, msg, sig []byte) bool
}

// Ed25519Signer is the v0 Signer: a single Ed25519 private key. An operator
// constructs one Signer per key index it holds. It keeps the private key
// in-process and never exposes it; only public keys ever leave the operator.
type Ed25519Signer struct {
	priv ed25519.PrivateKey
}

// NewEd25519Signer wraps an Ed25519 private key as a Signer.
func NewEd25519Signer(priv ed25519.PrivateKey) Ed25519Signer {
	return Ed25519Signer{priv: priv}
}

// Sign returns the Ed25519 signature over msg.
func (s Ed25519Signer) Sign(msg []byte) []byte {
	return ed25519.Sign(s.priv, msg)
}

// Ed25519Verifier is the v0 Verifier: stdlib Ed25519 verification.
type Ed25519Verifier struct{}

// Verify reports whether sig is a valid Ed25519 signature by pub over msg. It
// rejects a public key or signature of the wrong length before verifying.
func (Ed25519Verifier) Verify(pub, msg, sig []byte) bool {
	return len(pub) == ed25519.PublicKeySize &&
		len(sig) == ed25519.SignatureSize &&
		ed25519.Verify(ed25519.PublicKey(pub), msg, sig)
}

// AlgoEd25519 is the algorithm string bound into an operator message's
// canonical bytes (algorithm binding, blocking a downgrade). v0 permits only
// this value.
const AlgoEd25519 = "ed25519"

// expectedSigLen returns the signature length for a bound algorithm string, or
// 0 if the algorithm is unknown to this version. Per-algorithm length replaces
// a hardcoded 64 so a future ML-DSA rotation reuses the same gate.
func expectedSigLen(algo string) int {
	switch algo {
	case AlgoEd25519:
		return ed25519.SignatureSize // 64
	default:
		return 0
	}
}

// expectedPubLen returns the public-key length for a bound algorithm string, or
// 0 if the algorithm is unknown to this version.
func expectedPubLen(algo string) int {
	switch algo {
	case AlgoEd25519:
		return ed25519.PublicKeySize // 32
	default:
		return 0
	}
}
