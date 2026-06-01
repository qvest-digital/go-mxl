package main

import (
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestRunMissingFlowDef(t *testing.T) {
	err := run(nil, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "missing -flow-def") {
		t.Fatalf("run(nil) = %v, want missing -flow-def error", err)
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"-h"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("run(-h) = %v, want flag.ErrHelp", err)
	}
}

func TestRunUnknownFlag(t *testing.T) {
	if err := run([]string{"-nope"}, io.Discard); err == nil {
		t.Fatal("run(-nope) = nil, want a parse error")
	}
}

func TestFillGrainRamp(t *testing.T) {
	// The ramp must wrap at 256 and start from the grain index, so a
	// reader can recompute byte((i+idx)&0xFF) and detect gaps.
	buf := make([]byte, 5)
	fillGrainRamp(buf, 254)
	want := []byte{254, 255, 0, 1, 2}
	for i, w := range want {
		if buf[i] != w {
			t.Fatalf("fillGrainRamp(start=254) = %v, want %v", buf, want)
		}
	}
}
