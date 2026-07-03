// Package fees defines a coordinator's signed, timestamped statement of its fee
// terms. It exists so fees are legible and contestable rather than opaque
// (open problem #7): a node or frontend can read, archive, and challenge the
// terms a coordinator applied, and terms cannot change retroactively without a
// new signed declaration.
package fees

import (
	"crypto/ed25519"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
)

// Terms is a revenue split in basis points. ContributorShareBps and
// PlatformFeeBps SHOULD sum to 10000; a verifier MAY reject terms that do not.
type Terms struct {
	ContributorShareBps int // e.g. 6500 = 65% to the contributor
	PlatformFeeBps      int // e.g. 3500 = 35% to the coordinator
}

// Balanced reports whether the split sums to 100%.
func (t Terms) Balanced() bool {
	return t.ContributorShareBps+t.PlatformFeeBps == 10000
}

// FeeDeclaration is a coordinator's signed fee statement.
type FeeDeclaration struct {
	CoordinatorID string
	Terms         Terms
	EffectiveAt   time.Time
	Seq           uint64
	Signature     []byte // ed25519 by the coordinator
}

const domainFee = "sohocloud/fee/v0"

// CanonicalBytes returns the deterministic signing payload with Signature
// excluded.
func (f FeeDeclaration) CanonicalBytes() []byte {
	b := canon.New(domainFee)
	b.String(f.CoordinatorID)
	b.Int64(int64(f.Terms.ContributorShareBps))
	b.Int64(int64(f.Terms.PlatformFeeBps))
	b.Time(f.EffectiveAt)
	b.Uint64(f.Seq)
	return b.Sum()
}

// Sign sets Signature using the coordinator's private key.
func (f *FeeDeclaration) Sign(priv ed25519.PrivateKey) {
	f.Signature = ed25519.Sign(priv, f.CanonicalBytes())
}

// Verify reports whether Signature is a valid coordinator signature over the
// declaration.
func (f FeeDeclaration) Verify(pub ed25519.PublicKey) bool {
	return len(f.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, f.CanonicalBytes(), f.Signature)
}
