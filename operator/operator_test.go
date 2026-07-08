package operator

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

// sevenKeys derives KeyIndexCount deterministic keypairs from fixed seeds and
// returns the private keys plus the public-key registry a coordinator would
// hold. No crypto/rand: fully reproducible.
func sevenKeys() ([]ed25519.PrivateKey, map[int]KeyRecord) {
	privs := make([]ed25519.PrivateKey, KeyIndexCount)
	km := make(map[int]KeyRecord, KeyIndexCount)
	for i := 0; i < KeyIndexCount; i++ {
		seed := make([]byte, ed25519.SeedSize)
		for j := range seed {
			seed[j] = byte(0xA0 + i)
		}
		priv := ed25519.NewKeyFromSeed(seed)
		privs[i] = priv
		km[i] = KeyRecord{PublicKey: priv.Public().(ed25519.PublicKey), Algo: AlgoEd25519}
	}
	return privs, km
}

func sampleTransmission() OperatorTransmission {
	return OperatorTransmission{
		OperatorID: "cloudy",
		TsUnixNano: 1_700_000_000_000_000_123,
		Nonce:      bytes.Repeat([]byte{0x5a}, MinNonceLen),
		Seq:        7,
		Algo:       AlgoEd25519,
	}
}

func TestTransmissionRoundTrip(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Sign(privs[2], privs[5], 2, 5)
	if err := tx.Verify(km); err != nil {
		t.Fatalf("valid 2-of-7 transmission rejected: %v", err)
	}
}

func TestTransmissionTamperDetected(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Sign(privs[0], privs[3], 0, 3)
	tx.Seq = 8 // mutate a signed field after signing
	if err := tx.Verify(km); err == nil {
		t.Fatal("tampered transmission verified")
	}
}

func TestTransmissionDistinctIndexRequired(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Sign(privs[1], privs[1], 1, 1) // same index twice
	if err := tx.Verify(km); err != ErrSameIndex {
		t.Fatalf("want ErrSameIndex, got %v", err)
	}
}

// TestTransmissionDuplicateKeyRejected exercises the 2-of-7 -> 1-of-1
// degradation guard. Two DISTINCT indices are registered with the SAME public
// key (as a careless registration or a rotation could produce). A holder of the
// single matching private key signs once and submits that one signature under
// both indices. The indices differ (so ErrSameIndex does not fire) and each
// signature verifies against its index's key (which is the same key), so
// without the duplicate-key guard this would pass and collapse 2-of-7 to 1-of-1.
// Verify MUST reject it with ErrDuplicateSigningKey.
func TestTransmissionDuplicateKeyRejected(t *testing.T) {
	privs, km := sevenKeys()
	// Register index 5's slot with index 2's public key: now indices 2 and 5
	// resolve to the same key.
	rec := km[5]
	rec.PublicKey = km[2].PublicKey
	km[5] = rec

	tx := sampleTransmission()
	// Sign with index 2's key for BOTH signatures but claim distinct indices
	// 2 and 5. Both signatures are byte-identical and both verify against the
	// (identical) registered keys.
	tx.Sign(privs[2], privs[2], 2, 5)
	if err := tx.Verify(km); err != ErrDuplicateSigningKey {
		t.Fatalf("want ErrDuplicateSigningKey, got %v", err)
	}
}

// TestRotationDuplicateNewKeyRejected confirms a rotation cannot install a
// NewPublicKey that already exists at another registered index, which would set
// up the duplicate-key degradation above. Rotating in an already-present key at
// a DIFFERENT index is rejected with ErrDuplicateSigningKey.
func TestRotationDuplicateNewKeyRejected(t *testing.T) {
	privs, km := sevenKeys()
	r := OperatorRotation{
		OperatorID:   "cloudy",
		KeyIndex:     4,
		NewPublicKey: km[2].PublicKey, // already registered at index 2
		Algo:         AlgoEd25519,
		TsUnixNano:   1_700_000_050_500_000_000,
		Nonce:        bytes.Repeat([]byte{0x33}, MinNonceLen),
		Seq:          9,
	}
	r.Sign(privs[0], privs[1], 0, 1)
	if err := r.Verify(km); err != ErrDuplicateSigningKey {
		t.Fatalf("want ErrDuplicateSigningKey, got %v", err)
	}
}

// TestRotationSameIndexRefreshAllowed confirms the duplicate-key guard exempts
// the index being replaced: re-registering the same public key at its OWN index
// (a no-op key refresh) is not treated as a duplicate.
func TestRotationSameIndexRefreshAllowed(t *testing.T) {
	privs, km := sevenKeys()
	r := OperatorRotation{
		OperatorID:   "cloudy",
		KeyIndex:     4,
		NewPublicKey: km[4].PublicKey, // same key at its own index
		Algo:         AlgoEd25519,
		TsUnixNano:   1_700_000_050_500_000_000,
		Nonce:        bytes.Repeat([]byte{0x33}, MinNonceLen),
		Seq:          9,
	}
	r.Sign(privs[0], privs[1], 0, 1)
	if err := r.Verify(km); err != nil {
		t.Fatalf("same-index refresh rejected: %v", err)
	}
}

func TestTransmissionNonceMinLength(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Nonce = bytes.Repeat([]byte{0x01}, MinNonceLen-1) // one byte short
	tx.Sign(privs[2], privs[4], 2, 4)
	if err := tx.Verify(km); err != ErrNonceTooShort {
		t.Fatalf("want ErrNonceTooShort, got %v", err)
	}
}

func TestTransmissionAlgorithmBinding(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Sign(privs[0], privs[1], 0, 1)
	// Registry says index 1 is a different algorithm than the signed Algo.
	rec := km[1]
	rec.Algo = "mldsa65"
	km[1] = rec
	if err := tx.Verify(km); err != ErrAlgoMismatch {
		t.Fatalf("want ErrAlgoMismatch, got %v", err)
	}
}

func TestTransmissionUnknownAlgoRejected(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Algo = "mldsa65" // not supported in v0
	tx.Sign(privs[0], privs[1], 0, 1)
	if err := tx.Verify(km); err != ErrUnknownAlgo {
		t.Fatalf("want ErrUnknownAlgo, got %v", err)
	}
}

func TestTransmissionMissingKeyRejected(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	tx.Sign(privs[3], privs[6], 3, 6)
	delete(km, 6) // key not registered / revoked
	if err := tx.Verify(km); err != ErrKeyMissing {
		t.Fatalf("want ErrKeyMissing, got %v", err)
	}
}

func TestTransmissionWrongKeyRejected(t *testing.T) {
	privs, km := sevenKeys()
	tx := sampleTransmission()
	// Sign with index 4's key but claim index 0.
	tx.Sign(privs[4], privs[1], 0, 1)
	if err := tx.Verify(km); err != ErrBadSignature {
		t.Fatalf("want ErrBadSignature, got %v", err)
	}
}

func TestRotationRoundTrip(t *testing.T) {
	privs, km := sevenKeys()
	newSeed := bytes.Repeat([]byte{0x77}, ed25519.SeedSize)
	newPub := ed25519.NewKeyFromSeed(newSeed).Public().(ed25519.PublicKey)
	r := OperatorRotation{
		OperatorID:   "cloudy",
		KeyIndex:     4,
		NewPublicKey: newPub,
		Algo:         AlgoEd25519,
		TsUnixNano:   1_700_000_050_500_000_000,
		Nonce:        bytes.Repeat([]byte{0x33}, MinNonceLen),
		Seq:          9,
	}
	r.Sign(privs[0], privs[1], 0, 1)
	if err := r.Verify(km); err != nil {
		t.Fatalf("valid rotation rejected: %v", err)
	}
	// The new public key is inside the signed bytes: tampering breaks it.
	r.NewPublicKey = bytes.Repeat([]byte{0x00}, ed25519.PublicKeySize)
	if err := r.Verify(km); err == nil {
		t.Fatal("rotation verified after new public key was swapped")
	}
}

func TestConformanceDomainSeparation(t *testing.T) {
	privs, km := sevenKeys()
	challenge := bytes.Repeat([]byte{0x11}, 32)
	c := ConformanceResponse{OperatorID: "cloudy", Challenge: challenge, Algo: AlgoEd25519}
	c.Sign(privs[2], privs[3], 2, 3)
	if err := c.Verify(km); err != nil {
		t.Fatalf("valid conformance response rejected: %v", err)
	}
	// A conformance signature MUST NOT verify as a live transmission: build a
	// transmission whose canonical bytes would collide only if tags matched.
	tx := OperatorTransmission{
		OperatorID: "cloudy",
		Nonce:      challenge, // 32 bytes, satisfies MinNonceLen
		Algo:       AlgoEd25519,
		Idx0:       2,
		Idx1:       3,
		Sig0:       c.Sig0,
		Sig1:       c.Sig1,
	}
	if err := tx.Verify(km); err == nil {
		t.Fatal("conformance signature replayed as a valid transmission")
	}
}

// TestEd25519SeamMatchesDirect confirms the Signer/Verifier seam produces and
// accepts exactly the same bytes as calling ed25519 directly, so routing the
// operator layer through the interface changes nothing observable.
func TestEd25519SeamMatchesDirect(t *testing.T) {
	seed := bytes.Repeat([]byte{0x42}, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	msg := []byte("layer-c seam check")

	var s Signer = NewEd25519Signer(priv)
	sig := s.Sign(msg)
	if !bytes.Equal(sig, ed25519.Sign(priv, msg)) {
		t.Fatal("seam signature differs from direct ed25519.Sign")
	}
	var v Verifier = Ed25519Verifier{}
	if !v.Verify(pub, msg, sig) {
		t.Fatal("seam verifier rejected a valid signature")
	}
}
