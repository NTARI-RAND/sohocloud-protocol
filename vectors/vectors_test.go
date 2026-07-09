// Package vectors holds the deterministic, language-neutral conformance test
// vectors for sohocloud-protocol and the self-checking golden test that guards
// them.
//
// The committed testdata/vectors.json is ground truth: a foreign implementer
// bootstraps from the "primitives" section (low-level canon cases) and validates
// their message encoder/signer against the "messages" section. This test
// regenerates every vector from the CURRENT Go code and asserts byte-for-byte
// equality with the committed file, so any drift in canon or in a message's
// canonical field order fails the build and forces a conscious vectors update.
//
// Update the golden file after an INTENTIONAL change with either:
//
//	go test ./vectors -run TestVectors -update
//	SOHOCLOUD_UPDATE_VECTORS=1 go test ./vectors -run TestVectors
//
// Determinism: every key derives from a fixed 32-byte seed (ed25519 is
// deterministic per RFC 8032) and every timestamp is a hardcoded UnixNano
// constant. No crypto/rand, no time.Now — the same inputs always yield identical
// canonical_bytes and signatures.
package vectors

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/canon"
	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
	"github.com/NTARI-RAND/sohocloud-protocol/operator"
)

var updateFlag = flag.Bool("update", false, "rewrite testdata/vectors.json from current code")

const vectorsPath = "../testdata/vectors.json"

// --- On-disk vector file shape (language-neutral) ---

// vectorFile is the top-level document written to testdata/vectors.json.
type vectorFile struct {
	Note             string          `json:"note"`
	Encoding         encodingDoc     `json:"encoding"`
	Primitives       []primitiveCase `json:"primitives"`
	Messages         []messageCase   `json:"messages"`
	OperatorMessages []operatorCase  `json:"operator_messages"`
}

// encodingDoc restates the wire rules inline so the file is self-describing for a
// non-Go implementer. It is documentation, not something the test recomputes.
type encodingDoc struct {
	Varint string `json:"varint"`
	String string `json:"string"`
	Bytes  string `json:"bytes"`
	Bool   string `json:"bool"`
	Int64  string `json:"int64"`
	Uint64 string `json:"uint64"`
	Time   string `json:"time"`
	Domain string `json:"domain_tag"`
}

// primitiveCase is one low-level canon ground-truth case.
type primitiveCase struct {
	Kind     string `json:"kind"`
	Input    string `json:"input"`     // human-readable rendering of the input
	BytesHex string `json:"bytes_hex"` // the canon encoding of that input
}

// messageCase is one signed-message ground-truth case.
type messageCase struct {
	Name              string          `json:"name"`
	DomainTag         string          `json:"domain_tag"`
	Signer            string          `json:"signer"` // "node" | "coordinator"
	SeedHex           string          `json:"seed_hex"`
	PublicKeyHex      string          `json:"public_key_hex"`
	Fields            json.RawMessage `json:"fields"`
	CanonicalBytesHex string          `json:"canonical_bytes_hex"`
	SignatureHex      string          `json:"signature_hex"`
}

// operatorKey is one registered operator public key at an index (the network
// holds only these — never a private key).
type operatorKey struct {
	Index        int    `json:"index"`
	SeedHex      string `json:"seed_hex"`       // fixed seed the keypair derives from
	PublicKeyHex string `json:"public_key_hex"` // raw 32-byte ed25519 public key
	Algo         string `json:"algo"`
}

// operatorCase is one Layer-C operator message ground-truth case. Unlike a
// core messageCase it carries TWO signatures (2-of-7) over the SAME canonical
// bytes, records the two signing indices, and lists the full seven-key public
// registry a verifier resolves against. seed_hex is per index in `keys`; there
// is no server-side private key.
type operatorCase struct {
	Name              string          `json:"name"`
	DomainTag         string          `json:"domain_tag"`
	Signer            string          `json:"signer"` // always "operator"
	Keys              []operatorKey   `json:"keys"`
	Idx0              int             `json:"idx0"`
	Idx1              int             `json:"idx1"`
	Fields            json.RawMessage `json:"fields"`
	CanonicalBytesHex string          `json:"canonical_bytes_hex"`
	Sig0Hex           string          `json:"sig0_hex"`
	Sig1Hex           string          `json:"sig1_hex"`
}

// --- Fixed determinism inputs ---

// seedFill returns a fixed 32-byte seed whose bytes are all b. Distinct per
// message so each message has a distinct key, yet fully reproducible.
func seedFill(b byte) []byte {
	s := make([]byte, ed25519.SeedSize)
	for i := range s {
		s[i] = b
	}
	return s
}

// Fixed instants (UTC). Chosen to exercise epoch 0 and a specific nanosecond.
const (
	// 2023-11-14T22:13:20Z == unix 1_700_000_000; nanos chosen non-zero.
	fixedNanoA = int64(1_700_000_000_000_000_123)
	fixedNanoB = int64(1_700_000_050_500_000_000)
	fixedNanoC = int64(1_700_000_100_000_000_000)
	fixedNanoD = int64(1_699_999_990_000_000_777)
)

func tAt(nano int64) time.Time { return time.Unix(0, nano).UTC() }

// --- Primitive vector generation (calls the ACTUAL canon functions) ---

func hexOf(b []byte) string { return hex.EncodeToString(b) }

func genPrimitives() []primitiveCase {
	var out []primitiveCase

	countHex := func(n int) string { return hexOf((&canon.Buffer{}).Count(n).Sum()) }
	strHex := func(s string) string { return hexOf((&canon.Buffer{}).String(s).Sum()) }
	boolHex := func(v bool) string { return hexOf((&canon.Buffer{}).Bool(v).Sum()) }
	i64Hex := func(v int64) string { return hexOf((&canon.Buffer{}).Int64(v).Sum()) }
	u64Hex := func(v uint64) string { return hexOf((&canon.Buffer{}).Uint64(v).Sum()) }
	timeHex := func(t time.Time) string { return hexOf((&canon.Buffer{}).Time(t).Sum()) }

	// uvarint (canon.Count) — the LEB128 length/count prefix.
	for _, v := range []int{0, 1, 127, 128, 255, 256, 300, 16384, 1 << 30} {
		out = append(out, primitiveCase{
			Kind:     "uvarint",
			Input:    itoa(v),
			BytesHex: countHex(v),
		})
	}

	// int64 — fixed 8-byte big-endian two's complement.
	for _, v := range []int64{0, 1, -1, 2, -2, math.MaxInt64, math.MinInt64} {
		out = append(out, primitiveCase{
			Kind:     "int64",
			Input:    i64toa(v),
			BytesHex: i64Hex(v),
		})
	}

	// uint64 — fixed 8-byte big-endian.
	for _, v := range []uint64{0, 1, math.MaxUint64} {
		out = append(out, primitiveCase{
			Kind:     "uint64",
			Input:    u64toa(v),
			BytesHex: u64Hex(v),
		})
	}

	// bool — single byte 0x00 / 0x01.
	out = append(out,
		primitiveCase{Kind: "bool", Input: "false", BytesHex: boolHex(false)},
		primitiveCase{Kind: "bool", Input: "true", BytesHex: boolHex(true)},
	)

	// string — LEB128 length then raw UTF-8 (incl. a multibyte char).
	out = append(out,
		primitiveCase{Kind: "string", Input: "", BytesHex: strHex("")},
		primitiveCase{Kind: "string", Input: "café — 日本語 🚀", BytesHex: strHex("café — 日本語 🚀")},
	)

	// time — int64 UTC UnixNano. Epoch 0 and a fixed nanosecond instant.
	out = append(out,
		primitiveCase{Kind: "time", Input: "1970-01-01T00:00:00Z (unixnano 0)", BytesHex: timeHex(tAt(0))},
		primitiveCase{Kind: "time", Input: "unixnano " + i64toa(fixedNanoA), BytesHex: timeHex(tAt(fixedNanoA))},
	)

	return out
}

// --- Message vector generation (calls the ACTUAL CanonicalBytes/Sign) ---

// fieldsJSON marshals human-readable field values with stable key ordering
// (Go's encoding/json sorts map keys, and struct field order is fixed).
func fieldsJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal fields: %v", err)
	}
	return json.RawMessage(raw)
}

func genMessages(t *testing.T) []messageCase {
	t.Helper()
	var out []messageCase

	// 1. CapabilityListing — node-signed. Repeated Printers and GPUs with >=2
	// elements each so the count-prefixed repeat encoding is exercised.
	{
		seed := seedFill(0x11)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		l := listing.CapabilityListing{
			NodeID: identity.NodeID("node-alpha"),
			Class:  listing.ClassServer,
			Printers: []listing.PrinterCapability{
				{Kind: listing.PrinterTraditional, Model: "HP LaserJet"},
				{Kind: listing.Printer3D, Model: "Prusa MK4"},
			},
			GPUs: []listing.GPUCapability{
				{API: listing.GPUCUDA, Model: "RTX 4090", VRAMMB: 24576},
				{API: listing.GPUVulkan, Model: "Adreno 750", VRAMMB: 8192},
			},
			Capacity: listing.Capacity{VCPUs: 16, MemMB: 65536, DiskMB: 2_000_000, StorageCommitMB: 500_000, PrintQPS: 3},
			OptIn:    listing.WorkloadOptIn{Compute: true, Print: false, Storage: true},
			IssuedAt: tAt(fixedNanoA),
			Seq:      42,
		}
		l.Sign(priv)
		out = append(out, messageCase{
			Name:         "CapabilityListing",
			DomainTag:    "sohocloud/listing/v0",
			Signer:       "node",
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Fields: fieldsJSON(t, map[string]any{
				"NodeID": string(l.NodeID),
				"Class":  string(l.Class),
				"Printers": []map[string]string{
					{"Kind": string(l.Printers[0].Kind), "Model": l.Printers[0].Model},
					{"Kind": string(l.Printers[1].Kind), "Model": l.Printers[1].Model},
				},
				"GPUs": []map[string]any{
					{"API": string(l.GPUs[0].API), "Model": l.GPUs[0].Model, "VRAMMB": l.GPUs[0].VRAMMB},
					{"API": string(l.GPUs[1].API), "Model": l.GPUs[1].Model, "VRAMMB": l.GPUs[1].VRAMMB},
				},
				"Capacity": map[string]int{
					"VCPUs": l.Capacity.VCPUs, "MemMB": l.Capacity.MemMB,
					"DiskMB": l.Capacity.DiskMB, "StorageCommitMB": l.Capacity.StorageCommitMB,
					"PrintQPS": l.Capacity.PrintQPS,
				},
				"OptIn":    map[string]bool{"Compute": l.OptIn.Compute, "Print": l.OptIn.Print, "Storage": l.OptIn.Storage},
				"IssuedAt": "unixnano " + i64toa(fixedNanoA),
				"Seq":      l.Seq,
			}),
			CanonicalBytesHex: hexOf(l.CanonicalBytes()),
			SignatureHex:      hexOf(l.Signature),
		})
	}

	// 2. Heartbeat — node-signed.
	{
		seed := seedFill(0x22)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		h := liveness.Heartbeat{
			NodeID: identity.NodeID("node-alpha"),
			SentAt: tAt(fixedNanoB),
			Seq:    43,
		}
		h.Sign(priv)
		out = append(out, messageCase{
			Name:         "Heartbeat",
			DomainTag:    "sohocloud/heartbeat/v0",
			Signer:       "node",
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Fields: fieldsJSON(t, map[string]any{
				"NodeID": string(h.NodeID),
				"SentAt": "unixnano " + i64toa(fixedNanoB),
				"Seq":    h.Seq,
			}),
			CanonicalBytesHex: hexOf(h.CanonicalBytes()),
			SignatureHex:      hexOf(h.Signature),
		})
	}

	// 3. Assignment — coordinator-signed. Repeated Args with >=2 elements.
	{
		seed := seedFill(0x33)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		a := employment.Assignment{
			JobID:  "job-7f3a",
			NodeID: identity.NodeID("node-alpha"),
			Spec: employment.JobSpec{
				Workload:    "compute",
				Image:       "ghcr.io/ntari/render:1.2.3",
				Args:        []string{"--frames", "120", "--quality", "high"},
				PrinterKind: "",
			},
			Fee:       fees.Terms{ContributorShareBps: 6500, PlatformFeeBps: 3500},
			OfferedAt: tAt(fixedNanoC),
		}
		a.Sign(priv)
		out = append(out, messageCase{
			Name:         "Assignment",
			DomainTag:    "sohocloud/assignment/v0",
			Signer:       "coordinator",
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Fields: fieldsJSON(t, map[string]any{
				"JobID":  a.JobID,
				"NodeID": string(a.NodeID),
				"Spec": map[string]any{
					"Workload":    a.Spec.Workload,
					"Image":       a.Spec.Image,
					"Args":        a.Spec.Args,
					"PrinterKind": a.Spec.PrinterKind,
				},
				"Fee": map[string]int{
					"ContributorShareBps": a.Fee.ContributorShareBps,
					"PlatformFeeBps":      a.Fee.PlatformFeeBps,
				},
				"OfferedAt": "unixnano " + i64toa(fixedNanoC),
			}),
			CanonicalBytesHex: hexOf(a.CanonicalBytes()),
			SignatureHex:      hexOf(a.Signature),
		})
	}

	// 4. Decline — node-signed. Empty-args-equivalent: minimal fields.
	{
		seed := seedFill(0x44)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		d := employment.Decline{
			JobID:      "job-7f3a",
			NodeID:     identity.NodeID("node-alpha"),
			Reason:     employment.DeclineLocalPolicy,
			DeclinedAt: tAt(fixedNanoD),
		}
		d.Sign(priv)
		out = append(out, messageCase{
			Name:         "Decline",
			DomainTag:    "sohocloud/decline/v0",
			Signer:       "node",
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Fields: fieldsJSON(t, map[string]any{
				"JobID":      d.JobID,
				"NodeID":     string(d.NodeID),
				"Reason":     string(d.Reason),
				"DeclinedAt": "unixnano " + i64toa(fixedNanoD),
			}),
			CanonicalBytesHex: hexOf(d.CanonicalBytes()),
			SignatureHex:      hexOf(d.Signature),
		})
	}

	// 5. JobReport — node-signed. Exercises a negative-ish exit path + bool +
	//    a non-empty FailureCause and epoch-0 boundary is covered in primitives.
	{
		seed := seedFill(0x55)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		r := employment.JobReport{
			JobID:          "job-7f3a",
			NodeID:         identity.NodeID("node-alpha"),
			ExitCode:       137,
			FailureCause:   "oom_killed",
			TmpfsExhausted: true,
			StartedAt:      tAt(fixedNanoA),
			FinishedAt:     tAt(fixedNanoC),
		}
		r.Sign(priv)
		out = append(out, messageCase{
			Name:         "JobReport",
			DomainTag:    "sohocloud/jobreport/v0",
			Signer:       "node",
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Fields: fieldsJSON(t, map[string]any{
				"JobID":          r.JobID,
				"NodeID":         string(r.NodeID),
				"ExitCode":       r.ExitCode,
				"FailureCause":   r.FailureCause,
				"TmpfsExhausted": r.TmpfsExhausted,
				"StartedAt":      "unixnano " + i64toa(fixedNanoA),
				"FinishedAt":     "unixnano " + i64toa(fixedNanoC),
			}),
			CanonicalBytesHex: hexOf(r.CanonicalBytes()),
			SignatureHex:      hexOf(r.Signature),
		})
	}

	// 6. FeeDeclaration — coordinator-signed.
	{
		seed := seedFill(0x66)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		f := fees.FeeDeclaration{
			CoordinatorID: "coord-soholink",
			Terms:         fees.Terms{ContributorShareBps: 7000, PlatformFeeBps: 3000},
			EffectiveAt:   tAt(fixedNanoB),
			Seq:           5,
		}
		f.Sign(priv)
		out = append(out, messageCase{
			Name:         "FeeDeclaration",
			DomainTag:    "sohocloud/fee/v0",
			Signer:       "coordinator",
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Fields: fieldsJSON(t, map[string]any{
				"CoordinatorID":       f.CoordinatorID,
				"ContributorShareBps": f.Terms.ContributorShareBps,
				"PlatformFeeBps":      f.Terms.PlatformFeeBps,
				"EffectiveAt":         "unixnano " + i64toa(fixedNanoB),
				"Seq":                 f.Seq,
			}),
			CanonicalBytesHex: hexOf(f.CanonicalBytes()),
			SignatureHex:      hexOf(f.Signature),
		})
	}

	return out
}

// --- Operator (Layer C) vector generation ---

// operatorSeed returns the fixed 32-byte seed for operator key index i:
// 0xA0+i in every byte. Deterministic and distinct per index.
func operatorSeed(i int) []byte {
	s := make([]byte, ed25519.SeedSize)
	for j := range s {
		s[j] = byte(0xA0 + i)
	}
	return s
}

// operatorRegistry derives the seven fixed operator keypairs and returns the
// private keys plus the JSON-facing public registry (only public keys are ever
// registered with the network).
func operatorRegistry() ([]ed25519.PrivateKey, map[int]operator.KeyRecord, []operatorKey) {
	privs := make([]ed25519.PrivateKey, operator.KeyIndexCount)
	km := make(map[int]operator.KeyRecord, operator.KeyIndexCount)
	keysJSON := make([]operatorKey, operator.KeyIndexCount)
	for i := 0; i < operator.KeyIndexCount; i++ {
		seed := operatorSeed(i)
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		privs[i] = priv
		km[i] = operator.KeyRecord{PublicKey: pub, Algo: operator.AlgoEd25519}
		keysJSON[i] = operatorKey{
			Index:        i,
			SeedHex:      hexOf(seed),
			PublicKeyHex: hexOf(pub),
			Algo:         operator.AlgoEd25519,
		}
	}
	return privs, km, keysJSON
}

func genOperatorMessages(t *testing.T) []operatorCase {
	t.Helper()
	privs, km, keysJSON := operatorRegistry()
	var out []operatorCase

	// Fixed nonces/challenge — no crypto/rand.
	nonceA := bytes.Repeat([]byte{0x5a}, operator.MinNonceLen)
	nonceB := bytes.Repeat([]byte{0x33}, operator.MinNonceLen)
	challenge := bytes.Repeat([]byte{0x11}, 32)

	// 1. OperatorTransmission — signed with indices 2 and 5.
	{
		const idx0, idx1 = 2, 5
		tx := operator.OperatorTransmission{
			OperatorID: "cloudy",
			TsUnixNano: fixedNanoA,
			Nonce:      nonceA,
			Seq:        7,
			Algo:       operator.AlgoEd25519,
		}
		tx.Sign(privs[idx0], privs[idx1], idx0, idx1)
		if err := tx.Verify(km); err != nil {
			t.Fatalf("generated OperatorTransmission does not verify: %v", err)
		}
		out = append(out, operatorCase{
			Name:      "OperatorTransmission",
			DomainTag: "sohocloud/operator/v0",
			Signer:    "operator",
			Keys:      keysJSON,
			Idx0:      idx0,
			Idx1:      idx1,
			Fields: fieldsJSON(t, map[string]any{
				"OperatorID": tx.OperatorID,
				"Ts":         "unixnano " + i64toa(fixedNanoA),
				"Nonce":      hexOf(tx.Nonce),
				"Seq":        tx.Seq,
				"Algo":       tx.Algo,
				"Idx0":       tx.Idx0,
				"Idx1":       tx.Idx1,
			}),
			CanonicalBytesHex: hexOf(tx.CanonicalBytes()),
			Sig0Hex:           hexOf(tx.Sig0),
			Sig1Hex:           hexOf(tx.Sig1),
		})
	}

	// 2. OperatorRotation — new key at index 4, authorized by indices 0 and 1.
	{
		const idx0, idx1 = 0, 1
		newPub := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x77}, ed25519.SeedSize)).Public().(ed25519.PublicKey)
		r := operator.OperatorRotation{
			OperatorID:   "cloudy",
			KeyIndex:     4,
			NewPublicKey: newPub,
			Algo:         operator.AlgoEd25519,
			TsUnixNano:   fixedNanoB,
			Nonce:        nonceB,
			Seq:          9,
		}
		r.Sign(privs[idx0], privs[idx1], idx0, idx1)
		if err := r.Verify(km); err != nil {
			t.Fatalf("generated OperatorRotation does not verify: %v", err)
		}
		out = append(out, operatorCase{
			Name:      "OperatorRotation",
			DomainTag: "sohocloud/operator-rotate/v0",
			Signer:    "operator",
			Keys:      keysJSON,
			Idx0:      idx0,
			Idx1:      idx1,
			Fields: fieldsJSON(t, map[string]any{
				"OperatorID":   r.OperatorID,
				"KeyIndex":     r.KeyIndex,
				"NewPublicKey": hexOf(r.NewPublicKey),
				"Algo":         r.Algo,
				"Ts":           "unixnano " + i64toa(fixedNanoB),
				"Nonce":        hexOf(r.Nonce),
				"Seq":          r.Seq,
				"Idx0":         r.Idx0,
				"Idx1":         r.Idx1,
			}),
			CanonicalBytesHex: hexOf(r.CanonicalBytes()),
			Sig0Hex:           hexOf(r.Sig0),
			Sig1Hex:           hexOf(r.Sig1),
		})
	}

	// 3. ConformanceResponse — challenge signed by indices 2 and 3.
	{
		const idx0, idx1 = 2, 3
		c := operator.ConformanceResponse{
			OperatorID: "cloudy",
			Challenge:  challenge,
			Algo:       operator.AlgoEd25519,
		}
		c.Sign(privs[idx0], privs[idx1], idx0, idx1)
		if err := c.Verify(km); err != nil {
			t.Fatalf("generated ConformanceResponse does not verify: %v", err)
		}
		out = append(out, operatorCase{
			Name:      "ConformanceResponse",
			DomainTag: "sohocloud/operator-conformance/v0",
			Signer:    "operator",
			Keys:      keysJSON,
			Idx0:      idx0,
			Idx1:      idx1,
			Fields: fieldsJSON(t, map[string]any{
				"OperatorID": c.OperatorID,
				"Challenge":  hexOf(c.Challenge),
				"Algo":       c.Algo,
				"Idx0":       c.Idx0,
				"Idx1":       c.Idx1,
			}),
			CanonicalBytesHex: hexOf(c.CanonicalBytes()),
			Sig0Hex:           hexOf(c.Sig0),
			Sig1Hex:           hexOf(c.Sig1),
		})
	}

	return out
}

func buildVectorFile(t *testing.T) vectorFile {
	t.Helper()
	return vectorFile{
		Note: "sohocloud-protocol v0 (UNSTABLE) conformance vectors. Ground truth is " +
			"produced by the reference Go canon and message encoders. A foreign " +
			"implementer bootstraps from `primitives`, validates the six core " +
			"messages against `messages`, and validates the Layer-C operator " +
			"messages (2-of-7, two signatures each) against `operator_messages`. " +
			"All hex is lowercase. Regenerate via `go test ./vectors -update` after " +
			"an intentional change.",
		Encoding: encodingDoc{
			Varint: "unsigned LEB128 (encoding/binary.PutUvarint); used for lengths and repeated-field counts",
			String: "uvarint(byteLen) then raw UTF-8 bytes",
			Bytes:  "uvarint(byteLen) then raw bytes; the Signature field is never encoded",
			Bool:   "single byte 0x00 (false) or 0x01 (true)",
			Int64:  "fixed 8 bytes big-endian, two's-complement",
			Uint64: "fixed 8 bytes big-endian",
			Time:   "int64 UTC Unix nanoseconds (see int64 rule); location/monotonic dropped",
			Domain: "each message's canonical bytes begin with its domain-tag encoded as a string",
		},
		Primitives:       genPrimitives(),
		Messages:         genMessages(t),
		OperatorMessages: genOperatorMessages(t),
	}
}

// marshalVectors renders the vector file as stable, indented JSON with a
// trailing newline (gofmt-independent; json.MarshalIndent is deterministic).
func marshalVectors(t *testing.T, vf vectorFile) []byte {
	t.Helper()
	buf, err := json.MarshalIndent(vf, "", "  ")
	if err != nil {
		t.Fatalf("marshal vectors: %v", err)
	}
	return append(buf, '\n')
}

// TestVectors regenerates all vectors from the current code and, unless -update
// is set, asserts they equal the committed testdata/vectors.json byte-for-byte.
func TestVectors(t *testing.T) {
	vf := buildVectorFile(t)
	got := marshalVectors(t, vf)

	if *updateFlag || os.Getenv("SOHOCLOUD_UPDATE_VECTORS") != "" {
		if err := os.MkdirAll(filepath.Dir(vectorsPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(vectorsPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s (%d bytes)", vectorsPath, len(got))
		return
	}

	want, err := os.ReadFile(vectorsPath)
	if err != nil {
		t.Fatalf("read golden vectors (run `go test ./vectors -update` to create): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("vectors.json is stale: regenerated bytes differ from committed file.\n"+
			"If this change is intentional, run `go test ./vectors -update`.\n"+
			"got %d bytes, want %d bytes", len(got), len(want))
	}
}

// TestVectorSignaturesVerify re-parses the committed file and checks each
// message signature with the real Verify() against the derived public key, and
// confirms each canonical_bytes begins with the domain tag's canonical string
// encoding. This validates the file a foreign implementer actually consumes.
func TestVectorSignaturesVerify(t *testing.T) {
	raw, err := os.ReadFile(vectorsPath)
	if err != nil {
		t.Fatalf("read golden vectors (run `go test ./vectors -update`): %v", err)
	}
	var vf vectorFile
	if err := json.Unmarshal(raw, &vf); err != nil {
		t.Fatalf("unmarshal vectors: %v", err)
	}
	if len(vf.Messages) != 6 {
		t.Fatalf("expected 6 message vectors, got %d", len(vf.Messages))
	}

	for _, m := range vf.Messages {
		m := m
		t.Run(m.Name, func(t *testing.T) {
			seed, err := hex.DecodeString(m.SeedHex)
			if err != nil {
				t.Fatalf("bad seed hex: %v", err)
			}
			if len(seed) != ed25519.SeedSize {
				t.Fatalf("seed len = %d, want %d", len(seed), ed25519.SeedSize)
			}
			priv := ed25519.NewKeyFromSeed(seed)
			pub := priv.Public().(ed25519.PublicKey)

			// Derived public key must match the recorded one.
			if got := hexOf(pub); got != m.PublicKeyHex {
				t.Fatalf("public key mismatch: derived %s, recorded %s", got, m.PublicKeyHex)
			}

			cb, err := hex.DecodeString(m.CanonicalBytesHex)
			if err != nil {
				t.Fatalf("bad canonical_bytes hex: %v", err)
			}
			sig, err := hex.DecodeString(m.SignatureHex)
			if err != nil {
				t.Fatalf("bad signature hex: %v", err)
			}

			// Real signature verification against the derived key.
			if len(sig) != ed25519.SignatureSize {
				t.Fatalf("signature len = %d, want %d", len(sig), ed25519.SignatureSize)
			}
			if !ed25519.Verify(pub, cb, sig) {
				t.Fatalf("signature does not verify over canonical_bytes")
			}

			// canonical_bytes MUST begin with the domain tag's string encoding.
			wantPrefix := (&canon.Buffer{}).String(m.DomainTag).Sum()
			if !bytes.HasPrefix(cb, wantPrefix) {
				t.Fatalf("canonical_bytes does not begin with domain tag %q encoding\nprefix want %s\ngot        %s",
					m.DomainTag, hexOf(wantPrefix), hexOf(cb[:min(len(cb), len(wantPrefix))]))
			}
		})
	}
}

// TestOperatorVectorSignaturesVerify re-parses the committed operator_messages
// section and, for each case, rebuilds the seven-key public registry from the
// recorded seeds, then checks BOTH signatures with the real operator Verify
// against the derived keys and confirms each canonical_bytes begins with the
// message's domain-tag encoding. This validates the 2-of-7 file a foreign
// implementer actually consumes, using only public keys.
func TestOperatorVectorSignaturesVerify(t *testing.T) {
	raw, err := os.ReadFile(vectorsPath)
	if err != nil {
		t.Fatalf("read golden vectors (run `go test ./vectors -update`): %v", err)
	}
	var vf vectorFile
	if err := json.Unmarshal(raw, &vf); err != nil {
		t.Fatalf("unmarshal vectors: %v", err)
	}
	if len(vf.OperatorMessages) != 3 {
		t.Fatalf("expected 3 operator message vectors, got %d", len(vf.OperatorMessages))
	}

	for _, m := range vf.OperatorMessages {
		m := m
		t.Run(m.Name, func(t *testing.T) {
			if m.Signer != "operator" {
				t.Fatalf("operator vector signer = %q, want operator", m.Signer)
			}
			if len(m.Keys) != operator.KeyIndexCount {
				t.Fatalf("operator key registry has %d keys, want %d", len(m.Keys), operator.KeyIndexCount)
			}

			// Rebuild the public registry from the recorded seeds and confirm each
			// derived public key matches the recorded one (public keys only).
			km := make(map[int]operator.KeyRecord, len(m.Keys))
			for _, k := range m.Keys {
				seed, err := hex.DecodeString(k.SeedHex)
				if err != nil || len(seed) != ed25519.SeedSize {
					t.Fatalf("key %d bad seed hex", k.Index)
				}
				pub := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
				if hexOf(pub) != k.PublicKeyHex {
					t.Fatalf("key %d public mismatch: derived %s, recorded %s", k.Index, hexOf(pub), k.PublicKeyHex)
				}
				km[k.Index] = operator.KeyRecord{PublicKey: pub, Algo: k.Algo}
			}

			cb, err := hex.DecodeString(m.CanonicalBytesHex)
			if err != nil {
				t.Fatalf("bad canonical_bytes hex: %v", err)
			}
			sig0, err := hex.DecodeString(m.Sig0Hex)
			if err != nil {
				t.Fatalf("bad sig0 hex: %v", err)
			}
			sig1, err := hex.DecodeString(m.Sig1Hex)
			if err != nil {
				t.Fatalf("bad sig1 hex: %v", err)
			}

			// canonical_bytes MUST begin with the domain tag's string encoding.
			wantPrefix := (&canon.Buffer{}).String(m.DomainTag).Sum()
			if !bytes.HasPrefix(cb, wantPrefix) {
				t.Fatalf("canonical_bytes does not begin with domain tag %q encoding", m.DomainTag)
			}

			// Both signatures must verify at their indices via the real Verify.
			// Reconstruct the concrete message type from canonical bytes so we
			// use the actual Verify path (not just raw ed25519), exercising the
			// distinct-index + algorithm-binding checks.
			verifyOne := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("operator Verify rejected a committed vector: %v", err)
				}
			}
			switch m.Name {
			case "OperatorTransmission":
				var f struct {
					OperatorID string `json:"OperatorID"`
					Ts         string `json:"Ts"`
					Nonce      string `json:"Nonce"`
					Seq        uint64 `json:"Seq"`
					Algo       string `json:"Algo"`
					Idx0, Idx1 int
				}
				mustUnmarshal(t, m.Fields, &f)
				nonce, _ := hex.DecodeString(f.Nonce)
				tx := operator.OperatorTransmission{
					OperatorID: f.OperatorID, TsUnixNano: nanosOf(f.Ts), Nonce: nonce,
					Seq: f.Seq, Algo: f.Algo, Idx0: m.Idx0, Idx1: m.Idx1, Sig0: sig0, Sig1: sig1,
				}
				if !bytes.Equal(tx.CanonicalBytes(), cb) {
					t.Fatal("rebuilt OperatorTransmission canonical bytes differ from committed")
				}
				verifyOne(tx.Verify(km))
			case "OperatorRotation":
				var f struct {
					OperatorID   string `json:"OperatorID"`
					KeyIndex     int    `json:"KeyIndex"`
					NewPublicKey string `json:"NewPublicKey"`
					Algo         string `json:"Algo"`
					Ts           string `json:"Ts"`
					Nonce        string `json:"Nonce"`
					Seq          uint64 `json:"Seq"`
				}
				mustUnmarshal(t, m.Fields, &f)
				nonce, _ := hex.DecodeString(f.Nonce)
				npub, _ := hex.DecodeString(f.NewPublicKey)
				r := operator.OperatorRotation{
					OperatorID: f.OperatorID, KeyIndex: f.KeyIndex, NewPublicKey: npub, Algo: f.Algo,
					TsUnixNano: nanosOf(f.Ts), Nonce: nonce, Seq: f.Seq, Idx0: m.Idx0, Idx1: m.Idx1,
					Sig0: sig0, Sig1: sig1,
				}
				if !bytes.Equal(r.CanonicalBytes(), cb) {
					t.Fatal("rebuilt OperatorRotation canonical bytes differ from committed")
				}
				verifyOne(r.Verify(km))
			case "ConformanceResponse":
				var f struct {
					OperatorID string `json:"OperatorID"`
					Challenge  string `json:"Challenge"`
					Algo       string `json:"Algo"`
				}
				mustUnmarshal(t, m.Fields, &f)
				ch, _ := hex.DecodeString(f.Challenge)
				c := operator.ConformanceResponse{
					OperatorID: f.OperatorID, Challenge: ch, Algo: f.Algo,
					Idx0: m.Idx0, Idx1: m.Idx1, Sig0: sig0, Sig1: sig1,
				}
				if !bytes.Equal(c.CanonicalBytes(), cb) {
					t.Fatal("rebuilt ConformanceResponse canonical bytes differ from committed")
				}
				verifyOne(c.Verify(km))
			default:
				t.Fatalf("unknown operator vector %q", m.Name)
			}
		})
	}
}

func mustUnmarshal(t *testing.T, raw json.RawMessage, v any) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("unmarshal operator fields: %v", err)
	}
}

// nanosOf extracts the integer nanoseconds from an "unixnano <int>" input.
func nanosOf(s string) int64 {
	n, err := strconv.ParseInt(s[len("unixnano "):], 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// --- stdlib integer formatters (human-readable inputs in the vector file) ---

func itoa(v int) string      { return strconv.Itoa(v) }
func i64toa(v int64) string  { return strconv.FormatInt(v, 10) }
func u64toa(v uint64) string { return strconv.FormatUint(v, 10) }
