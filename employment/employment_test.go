package employment

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/NTARI-RAND/sohocloud-protocol/fees"
)

func TestJobReportRoundTripAndTamper(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	r := JobReport{
		JobID:      "job-1",
		NodeID:     "node-1",
		ExitCode:   0,
		StartedAt:  time.Unix(1_700_000_000, 0),
		FinishedAt: time.Unix(1_700_000_100, 0),
	}
	r.Sign(priv)
	if !r.Verify(pub) {
		t.Fatal("valid report rejected")
	}
	r.ExitCode = 1 // flip success->failure after signing
	if r.Verify(pub) {
		t.Fatal("tampered report verified")
	}
}

func TestAssignmentSignedByCoordinator(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	a := Assignment{
		JobID:     "job-1",
		NodeID:    "node-1",
		Spec:      JobSpec{Workload: "compute", Image: "busybox", Args: []string{"echo", "hi"}},
		Fee:       fees.Terms{ContributorShareBps: 6500, PlatformFeeBps: 3500},
		OfferedAt: time.Unix(1, 0),
	}
	a.Sign(priv)
	if !a.Verify(pub) {
		t.Fatal("valid assignment rejected")
	}
	a.Spec.Image = "malicious" // swap the image after signing
	if a.Verify(pub) {
		t.Fatal("tampered assignment verified")
	}
}

func TestDeclineRoundTrip(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	d := Decline{JobID: "job-1", NodeID: "node-1", Reason: DeclineLocalPolicy, DeclinedAt: time.Unix(1, 0)}
	d.Sign(priv)
	if !d.Verify(pub) {
		t.Fatal("valid decline rejected")
	}
}
