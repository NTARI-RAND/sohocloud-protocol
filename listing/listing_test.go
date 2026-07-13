package listing

import (
	"crypto/ed25519"
	"testing"
	"time"
)

func TestListingRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	l := CapabilityListing{
		NodeID:   "node-1",
		Class:    ClassStandard,
		Printers: []PrinterCapability{{Kind: Printer3D, Model: "Prusa MK4"}},
		GPUs:     []GPUCapability{{API: GPUVulkan, Model: "Adreno 750", VRAMMB: 8192}},
		Capacity: Capacity{VCPUs: 4, MemMB: 8192, DiskMB: 100000, StorageCommitMB: 500000},
		OptIn:    WorkloadOptIn{Compute: true, Print: true, Storage: true},
		IssuedAt: time.Unix(1_700_000_000, 0),
		Seq:      7,
	}
	l.Sign(priv)
	if !l.Verify(pub) {
		t.Fatal("valid signature rejected")
	}
}

func TestGPUTamperDetected(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := CapabilityListing{
		NodeID:   "node-1",
		Class:    ClassMicro,
		GPUs:     []GPUCapability{{API: GPUNNAPI, Model: "Tensor G4", VRAMMB: 4096}},
		IssuedAt: time.Unix(1, 0),
		Seq:      1,
	}
	l.Sign(priv)
	l.GPUs[0].VRAMMB = 24576 // inflate the advertised GPU after signing
	if l.Verify(pub) {
		t.Fatal("tampered GPU capability verified")
	}
}

func TestStorageCommitTamperDetected(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := CapabilityListing{NodeID: "node-1", Class: ClassMicro, Seq: 1, IssuedAt: time.Unix(1, 0)}
	l.Sign(priv)
	l.Capacity.StorageCommitMB = 1_000_000 // claim storage after signing
	if l.Verify(pub) {
		t.Fatal("tampered storage commitment verified")
	}
}

func TestListingTamperDetected(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := CapabilityListing{NodeID: "node-1", Class: ClassMicro, Seq: 1, IssuedAt: time.Unix(1, 0)}
	l.Sign(priv)
	l.OptIn.Compute = true // mutate a hint field after signing
	if l.Verify(pub) {
		t.Fatal("tampered listing verified")
	}
}

func TestSeqIsSigned(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	l := CapabilityListing{NodeID: "node-1", Class: ClassMicro, Seq: 1, IssuedAt: time.Unix(1, 0)}
	l.Sign(priv)
	l.Seq = 2 // rollback/replay attempt
	if l.Verify(pub) {
		t.Fatal("Seq not covered by signature")
	}
}

func TestWrongKeyRejected(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)
	other, _, _ := ed25519.GenerateKey(nil)
	l := CapabilityListing{NodeID: "node-1", Class: ClassMicro, Seq: 1, IssuedAt: time.Unix(1, 0)}
	l.Sign(priv)
	if l.Verify(other) {
		t.Fatal("signature verified under the wrong public key")
	}
}
