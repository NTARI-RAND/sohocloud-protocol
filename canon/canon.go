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
	"encoding/binary"
	"time"
)

// Buffer accumulates canonical bytes.
type Buffer struct {
	b []byte
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
	return buf.Int64(t.UTC().UnixNano())
}

// Sum returns a copy of the accumulated canonical bytes. The caller signs or
// verifies over exactly these bytes.
func (buf *Buffer) Sum() []byte {
	out := make([]byte, len(buf.b))
	copy(out, buf.b)
	return out
}
