//go:build mxl_integration

package mxl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qvest-digital/go-mxl/mxl"
)

// A deliberately small flow: these tests care about the flow directory's
// existence, not its contents, and several of them hold two rings on one
// domain at once.
const smallFlowJSON = `{
  "description": "go-mxl detach test",
  "id": "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0011",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl detach test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl detach test:Video"] },
  "parents": [],
  "media_type": "video/v210",
  "grain_rate": { "numerator": 30, "denominator": 1 },
  "frame_width": 128,
  "frame_height": 72,
  "interlace_mode": "progressive",
  "colorspace": "BT709",
  "components": [
    { "name": "Y",  "width": 128, "height": 72, "bit_depth": 10 },
    { "name": "Cb", "width": 64,  "height": 72, "bit_depth": 10 },
    { "name": "Cr", "width": 64,  "height": 72, "bit_depth": 10 }
  ]
}`

const smallFlowID = "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0011"

func flowDir(inst *mxl.Instance) string {
	return filepath.Join(inst.Domain(), smallFlowID+".mxl-flow")
}

func TestCloseDeletesTheFlow(t *testing.T) {
	// The delete-on-last-release behaviour Detach exists to opt out of.
	inst := newDomain(t)
	w, _, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	dir := flowDir(inst)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("flow directory missing while the writer is live: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("want the flow removed after Close, stat gave %v", err)
	}
}

func TestDetachKeepsTheFlow(t *testing.T) {
	inst := newDomain(t)
	w, _, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	dir := flowDir(inst)

	if err := w.Detach(); err != nil {
		t.Fatalf("Detach: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("want the flow kept after Detach, stat gave %v", err)
	}

	// A detached flow is still openable, which is the point: something
	// other than this writer owns it now.
	w2, created, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter after Detach: %v", err)
	}
	if created {
		t.Fatal("want the detached flow reopened, not recreated")
	}
	if err := w2.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestDetachIsIdempotent(t *testing.T) {
	inst := newDomain(t)
	w, _, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if err := w.Detach(); err != nil {
		t.Fatalf("first Detach: %v", err)
	}
	if err := w.Detach(); err != nil {
		t.Fatalf("second Detach: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close after Detach: %v", err)
	}
	if _, err := os.Stat(flowDir(inst)); err != nil {
		t.Fatalf("a released writer must not delete the flow later: %v", err)
	}
}

func TestCloseKeepsAFlowThatReplacedItsOwn(t *testing.T) {
	// libmxl decides whether to delete on the releasing writer's own
	// descriptor but performs the delete by path. Once the flow
	// directory has been replaced underneath a live writer the two name
	// different flows, and releasing the old writer would remove the
	// new flow along with the directory its writer is publishing into.
	inst := newDomain(t)
	first, _, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	dir := flowDir(inst)

	// Replace the directory the way an external cleanup would, then let
	// a second writer create the flow again at the same path.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove flow dir: %v", err)
	}
	second, created, err := mustSecondWriter(t, inst.Domain())
	if err != nil {
		t.Fatalf("second NewWriter: %v", err)
	}
	if !created {
		t.Fatal("want the second writer to create a fresh flow")
	}
	t.Cleanup(func() { second.Close() })

	if err := first.Close(); err != nil {
		t.Fatalf("Close of the superseded writer: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("the live writer's flow was removed by the superseded writer's Close: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data")); err != nil {
		t.Fatalf("the live writer's data file was removed: %v", err)
	}
}

// mustSecondWriter opens an independent instance on the same domain so
// the two writers do not share libmxl's per-instance writer table.
func mustSecondWriter(t *testing.T, domain string) (*mxl.Writer, bool, error) {
	t.Helper()
	inst, err := mxl.NewInstance(domain, "")
	if err != nil {
		return nil, false, err
	}
	t.Cleanup(func() { inst.Close() })
	return inst.NewWriter(smallFlowJSON)
}

func TestDetachKeepsAFlowThatReplacedItsOwn(t *testing.T) {
	// The same collision as above along the path a caller takes when it
	// already knows something else owns the flow.
	inst := newDomain(t)
	first, _, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	dir := flowDir(inst)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove flow dir: %v", err)
	}
	second, _, err := mustSecondWriter(t, inst.Domain())
	if err != nil {
		t.Fatalf("second NewWriter: %v", err)
	}
	t.Cleanup(func() { second.Close() })

	if err := first.Detach(); err != nil {
		t.Fatalf("Detach of the superseded writer: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data")); err != nil {
		t.Fatalf("the live writer's flow was removed by the superseded writer's Detach: %v", err)
	}
}
