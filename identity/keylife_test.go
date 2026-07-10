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
