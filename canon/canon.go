// Package canon defines the deterministic byte encoding over which every
// sohocloud-protocol signature is computed. A signature is always taken over
// the canonical bytes of a message WITH ITS SIGNATURE FIELD EXCLUDED.
//
// The encoding is length-prefixed and domain-tagged so that no two distinct
// field sequences can collide, and so that a signature over one message type
// can never be replayed as a valid signature over another. It is
// standard-library only and byte-stable across platforms and Go versions.
// SPEC.md documents the wire format precisely so a non-Go implementation can
// reproduce identical bytes and thus verify signatures produced here.
//
// Deliberately, this encoding is NOT encoding/json. JSON marshaling is used as
// a transport in transport/httpjson, but signing must not depend on JSON field
// ordering, whitespace, or number formatting, all of which can drift and
// silently invalidate signatures across versions or languages.
//
// v0: UNSTABLE. The format may change without compatibility guarantees until
// v1.
package canon

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"time"
)

// ErrTimeOutOfRange is latched on a Buffer when Time is given an instant
// outside the representable int64-UTC-nanoseconds range (SPEC §3). It is an
// encoding error, never a wrap or clamp.
var ErrTimeOutOfRange = errors.New("canon: time out of representable int64-nanosecond range")

// minRepresentableUnixNano / maxRepresentableUnixNano bound the instants that
// encode without int64 nanosecond overflow (approximately 1678..2262).
var (
	minRepresentableUnixNano = time.Unix(0, minInt64)
	maxRepresentableUnixNano = time.Unix(0, maxInt64)
)

const (
	minInt64 = -1 << 63
	maxInt64 = 1<<63 - 1
)

// ValidTime reports whether t encodes as int64 UTC Unix nanoseconds without
// overflow. Producers MUST validate timestamps with this before signing, and
// verifiers/decoders MUST reject a message whose timestamp fails it — an
// out-of-range instant is an encoding error, not a value to be wrapped
// (SPEC §3).
func ValidTime(t time.Time) bool {
	u := t.UTC()
	return !u.Before(minRepresentableUnixNano) && !u.After(maxRepresentableUnixNano)
}

// Buffer accumulates canonical bytes.
type Buffer struct {
	b   []byte
	err error
}

// New returns a Buffer whose first entry is a domain tag identifying the
// message type. Every message's canonical encoding MUST begin with a distinct
// tag so signatures are not transferable between message types.
func New(domainTag string) *Buffer {
	buf := &Buffer{}
	return buf.String(domainTag)
}

// Count appends an unsigned-varint length/count prefix for a following
// sequence of that many elements.
func (buf *Buffer) Count(n int) *Buffer {
	var tmp [binary.MaxVarintLen64]byte
	m := binary.PutUvarint(tmp[:], uint64(n))
	buf.b = append(buf.b, tmp[:m]...)
	return buf
}

// String appends a length-prefixed UTF-8 string.
func (buf *Buffer) String(s string) *Buffer {
	buf.Count(len(s))
	buf.b = append(buf.b, s...)
	return buf
}

// Bytes appends a length-prefixed byte slice.
func (buf *Buffer) Bytes(p []byte) *Buffer {
	buf.Count(len(p))
	buf.b = append(buf.b, p...)
	return buf
}

// Bool appends a single byte, 0 or 1.
func (buf *Buffer) Bool(v bool) *Buffer {
	if v {
		buf.b = append(buf.b, 1)
	} else {
		buf.b = append(buf.b, 0)
	}
	return buf
}

// Int64 appends a fixed 8-byte big-endian two's-complement integer.
func (buf *Buffer) Int64(v int64) *Buffer {
	return buf.Uint64(uint64(v))
}

// Uint64 appends a fixed 8-byte big-endian integer.
func (buf *Buffer) Uint64(v uint64) *Buffer {
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], v)
	buf.b = append(buf.b, tmp[:]...)
	return buf
}

// Time appends the instant as int64 UTC Unix nanoseconds. Location and
// monotonic-clock components are dropped so the encoding is stable across a
// JSON transport round-trip and across time zones.
func (buf *Buffer) Time(t time.Time) *Buffer {
	if !ValidTime(t) {
		// Latch an encoding error and append nothing rather than wrapping —
		// a wrapped value would silently alias two instants ~584 years apart
		// to identical bytes (SPEC §3). Callers surface this via Err().
		if buf.err == nil {
			buf.err = ErrTimeOutOfRange
		}
		return buf
	}
	return buf.Int64(t.UTC().UnixNano())
}

// Err returns the first encoding error latched during construction (currently
// only ErrTimeOutOfRange), or nil. A caller that builds canonical bytes from
// untrusted field values (a verifier or transport decoder) MUST check Err()
// and reject the message before trusting Sum().
func (buf *Buffer) Err() error {
	return buf.err
}

// Sum returns a copy of the accumulated canonical bytes. The caller signs or
// verifies over exactly these bytes. When Err() is non-nil the returned bytes
// are incomplete by construction (the offending field was skipped, not
// wrapped), so any signature over them will fail to verify — callers should
// check Err() rather than relying on that.
func (buf *Buffer) Sum() []byte {
	out := make([]byte, len(buf.b))
	copy(out, buf.b)
	return out
}

// VerifySig reports whether sig is a valid ed25519 signature by pub over msg.
// Unlike calling ed25519.Verify directly it is panic-safe: it rejects (returns
// false) when pub is not ed25519.PublicKeySize or sig is not
// ed25519.SignatureSize, instead of panicking. Every message Verify method
// routes through here so a malformed key on the wire can never crash a
// verifier.
func VerifySig(pub ed25519.PublicKey, msg, sig []byte) bool {
	return len(pub) == ed25519.PublicKeySize &&
		len(sig) == ed25519.SignatureSize &&
		ed25519.Verify(pub, msg, sig)
}
