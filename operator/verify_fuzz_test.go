package operator

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"testing"
)

func FuzzOperatorVerification(f *testing.F) {
	privs, keys := sevenKeys()
	tx := sampleTransmission()
	tx.Sign(privs[2], privs[5], 2, 5)
	next := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, 32)).Public().(ed25519.PublicKey)
	rotation := OperatorRotation{OperatorID: "cloudy", KeyIndex: 3, NewPublicKey: next, Algo: AlgoEd25519, Nonce: bytes.Repeat([]byte{1}, 16), Seq: 1}
	rotation.Sign(privs[2], privs[5], 2, 5)
	response := ConformanceResponse{OperatorID: "cloudy", Challenge: []byte("challenge"), Algo: AlgoEd25519}
	response.Sign(privs[2], privs[5], 2, 5)
	for i, m := range []any{tx, rotation, response} {
		b, err := json.Marshal(m)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(uint8(i), string(b), keys[2].PublicKey)
	}
	f.Add(uint8(0), `{"Idx0":-1,"Idx1":7,"Algo":"unknown"}`, []byte{})
	f.Fuzz(func(t *testing.T, kind uint8, payload string, key []byte) {
		// A registry supplied by an external coordinator may contain malformed
		// key material; verification must fail safely rather than panic.
		_, km := sevenKeys()
		km[2] = KeyRecord{PublicKey: key, Algo: AlgoEd25519}
		switch kind % 3 {
		case 0:
			var m OperatorTransmission
			if json.Unmarshal([]byte(payload), &m) != nil {
				return
			}
			if m.Verify(km) == nil {
				m.OperatorID += "changed"
				if m.Verify(km) == nil {
					t.Fatal("transmission identity not signed")
				}
			}
		case 1:
			var m OperatorRotation
			if json.Unmarshal([]byte(payload), &m) != nil {
				return
			}
			if m.Verify(km) == nil {
				m.OperatorID += "changed"
				if m.Verify(km) == nil {
					t.Fatal("rotation identity not signed")
				}
			}
		case 2:
			var m ConformanceResponse
			if json.Unmarshal([]byte(payload), &m) != nil {
				return
			}
			if m.Verify(km) == nil {
				m.Challenge = append(m.Challenge, 1)
				if m.Verify(km) == nil {
					t.Fatal("challenge not signed")
				}
			}
		}
	})
}
