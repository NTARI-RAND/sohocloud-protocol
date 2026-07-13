package operator

import (
	"bytes"
	"crypto/ed25519"
	"errors"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
)

// MinNonceLen is the minimum length of the per-transmission nonce. A shorter or
// absent nonce is rejected; an empty nonce is NEVER treated as "skip the replay
// cache" (FIX: crypto-7).
const MinNonceLen = 16

// KeyIndexCount is the number of keypairs an operator holds. Indices are
// 0..KeyIndexCount-1. Each transmission is signed with two DISTINCT indices
// (the 2-of-7 discipline).
const KeyIndexCount = 7

// Domain tags. Each is distinct so a signature over one operator message type
// can never be replayed as another — in particular a conformance-challenge
// signature can never be replayed as a live operator transmission.
const (
	domainTransmission = "sohocloud/operator/v0"        // v0 = Ed25519-only
	domainRotation     = "sohocloud/operator-rotate/v0" // authorizes a key swap
	domainConformance  = "sohocloud/operator-conformance/v0"
)

// Errors returned by the operator Verify paths. They are distinguishable so a
// caller (coordinator middleware) can map each to the right rejection.
var (
	ErrSameIndex     = errors.New("operator: the two key indices must differ")
	ErrNonceTooShort = errors.New("operator: nonce absent or shorter than MinNonceLen")
	ErrUnknownAlgo   = errors.New("operator: unknown or unsupported algorithm")
	ErrAlgoMismatch  = errors.New("operator: signed algorithm does not match the registered key at that index")
	ErrKeyMissing    = errors.New("operator: no registered key at a signing index")
	ErrBadSignature  = errors.New("operator: a signature failed to verify")
	ErrBadSigLength  = errors.New("operator: a signature has the wrong length for the bound algorithm")
	ErrIndexRange    = errors.New("operator: a key index is out of range")
	// ErrDuplicateSigningKey is returned when the two signing indices, though
	// distinct as indices, resolve to the SAME registered public key. Without
	// this check the 2-of-7 anti-substitution property silently collapses to
	// 1-of-1: one private key can produce both signatures (see SPEC §11.0
	// rule 5a). Enforced at verify time so the discipline is self-enforcing and
	// does not depend on an unstated out-of-band registration invariant.
	ErrDuplicateSigningKey = errors.New("operator: the two signing indices resolve to the same public key")
)

// KeyRecord is a registered operator public key at one index. The network
// stores only these; it never holds an operator private key. Algo pins the
// algorithm expected at this index (algorithm binding): a transmission whose
// signed Algo does not equal this value is rejected. An operator's active set
// MUST NOT mix algorithms, and the seven registered public keys MUST be
// pairwise distinct — Verify enforces the latter at signing time (two indices
// resolving to the same key are rejected as ErrDuplicateSigningKey) so 2-of-7
// cannot silently degrade to 1-of-1.
type KeyRecord struct {
	PublicKey []byte // raw 32-byte Ed25519 for v0
	Algo      string // e.g. AlgoEd25519
}

// OperatorTransmission is a frontend's authenticated transmission to a
// coordinator, signed with two of the operator's seven keys over the SAME
// canonical bytes. The two signatures are carried alongside and are EXCLUDED
// from the canonical bytes (as with every signed message, SPEC §3).
//
// TsUnixNano is a raw int64 UTC Unix-nanosecond value, deliberately NOT a
// time.Time: the value the coordinator range-checks is the exact int64 the
// canonical bytes are built from, with no time.Time round-trip between decode
// and recompute (FIX: crypto-2).
type OperatorTransmission struct {
	OperatorID string
	TsUnixNano int64  // raw int64 UnixNano; canon signs this exact value
	Nonce      []byte // >= MinNonceLen bytes, MUST be present
	Seq        uint64
	Algo       string // bound into the signed bytes (algorithm binding)
	Idx0, Idx1 int    // the two signing key indices; MUST differ
	Sig0, Sig1 []byte // over CanonicalBytes; excluded from those bytes
}

// CanonicalBytes returns the deterministic signing payload with the two
// signatures excluded. Field order is fixed by SPEC §11.1 and MUST match the
// operator model: OperatorID, TsUnixNano, Nonce, Seq, Algo, Idx0, Idx1.
func (t OperatorTransmission) CanonicalBytes() []byte {
	b := canon.New(domainTransmission)
	b.String(t.OperatorID)
	b.Int64(t.TsUnixNano)
	b.Bytes(t.Nonce)
	b.Uint64(t.Seq)
	b.String(t.Algo)
	b.Uint64(uint64(t.Idx0))
	b.Uint64(uint64(t.Idx1))
	return b.Sum()
}

// Sign fills Sig0 and Sig1 by signing the canonical bytes with the two private
// keys, and records their indices. priv0 corresponds to idx0, priv1 to idx1.
// The caller MUST pass distinct indices; Verify enforces distinctness on the
// far side, and Sign records whatever indices it is given.
//
// Sign takes exactly two private keys (2-of-7). It stores no key anywhere: the
// keys live only in the caller's memory. Only public keys are ever registered
// with the network.
func (t *OperatorTransmission) Sign(priv0, priv1 ed25519.PrivateKey, idx0, idx1 int) {
	t.Idx0, t.Idx1 = idx0, idx1
	if t.Algo == "" {
		t.Algo = AlgoEd25519
	}
	msg := t.CanonicalBytes()
	t.Sig0 = ed25519.Sign(priv0, msg)
	t.Sig1 = ed25519.Sign(priv1, msg)
}

// Verify checks the transmission against a map of registered public keys keyed
// by index. It enforces, in order: nonce present and >= MinNonceLen; a
// supported Algo; distinct indices; indices in range; a key registered at each
// index whose Algo matches the signed Algo (algorithm binding); each signature
// the right length for the algorithm; and finally both signatures valid over
// the recomputed canonical bytes.
//
// Verify takes ONLY public keys (a KeyRecord map): nothing here stores or
// requires an operator private key. It returns nil on success or a specific
// error on the first failing check.
//
// Verify is necessary but NOT sufficient on its own: it performs no replay,
// nonce-set, or Seq-window check. Per SPEC §11.0 those are a coordinator-side
// obligation (a durable, fail-closed sliding-window Seq + nonce set scoped per
// (operator, coordinator)) layered on top. A byte-identical resubmission passes
// Verify; the coordinator middleware, not this method, must reject the replay.
func (t OperatorTransmission) Verify(keymap map[int]KeyRecord) error {
	if len(t.Nonce) < MinNonceLen {
		return ErrNonceTooShort
	}
	if expectedSigLen(t.Algo) == 0 {
		return ErrUnknownAlgo
	}
	return verifyPair(keymap, t.Idx0, t.Idx1, t.Algo, t.CanonicalBytes(), t.Sig0, t.Sig1)
}

// verifyAt checks one signature at one index: range, key presence, per-index
// algorithm binding, signature length, and the signature itself.
func verifyAt(keymap map[int]KeyRecord, idx int, algo string, msg, sig []byte) error {
	if idx < 0 || idx >= KeyIndexCount {
		return ErrIndexRange
	}
	rec, ok := keymap[idx]
	if !ok {
		return ErrKeyMissing
	}
	if rec.Algo != algo {
		return ErrAlgoMismatch
	}
	if len(sig) != expectedSigLen(algo) {
		return ErrBadSigLength
	}
	if len(rec.PublicKey) != expectedPubLen(algo) {
		return ErrKeyMissing
	}
	if !(Ed25519Verifier{}).Verify(rec.PublicKey, msg, sig) {
		return ErrBadSignature
	}
	return nil
}

// verifyPair enforces the 2-of-7 discipline for a message: distinct indices,
// distinct registered PUBLIC KEYS at those indices (not merely distinct
// indices), and both signatures valid at their indices over msg. Rejecting a
// duplicate key is what keeps 2-of-7 from silently degrading to 1-of-1 when a
// careless registration or rotation installs the same public key at two indices
// (SPEC §11.0 rule 3 and 5a). The key-distinctness check runs BEFORE signature
// verification so a duplicate is reported as ErrDuplicateSigningKey regardless
// of whether the signatures would have verified.
func verifyPair(keymap map[int]KeyRecord, idx0, idx1 int, algo string, msg, sig0, sig1 []byte) error {
	if idx0 == idx1 {
		return ErrSameIndex
	}
	if err := distinctKeysAt(keymap, idx0, idx1); err != nil {
		return err
	}
	if err := verifyAt(keymap, idx0, algo, msg, sig0); err != nil {
		return err
	}
	if err := verifyAt(keymap, idx1, algo, msg, sig1); err != nil {
		return err
	}
	return nil
}

// distinctKeysAt reports ErrDuplicateSigningKey when the registered public keys
// at idx0 and idx1 are byte-identical. It only checks keys that are actually
// present; a missing key at either index is left for verifyAt to report as
// ErrKeyMissing (with a range check first), so error precedence stays: range,
// then duplicate-key, then per-index resolution.
func distinctKeysAt(keymap map[int]KeyRecord, idx0, idx1 int) error {
	if idx0 < 0 || idx0 >= KeyIndexCount || idx1 < 0 || idx1 >= KeyIndexCount {
		return ErrIndexRange
	}
	rec0, ok0 := keymap[idx0]
	rec1, ok1 := keymap[idx1]
	if ok0 && ok1 && bytes.Equal(rec0.PublicKey, rec1.PublicKey) {
		return ErrDuplicateSigningKey
	}
	return nil
}

// OperatorRotation authorizes swapping in a NEW public key at KeyIndex. The
// operator (never the network) generates the new keypair; this message binds
// the new public key into signed bytes so neither a MITM of the out-of-band
// registration nor a compromised admin channel can inject key material
// (FIX: crypto-4). It is signed with two CURRENT keys (Idx0, Idx1), exactly
// like a transmission.
type OperatorRotation struct {
	OperatorID   string
	KeyIndex     int    // the index whose key is being replaced (0..KeyIndexCount-1)
	NewPublicKey []byte // the operator-generated replacement public key
	Algo         string // algorithm of the new key AND of the two authorizing signatures
	TsUnixNano   int64
	Nonce        []byte // >= MinNonceLen bytes
	Seq          uint64
	Idx0, Idx1   int    // two CURRENT authorizing key indices; MUST differ
	Sig0, Sig1   []byte // over CanonicalBytes; excluded
}

// CanonicalBytes returns the deterministic signing payload with the two
// signatures excluded. Field order is fixed by SPEC §11.2: OperatorID,
// KeyIndex, NewPublicKey, Algo, TsUnixNano, Nonce, Seq, Idx0, Idx1.
func (r OperatorRotation) CanonicalBytes() []byte {
	b := canon.New(domainRotation)
	b.String(r.OperatorID)
	b.Uint64(uint64(r.KeyIndex))
	b.Bytes(r.NewPublicKey)
	b.String(r.Algo)
	b.Int64(r.TsUnixNano)
	b.Bytes(r.Nonce)
	b.Uint64(r.Seq)
	b.Uint64(uint64(r.Idx0))
	b.Uint64(uint64(r.Idx1))
	return b.Sum()
}

// Sign fills Sig0 and Sig1 with the two CURRENT authorizing keys.
func (r *OperatorRotation) Sign(priv0, priv1 ed25519.PrivateKey, idx0, idx1 int) {
	r.Idx0, r.Idx1 = idx0, idx1
	if r.Algo == "" {
		r.Algo = AlgoEd25519
	}
	msg := r.CanonicalBytes()
	r.Sig0 = ed25519.Sign(priv0, msg)
	r.Sig1 = ed25519.Sign(priv1, msg)
}

// Verify checks the rotation authorization against the CURRENT registered keys.
// Same discipline as OperatorTransmission.Verify, plus a well-formed new public
// key: it rejects a NewPublicKey whose length is wrong for Algo. It does NOT
// itself perform the swap; a coordinator applies the registry change only after
// this returns nil.
func (r OperatorRotation) Verify(keymap map[int]KeyRecord) error {
	if len(r.Nonce) < MinNonceLen {
		return ErrNonceTooShort
	}
	if expectedSigLen(r.Algo) == 0 {
		return ErrUnknownAlgo
	}
	if r.KeyIndex < 0 || r.KeyIndex >= KeyIndexCount {
		return ErrIndexRange
	}
	if len(r.NewPublicKey) != expectedPubLen(r.Algo) {
		return ErrUnknownAlgo
	}
	// Defense-in-depth against the 2-of-7 -> 1-of-1 degradation: refuse a
	// rotation that would install a NewPublicKey already registered at another
	// index. This makes the "no duplicate keys" invariant machine-enforced at
	// the moment a coordinator would apply the swap, not just at later verify
	// time (SPEC §11.0 rule 5a, §11.2). The index being replaced is exempt: a
	// rotation MAY re-register the same key at its own index (a no-op refresh).
	// The same loop enforces the no-mixed-algorithms invariant (SPEC §11.2): the
	// new key's Algo (== r.Algo) MUST match every OTHER registered key's algo,
	// so a rotation can never introduce a heterogeneous, mixed-strength set.
	// Two passes so the returned error is deterministic regardless of Go's map
	// iteration order: duplicate-key always takes precedence over algo-mismatch
	// when both are present, rather than whichever the range happened to hit
	// first.
	for idx, rec := range keymap {
		if idx == r.KeyIndex {
			continue
		}
		if bytes.Equal(rec.PublicKey, r.NewPublicKey) {
			return ErrDuplicateSigningKey
		}
	}
	for idx, rec := range keymap {
		if idx == r.KeyIndex {
			continue
		}
		if rec.Algo != r.Algo {
			return ErrAlgoMismatch
		}
	}
	return verifyPair(keymap, r.Idx0, r.Idx1, r.Algo, r.CanonicalBytes(), r.Sig0, r.Sig1)
}

// ConformanceResponse is an operator's signed response to a conformance
// challenge. Its distinct domain tag (sohocloud/operator-conformance/v0)
// domain-separates it so a signature produced during conformance testing can
// NEVER be replayed as a live OperatorTransmission, and vice versa. It is
// signed with two keys like a transmission.
type ConformanceResponse struct {
	OperatorID string
	Challenge  []byte // the verifier-supplied challenge bytes
	Algo       string
	Idx0, Idx1 int
	Sig0, Sig1 []byte // over CanonicalBytes; excluded
}

// CanonicalBytes returns the deterministic signing payload with the two
// signatures excluded. Field order is fixed by SPEC §11.3: OperatorID,
// Challenge, Algo, Idx0, Idx1.
func (c ConformanceResponse) CanonicalBytes() []byte {
	b := canon.New(domainConformance)
	b.String(c.OperatorID)
	b.Bytes(c.Challenge)
	b.String(c.Algo)
	b.Uint64(uint64(c.Idx0))
	b.Uint64(uint64(c.Idx1))
	return b.Sum()
}

// Sign fills Sig0 and Sig1 over the challenge with the two given keys.
func (c *ConformanceResponse) Sign(priv0, priv1 ed25519.PrivateKey, idx0, idx1 int) {
	c.Idx0, c.Idx1 = idx0, idx1
	if c.Algo == "" {
		c.Algo = AlgoEd25519
	}
	msg := c.CanonicalBytes()
	c.Sig0 = ed25519.Sign(priv0, msg)
	c.Sig1 = ed25519.Sign(priv1, msg)
}

// Verify checks a conformance response against the registered keys. Same
// distinct-index + algorithm-binding + signature checks as a transmission.
// Because the domain tag differs, a valid response here is not a valid
// transmission.
func (c ConformanceResponse) Verify(keymap map[int]KeyRecord) error {
	if expectedSigLen(c.Algo) == 0 {
		return ErrUnknownAlgo
	}
	return verifyPair(keymap, c.Idx0, c.Idx1, c.Algo, c.CanonicalBytes(), c.Sig0, c.Sig1)
}
