package mxl

import "testing"

func TestUUIDString(t *testing.T) {
	// The flow directory is named after this rendering, so a change
	// here silently stops every guard from finding its data file.
	id := [16]byte{
		0x1d, 0x7b, 0x6c, 0x40, 0x0f, 0x2a, 0x4a, 0x1e,
		0x9c, 0x33, 0x6a, 0x2b, 0x5e, 0x4d, 0x00, 0x11,
	}
	if got, want := uuidString(id), "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0011"; got != want {
		t.Fatalf("uuidString = %q, want %q", got, want)
	}
}

func TestUUIDStringPadsLeadingZeroes(t *testing.T) {
	var id [16]byte
	if got, want := uuidString(id), "00000000-0000-0000-0000-000000000000"; got != want {
		t.Fatalf("uuidString = %q, want %q", got, want)
	}
	id[0] = 0x01
	id[15] = 0x02
	if got, want := uuidString(id), "01000000-0000-0000-0000-000000000002"; got != want {
		t.Fatalf("uuidString = %q, want %q", got, want)
	}
}

func TestFlowDataPath(t *testing.T) {
	got := flowDataPath("/run/mxl/domain", "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0011")
	want := "/run/mxl/domain/1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0011.mxl-flow/data"
	if got != want {
		t.Fatalf("flowDataPath = %q, want %q", got, want)
	}
}
