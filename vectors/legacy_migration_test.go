package vectors_test

import (
	"crypto/ed25519"
	"encoding/json"
	"os"
	"testing"

	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
)

// These signatures were generated and verified by the actual v0.1.0 module.
// Earlier releases changed listing/assignment bytes, so upgrading consumers
// must not assume that old nodes and new coordinators can interoperate.
func TestLegacySignaturesRequireCoordinatedMigration(t *testing.T) {
	data, err := os.ReadFile("../testdata/legacy_v0_1_signatures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		PublicKey  ed25519.PublicKey
		Listing    listing.CapabilityListing
		Assignment employment.Assignment
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Listing.Verify(fixture.PublicKey) {
		t.Fatal("legacy listing unexpectedly verified; reassess migration boundary")
	}
	if fixture.Assignment.Verify(fixture.PublicKey) {
		t.Fatal("legacy assignment unexpectedly verified; reassess migration boundary")
	}
}
