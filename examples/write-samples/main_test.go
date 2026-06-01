package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/qvest-digital/go-mxl/mxl"
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

func TestDefaultBatch(t *testing.T) {
	cases := []struct {
		name     string
		rate     mxl.Rational
		override int
		want     uint64
	}{
		{"48k auto ~10ms", mxl.Rational{Num: 48000, Den: 1}, 0, 480},
		{"explicit override", mxl.Rational{Num: 48000, Den: 1}, 256, 256},
		{"sub-100Hz clamps to 1", mxl.Rational{Num: 50, Den: 1}, 0, 1},
	}
	for _, c := range cases {
		if got := defaultBatch(c.rate, c.override); got != c.want {
			t.Errorf("%s: defaultBatch(%v, %d) = %d, want %d", c.name, c.rate, c.override, got, c.want)
		}
	}
}

func TestFillSampleRamp(t *testing.T) {
	// Ramp wraps at 256 and spans the two fragments contiguously; the
	// return value is the next sample value so callers can chain batches.
	f1 := make([]byte, 3)
	f2 := make([]byte, 2)
	next := fillSampleRamp(f1, f2, 254)
	if want := []byte{254, 255, 0}; !bytes.Equal(f1, want) {
		t.Errorf("f1 = %v, want %v", f1, want)
	}
	if want := []byte{1, 2}; !bytes.Equal(f2, want) {
		t.Errorf("f2 = %v, want %v", f2, want)
	}
	if next != 259 {
		t.Errorf("next = %d, want 259", next)
	}
}
