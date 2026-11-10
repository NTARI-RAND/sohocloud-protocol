package sohocloud_test

import (
	"bytes"
	"crypto/ed25519"
	"reflect"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/lease"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
)

type nodeMessage interface {
	CanonicalBytes() []byte
	Sign(ed25519.PrivateKey)
	Verify(ed25519.PublicKey) bool
}

func nodeMessages(pub ed25519.PublicKey) []nodeMessage {
	return []nodeMessage{
		&listing.CapabilityListing{NodeID: "node", Seq: 1},
		&liveness.Heartbeat{NodeID: "node", Seq: 1},
		&employment.Assignment{NodeID: "node", JobID: "job"},
		&employment.Decline{NodeID: "node", JobID: "job"},
		&employment.JobReport{NodeID: "node", JobID: "job"},
		&fees.FeeDeclaration{CoordinatorID: "coord", Seq: 1},
		&lease.StorageLease{NodeID: "node", LeaseID: "lease", Seq: 1},
		&lease.LeaseDecline{NodeID: "node", LeaseID: "lease"},
		&lease.LeaseRelease{NodeID: "node", LeaseID: "lease"},
		&lease.ProofResponse{NodeID: "node", LeaseID: "lease"},
		&identity.KeyRotation{NodeID: "node", Algo: "ed25519", NewPublicKey: pub, Seq: 1},
		&identity.KeyRevocation{NodeID: "node", RevokedPublicKey: pub, Seq: 1},
	}
}

func setMessageTimes(m nodeMessage, stamp time.Time) {
	value := reflect.ValueOf(m).Elem()
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Type() == reflect.TypeOf(time.Time{}) {
			value.Field(i).Set(reflect.ValueOf(stamp))
		}
	}
}

// An attacker can sign the encoder's incomplete output. SPEC §3 still
// requires rejection: valid cryptography cannot make an invalid time valid.
func TestVerifyRejectsUnrepresentableTime(t *testing.T) {
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, 32))
	pub := priv.Public().(ed25519.PublicKey)
	good := time.Unix(1700000000, 0).UTC()
	for _, m := range nodeMessages(pub) {
		value := reflect.ValueOf(m).Elem()
		for field := 0; field < value.NumField(); field++ {
			if value.Field(field).Type() != reflect.TypeOf(time.Time{}) {
				continue
			}
			for _, bad := range []time.Time{time.Time{}, time.Unix(0, -1<<63).Add(-time.Nanosecond), time.Unix(0, 1<<63-1).Add(time.Nanosecond), time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC)} {
				t.Run(value.Type().Name()+"/"+value.Type().Field(field).Name+"/"+bad.Format(time.RFC3339Nano), func(t *testing.T) {
					setMessageTimes(m, good)
					value.Field(field).Set(reflect.ValueOf(bad))
					value.FieldByName("Signature").SetBytes(ed25519.Sign(priv, m.CanonicalBytes()))
					if m.Verify(pub) {
						t.Fatal("accepted signed message with unrepresentable timestamp")
					}
				})
			}
		}
	}
}

// Also guard both ends of the valid interval: the fix must reject out-of-range
// values without rejecting legitimate int64 nanoseconds or nanosecond precision.
func TestVerifyRepresentableTimeBoundaries(t *testing.T) {
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, 32))
	pub := priv.Public().(ed25519.PublicKey)
	for _, m := range nodeMessages(pub) {
		for _, stamp := range []time.Time{time.Unix(0, -1<<63), time.Unix(0, 1<<63-1), time.Unix(1700000000, 123456789)} {
			setMessageTimes(m, stamp)
			m.Sign(priv)
			if !m.Verify(pub) {
				t.Fatalf("%T rejected valid time %v", m, stamp)
			}
		}
	}
}

func FuzzMessageVerification(f *testing.F) {
	f.Add(uint8(0), int64(1700000000), int32(123456789), []byte{}, []byte{})
	f.Add(uint8(11), int64(16725225600), int32(0), []byte{0}, []byte{0})      // year 2500
	f.Add(uint8(6), int64(-9223372037), int32(145224192), []byte{}, []byte{}) // minimum
	f.Fuzz(func(t *testing.T, kind uint8, seconds int64, nanos int32, key, sig []byte) {
		priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, 32))
		pub := priv.Public().(ed25519.PublicKey)
		messages := nodeMessages(pub)
		m := messages[int(kind)%len(messages)]
		stamp := time.Unix(seconds, int64(nanos)).UTC()
		setMessageTimes(m, stamp)
		// Derive validity independently of canon.ValidTime, including normalized
		// fractional seconds at both interval endpoints.
		sec, ns := stamp.Unix(), stamp.Nanosecond()
		valid := (sec > -9223372037 || sec == -9223372037 && ns >= 145224192) && (sec < 9223372036 || sec == 9223372036 && ns <= 854775807)
		signature := ed25519.Sign(priv, m.CanonicalBytes())
		value := reflect.ValueOf(m).Elem()
		value.FieldByName("Signature").SetBytes(signature)
		if got := m.Verify(pub); got != valid {
			t.Fatalf("%T time %v: verified %v, want %v", m, stamp, got, valid)
		}
		// Arbitrary untrusted key/signature lengths must never panic. Matching keys
		// may pass, so do not equate random input with invalid cryptography.
		value.FieldByName("Signature").SetBytes(sig)
		_ = m.Verify(ed25519.PublicKey(key))
		// The last byte is a signed field in every family (never the domain prefix).
		payload := m.CanonicalBytes()
		payload[len(payload)-1] ^= 1
		value.FieldByName("Signature").SetBytes(ed25519.Sign(priv, payload))
		if m.Verify(pub) {
			t.Fatalf("%T accepted signature over altered canonical bytes", m)
		}
	})
}
