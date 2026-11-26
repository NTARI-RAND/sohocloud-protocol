//go:build !race

package canon

import (
	"bytes"
	"strings"
	"testing"
)

// Race instrumentation adds allocations, so measure only the normal build.
func TestLengthPrefixedFieldAllocatesOnce(t *testing.T) {
	text := strings.Repeat("x", 128)
	payload := []byte(text)
	want := append([]byte{0x80, 0x01}, payload...)
	cases := []struct {
		name  string
		write func(*Buffer)
	}{
		{"String", func(b *Buffer) { b.String(text) }},
		{"Bytes", func(b *Buffer) { b.Bytes(payload) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b Buffer
			allocations := testing.AllocsPerRun(100, func() {
				b = Buffer{}
				tc.write(&b)
			})
			if !bytes.Equal(b.Sum(), want) {
				t.Fatal("field bytes differ from the two-byte length prefix followed by the payload")
			}
			if allocations > 1 {
				t.Errorf("encoding one field allocated %.0f times, want at most 1", allocations)
			}
		})
	}
}
