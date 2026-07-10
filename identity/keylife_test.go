package identity

import (
	"crypto/ed25519"
	"testing"
	"time"
)

func TestRotationRoundTrip(t *testing.T) {
	oldPub, oldPriv, _ := ed25519.GenerateKey(nil)
	newPub, _, _ := ed25519.GenerateKey(nil)
	k := KeyRotation{
		NodeID:       "node-1",
		NewPublicKey: newPub,
		Algo:         "ed25519",
		RotatedAt:    time.Unix(1_700_000_000, 0),
		Seq:          1,
	}
	k.Sign(oldPriv)
	if !k.Verify(oldPub) {
		t.Fatal("valid rotation rejected under the current key")
	}
}

func TestRotationSuccessorNotSelfAuthorizing(t *testing.T) {
	oldPub, _, _ := ed25519.GenerateKey(nil)
	newPub, newPriv, _ := ed25519.GenerateKey(nil)
	k := KeyRotation{NodeID: "node-1", NewPublicKey: newPub, Algo: "ed25519", RotatedAt: time.Unix(1, 0), Seq: 1}
	k.Sign(newPriv) // the NEW key trying to install itself
	if k.Verify(oldPub) {
		t.Fatal("rotation signed by the successor verified — key theft installs itself")
	}
}

func TestRotationTamperAndSeq(t *testing.T) {
	oldPub, oldPriv, _ := ed25519.GenerateKey(nil)
	newPub, _, _ := ed25519.GenerateKey(nil)
	attacker, _, _ := ed25519.GenerateKey(nil)
	k := KeyRotation{NodeID: "node-1", NewPublicKey: newPub, Algo: "ed25519", RotatedAt: time.Unix(1, 0), Seq: 3}
	k.Sign(oldPriv)
	swapped := k
	swapped.NewPublicKey = attacker // redirect the succession after signing
	if swapped.Verify(oldPub) {
		t.Fatal("rotation with swapped successor verified")
	}
	rollback := k
	rollback.Seq = 4 // replay an old rotation as a newer one
	if rollback.Verify(oldPub) {
		t.Fatal("Seq not covered by rotation signature")
	}
}

func TestRevocationRoundTrip(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	k := KeyRevocation{NodeID: "node-1", RevokedPublicKey: pub, RevokedAt: time.Unix(1, 0), Seq: 1}
	k.Sign(priv)
	if !k.Verify(pub) {
		t.Fatal("valid revocation rejected")
	}
	k.NodeID = "node-2" // retarget the kill after signing
	if k.Verify(pub) {
		t.Fatal("tampered revocation verified")
	}
}

// M3: a rotation naming a malformed successor must be rejected, so a
// coordinator never hands a bad key to ed25519.Verify and panics.
func TestRotationRejectsMalformedSuccessor(t *testing.T) {
	oldPub, oldPriv, _ := ed25519.GenerateKey(nil)
	for _, bad := range []struct {
		name string
		k    KeyRotation
	}{
		{"empty successor", KeyRotation{NodeID: "n1", NewPublicKey: []byte{}, Algo: "ed25519", RotatedAt: time.Unix(1, 0), Seq: 1}},
		{"short successor", KeyRotation{NodeID: "n1", NewPublicKey: make([]byte, 16), Algo: "ed25519", RotatedAt: time.Unix(1, 0), Seq: 1}},
		{"unknown algo", KeyRotation{NodeID: "n1", NewPublicKey: make([]byte, ed25519.PublicKeySize), Algo: "rsa", RotatedAt: time.Unix(1, 0), Seq: 1}},
	} {
		k := bad.k
		k.Sign(oldPriv) // validly signed by the current key
		if k.Verify(oldPub) {
			t.Fatalf("%s: rotation with malformed successor verified", bad.name)
		}
	}
}

// M4: an attacker self-signs a revocation naming a victim's NodeID with the
// attacker's own key. Verified against the key the coordinator TRUSTS for the
// victim, it must fail — the victim cannot be knocked offline by a stranger.
func TestRevocationVictimVectorRejected(t *testing.T) {
	victimTrusted, _, _ := ed25519.GenerateKey(nil)
	attackerPub, attackerPriv, _ := ed25519.GenerateKey(nil)
	k := KeyRevocation{NodeID: "victim", RevokedPublicKey: attackerPub, RevokedAt: time.Unix(1, 0), Seq: 1}
	k.Sign(attackerPriv) // perfectly self-signed by the attacker's own key
	if k.Verify(victimTrusted) {
		t.Fatal("attacker revocation naming a victim verified against the victim's trusted key")
	}
	// And a revocation whose named key isn't the trusted one is rejected even
	// if self-signed correctly.
	if k.Verify(attackerPub) != true {
		t.Fatal("self-consistent revocation of the attacker's OWN key should verify against that same key")
	}
}

// M3 broad: a wrong-length public key must make Verify return false, never
// panic (the pre-fix ed25519.Verify would panic).
func TestVerifyPanicSafeOnBadKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)
	k := KeyRotation{NodeID: "n1", NewPublicKey: make([]byte, ed25519.PublicKeySize), Algo: "ed25519", RotatedAt: time.Unix(1, 0), Seq: 1}
	k.Sign(priv)
	if k.Verify([]byte{1, 2, 3}) {
		t.Fatal("verify accepted a malformed current key")
	}
}
