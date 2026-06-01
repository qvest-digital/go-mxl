package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
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

// With a readable flow definition present, an unknown -provider is rejected
// by ParseProvider before any MXL instance is opened, so the failure is
// observable without a live domain.
func TestRunBadProvider(t *testing.T) {
	def := filepath.Join(t.TempDir(), "flow.json")
	if err := os.WriteFile(def, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write flow def: %v", err)
	}
	err := run([]string{"-flow-def", def, "-provider", "not-a-provider"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("run(bad -provider) = %v, want provider error", err)
	}
}
