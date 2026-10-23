package canon

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func FuzzVerifySig(f *testing.F) {
	// Seed with the normative signatures, not just random invalid byte strings.
	data, err := os.ReadFile("../testdata/vectors.json")
	if err != nil {
		f.Fatal(err)
	}
	var fixture struct {
		Messages []struct {
			PublicKey string `json:"public_key_hex"`
			Payload   string `json:"canonical_bytes_hex"`
			Signature string `json:"signature_hex"`
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		f.Fatal(err)
	}
	for _, m := range fixture.Messages {
		pub, err := hex.DecodeString(m.PublicKey)
		if err != nil {
			f.Fatal(err)
		}
		payload, err := hex.DecodeString(m.Payload)
		if err != nil {
			f.Fatal(err)
		}
		sig, err := hex.DecodeString(m.Signature)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(pub, payload, sig)
	}
	f.Add([]byte{}, []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, pub, payload, sig []byte) {
		got := VerifySig(pub, payload, sig)
		if len(pub) != 32 || len(sig) != 64 {
			if got {
				t.Fatal("accepted malformed key/signature")
			}
			return
		}
		if want := ed25519.Verify(pub, payload, sig); got != want {
			t.Fatalf("signature result %v, want %v", got, want)
		}
	})
}
