package mxl

import (
	"bytes"
	"testing"
)

func TestGrainComplete(t *testing.T) {
	cases := []struct {
		g    Grain
		want bool
	}{
		{Grain{}, false},
		{Grain{TotalSlices: 0, ValidSlices: 0}, false},
		{Grain{TotalSlices: 4, ValidSlices: 0}, false},
		{Grain{TotalSlices: 4, ValidSlices: 3}, false},
		{Grain{TotalSlices: 4, ValidSlices: 4}, true},
	}
	for _, c := range cases {
		if got := c.g.Complete(); got != c.want {
			t.Errorf("Grain%+v.Complete() = %v, want %v", c.g, got, c.want)
		}
	}
}

func TestGrainInvalid(t *testing.T) {
	if (Grain{}).Invalid() {
		t.Error("zero Grain is not invalid")
	}
	if !(Grain{Flags: GrainFlagInvalid}).Invalid() {
		t.Error("Grain with GrainFlagInvalid set should be invalid")
	}
	if !(Grain{Flags: GrainFlagInvalid | 0x80}).Invalid() {
		t.Error("Grain with mixed flags including GrainFlagInvalid should be invalid")
	}
}

func TestGrainCopyEmpty(t *testing.T) {
	if got := (Grain{}).Copy(); got != nil {
		t.Errorf("empty Grain.Copy() = %v, want nil", got)
	}
}

func TestGrainCopyIndependent(t *testing.T) {
	src := []byte{1, 2, 3, 4}
	g := Grain{Payload: src}
	cp := g.Copy()
	if !bytes.Equal(cp, src) {
		t.Fatalf("Copy() = %v, want %v", cp, src)
	}
	cp[0] = 0xFF
	if src[0] == 0xFF {
		t.Fatal("Copy() returned a slice aliasing the source")
	}
}
