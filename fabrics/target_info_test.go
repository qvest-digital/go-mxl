package fabrics

import (
	"testing"
)

// newTestTargetInfo brings up a Target on a fresh TCP endpoint and returns
// its TargetInfo. Skips when /dev/shm is unavailable.
func newTestTargetInfo(t *testing.T) *TargetInfo {
	t.Helper()
	_, fi, w := newTestFabrics(t)
	tgt, err := fi.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	t.Cleanup(func() { tgt.Close() })
	info, err := tgt.Setup(TargetConfig{
		Interface: InterfaceConfig{
			Provider: ProviderTCP,
			Address:  EndpointAddress{Node: "127.0.0.1", Service: "0"},
		},
		Writer: w,
	})
	if err != nil {
		t.Fatalf("Target.Setup: %v", err)
	}
	t.Cleanup(func() { info.Close() })
	return info
}

func TestParseTargetInfoInvalid(t *testing.T) {
	if _, err := ParseTargetInfo("not-a-target-info"); err == nil {
		t.Fatal("ParseTargetInfo(garbage) returned nil error")
	}
}

func TestParseTargetInfoEmpty(t *testing.T) {
	if _, err := ParseTargetInfo(""); err == nil {
		t.Fatal("ParseTargetInfo(\"\") returned nil error")
	}
}

func TestTargetInfoMarshalRoundTrip(t *testing.T) {
	info := newTestTargetInfo(t)
	s, err := info.MarshalString()
	if err != nil {
		t.Fatalf("MarshalString: %v", err)
	}
	if s == "" {
		t.Fatal("MarshalString returned empty string")
	}

	roundTripped, err := ParseTargetInfo(s)
	if err != nil {
		t.Fatalf("ParseTargetInfo: %v", err)
	}
	t.Cleanup(func() { roundTripped.Close() })

	s2, err := roundTripped.MarshalString()
	if err != nil {
		t.Fatalf("MarshalString of round-tripped: %v", err)
	}
	if s2 != s {
		t.Fatalf("round-trip mismatch:\n  got: %q\n want: %q", s2, s)
	}
}

func TestTargetInfoCloseIdempotent(t *testing.T) {
	info := newTestTargetInfo(t)
	if err := info.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := info.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestTargetInfoMarshalAfterClose(t *testing.T) {
	info := newTestTargetInfo(t)
	if err := info.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := info.MarshalString(); err == nil {
		t.Fatal("MarshalString after Close returned nil error")
	}
}
