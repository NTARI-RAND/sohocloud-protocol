package canon

import (
	"crypto/ed25519"
	"testing"
	"time"
)

func TestVerifySigPanicSafe(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	msg := []byte("hello")
	sig := ed25519.Sign(priv, msg)
	if !VerifySig(pub, msg, sig) {
		t.Fatal("valid signature rejected")
	}
	// Wrong-length key must return false, not panic.
	if VerifySig([]byte{1, 2, 3}, msg, sig) {
		t.Fatal("accepted a malformed public key")
	}
	// Wrong-length signature must return false, not panic.
	if VerifySig(pub, msg, []byte{4, 5}) {
		t.Fatal("accepted a malformed signature")
	}
}

func TestValidTimeAndLatch(t *testing.T) {
	if !ValidTime(time.Unix(1_700_000_000, 0)) {
		t.Fatal("a normal instant reported out of range")
	}
	// Far outside the int64-nanosecond window (year ~5000).
	far := time.Date(5000, 1, 1, 0, 0, 0, 0, time.UTC)
	if ValidTime(far) {
		t.Fatal("an out-of-range instant reported valid")
	}
	b := New("test/domain").String("x").Time(far).Uint64(7)
	if b.Err() == nil {
		t.Fatal("out-of-range Time did not latch an error")
	}
	// In-range time must not latch.
	b2 := New("test/domain").Time(time.Unix(1, 0))
	if b2.Err() != nil {
		t.Fatalf("in-range Time latched an error: %v", b2.Err())
	}
}
