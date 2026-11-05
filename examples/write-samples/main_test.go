package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
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

const audioFlowDef = `{"description":"write-samples test","format":"urn:x-nmos:format:audio",
"label":"write-samples test","id":"b3bb5be7-9fe9-4324-a5bb-4c70e1084449",
"tags":{"urn:x-nmos:tag:grouphint/v1.0":["write-samples test:Audio"]},
"media_type":"audio/float32","sample_rate":{"numerator":48000},"channel_count":2,"bit_depth":32,"parents":[]}`

// runLogged runs write-samples for one batch on a fresh domain and
// returns what it logged.
func runLogged(t *testing.T, args ...string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	def := filepath.Join(dir, "flow.json")
	if err := os.WriteFile(def, []byte(audioFlowDef), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	log.SetOutput(&out)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	domain := filepath.Join(dir, "domain")
	if err := os.Mkdir(domain, 0o755); err != nil {
		t.Fatal(err)
	}
	err := run(append([]string{"-domain", domain, "-flow-def", def, "-count", "480"}, args...), io.Discard)
	return out.String(), err
}

// The options set the batch size hints of the flow the writer creates,
// which a reader of the flow, a fabrics target among them, is sized by.
func TestRunOptionsSetTheFlowsHints(t *testing.T) {
	out, err := runLogged(t, "-options", `{"maxSyncBatchSizeHint":960,"maxCommitBatchSizeHint":480}`)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if !strings.Contains(out, "syncHint=960 commitHint=480") {
		t.Errorf("log = %q, want syncHint=960 commitHint=480", out)
	}
}

func TestRunWithoutOptionsKeepsLibmxlDefaults(t *testing.T) {
	out, err := runLogged(t)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if !strings.Contains(out, "syncHint=480 commitHint=480") {
		t.Errorf("log = %q, want libmxl's 10 ms default syncHint=480 commitHint=480", out)
	}
}

func TestRunRejectsInvalidOptions(t *testing.T) {
	if _, err := runLogged(t, "-options", `{"maxSyncBatchSizeHint":0}`); err == nil {
		t.Fatal("run with a zero sync hint = nil, want libmxl's rejection")
	}
}
