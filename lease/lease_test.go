package lease

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/fees"
)

func testLease() StorageLease {
	return StorageLease{
		LeaseID:   "lease-1",
		NodeID:    "node-1",
		ShardRef:  [32]byte{0xAA, 0xBB},
		SizeClass: 1 << 20,
		Fee:       fees.Terms{ContributorShareBps: 6500, PlatformFeeBps: 3500},
		IssuedAt:  time.Unix(1_700_000_000, 0),
		ExpiresAt: time.Unix(1_700_000_000, 0).Add(30 * 24 * time.Hour),
		Seq:       1,
	}
}

func TestLeaseRoundTrip(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := testLease()
	l.Sign(priv)
	if !l.Verify(pub) {
		t.Fatal("valid lease signature rejected")
	}
}

func TestLeaseTamperDetected(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := testLease()
	l.Sign(priv)
	l.SizeClass = 64 << 20 // inflate the obligation after signing
	if l.Verify(pub) {
		t.Fatal("tampered lease verified")
	}
}

func TestLeaseRenewalSeqIsSigned(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := testLease()
	l.Sign(priv)
	l.Seq = 2 // replay an old lease as a newer renewal
	if l.Verify(pub) {
		t.Fatal("Seq not covered by lease signature")
	}
}

func TestDeclineAndReleaseRoundTrip(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	d := LeaseDecline{LeaseID: "lease-1", NodeID: "node-1", Reason: DeclineLocalPolicy, DeclinedAt: time.Unix(1, 0)}
	d.Sign(priv)
	if !d.Verify(pub) {
		t.Fatal("valid decline rejected")
	}
	d.Reason = DeclineCapacity
	if d.Verify(pub) {
		t.Fatal("tampered decline verified")
	}
	r := LeaseRelease{LeaseID: "lease-1", NodeID: "node-1", ReleasedAt: time.Unix(2, 0)}
	r.Sign(priv)
	if !r.Verify(pub) {
		t.Fatal("valid release rejected")
	}
	r.LeaseID = "lease-2"
	if r.Verify(pub) {
		t.Fatal("tampered release verified")
	}
}

func TestProofResponseRoundTrip(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	sealed := bytes.Repeat([]byte{0x5C}, 4096)
	nonce := [16]byte{0x01, 0x02}
	digest := ProofDigest(nonce, 128, 256, sealed[128:384])
	p := ProofResponse{
		LeaseID: "lease-1", NodeID: "node-1",
		Offset: 128, Length: 256, Nonce: nonce, Digest: digest,
		RespondedAt: time.Unix(1_700_000_000, 0),
	}
	p.Sign(priv)
	if !p.Verify(pub) {
		t.Fatal("valid proof response rejected")
	}
	p.Digest[0] ^= 0x01 // forge the possession claim after signing
	if p.Verify(pub) {
		t.Fatal("tampered proof response verified")
	}
}

func TestProofDigestBindsParameters(t *testing.T) {
	sealed := bytes.Repeat([]byte{0x42}, 1024) // uniform bytes: ranges look alike
	nonce := [16]byte{0x07}
	a := ProofDigest(nonce, 0, 64, sealed[0:64])
	b := ProofDigest(nonce, 64, 64, sealed[64:128])
	if a == b {
		t.Fatal("digests for distinct ranges collide — parameters not bound")
	}
	other := [16]byte{0x08}
	c := ProofDigest(other, 0, 64, sealed[0:64])
	if a == c {
		t.Fatal("digest ignores nonce — precomputable")
	}
}

// L5: ProofOverShard bounds-checks before slicing so a malicious challenge
// cannot panic a node.
func TestProofOverShardBounds(t *testing.T) {
	sealed := bytes.Repeat([]byte{0x5C}, 4096)
	nonce := [16]byte{0x01}
	ok := ProofChallenge{LeaseID: "l1", Offset: 128, Length: 256, Nonce: nonce}
	got, err := ProofOverShard(sealed, ok)
	if err != nil {
		t.Fatalf("in-range challenge errored: %v", err)
	}
	if got != ProofDigest(nonce, 128, 256, sealed[128:384]) {
		t.Fatal("ProofOverShard digest differs from ProofDigest for the same range")
	}
	for _, bad := range []ProofChallenge{
		{Offset: -1, Length: 10},
		{Offset: 0, Length: 0},
		{Offset: 4000, Length: 200}, // runs past the end
		{Offset: 1 << 62, Length: 1 << 62},
	} {
		if _, err := ProofOverShard(sealed, bad); !errors.Is(err, ErrChallengeRange) {
			t.Fatalf("out-of-range challenge %+v: got %v, want ErrChallengeRange", bad, err)
		}
	}
}

func TestWrongKeyRejected(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)
	otherPub, _, _ := ed25519.GenerateKey(nil)
	l := testLease()
	l.Sign(priv)
	if l.Verify(otherPub) {
		t.Fatal("lease verified under the wrong public key")
	}
}
