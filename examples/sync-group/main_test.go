package main

import (
	"errors"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/qvest-digital/go-mxl/mxl"
)

func TestRunMissingFlow(t *testing.T) {
	err := run(nil, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "at least one -flow") {
		t.Fatalf("run(nil) = %v, want missing -flow error", err)
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

// A bad -rate is rejected before any domain is opened, so it surfaces as a
// plain run() error even without a live MXL instance.
func TestRunBadRate(t *testing.T) {
	err := run([]string{"-flow", "f", "-rate", "notarate"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "parse -rate") {
		t.Fatalf("run(bad -rate) = %v, want parse -rate error", err)
	}
}

func TestParseRate(t *testing.T) {
	got, err := parseRate("30000/1001")
	if err != nil {
		t.Fatalf("parseRate(valid) error: %v", err)
	}
	if got != (mxl.Rational{Num: 30000, Den: 1001}) {
		t.Fatalf("parseRate(30000/1001) = %v, want {30000 1001}", got)
	}
	for _, bad := range []string{"", "30000", "30000/", "/1001", "a/b", "30000/1001/1"} {
		if _, err := parseRate(bad); err == nil {
			t.Errorf("parseRate(%q) = nil error, want failure", bad)
		}
	}
}

func TestFlowListSetString(t *testing.T) {
	var fl flowList
	for _, id := range []string{"a", "b", "c"} {
		if err := fl.Set(id); err != nil {
			t.Fatalf("Set(%q): %v", id, err)
		}
	}
	if len(fl) != 3 {
		t.Fatalf("len = %d, want 3", len(fl))
	}
	if got := fl.String(); got != "a,b,c" {
		t.Fatalf("String() = %q, want \"a,b,c\"", got)
	}
}
