package main

import (
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestRunMissingFlow(t *testing.T) {
	err := run(nil, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "missing -flow") {
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

// An unknown -provider is rejected before the target-info file is read or
// any MXL instance is opened, so the failure surfaces without a live domain.
func TestRunBadProvider(t *testing.T) {
	err := run([]string{"-flow", "f", "-provider", "not-a-provider"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("run(bad -provider) = %v, want provider error", err)
	}
}
