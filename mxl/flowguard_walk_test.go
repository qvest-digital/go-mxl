//go:build mxl_integration

package mxl

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

const walkFlowJSON = `{
  "description": "go-mxl walk test",
  "id": "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0021",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl walk test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl walk test:Video"] },
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

// A flow directory removed underneath a live writer leaves Close a path
// to a name that is gone. On a domain that waits for a missing flow name
// to appear, resolving it blocks Close for the whole wait, and Close
// already knows the answer from the file it holds.
func TestCloseDoesNotResolveTheNameOfARemovedFlow(t *testing.T) {
	inst := newWalkDomain(t)
	w, _, err := inst.NewWriter(walkFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	dir := filepath.Join(inst.Domain(), "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0021.mxl-flow")
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove flow dir: %v", err)
	}

	var walked []string
	statPath = func(path string, st *syscall.Stat_t) error {
		walked = append(walked, path)
		return syscall.Stat(path, st)
	}
	t.Cleanup(func() { statPath = syscall.Stat })

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if len(walked) != 0 {
		t.Fatalf("Close resolved %v after its flow directory was removed", walked)
	}
}

// A flow still in place is checked by path, which is what tells Close
// that the directory has not been replaced.
func TestCloseResolvesTheNameOfAFlowInPlace(t *testing.T) {
	inst := newWalkDomain(t)
	w, _, err := inst.NewWriter(walkFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	var walked []string
	statPath = func(path string, st *syscall.Stat_t) error {
		walked = append(walked, path)
		return syscall.Stat(path, st)
	}
	t.Cleanup(func() { statPath = syscall.Stat })

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if len(walked) != 1 {
		t.Fatalf("want one path resolved for a flow in place, got %v", walked)
	}
}

// newWalkDomain opens an instance on a fresh domain in /dev/shm, as the
// external integration tests do.
func newWalkDomain(t *testing.T) *Instance {
	t.Helper()
	dir, err := os.MkdirTemp("/dev/shm", "go-mxl-walk-*")
	if err != nil {
		t.Skipf("cannot create dir in /dev/shm: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	inst, err := NewInstance(dir, "")
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	t.Cleanup(func() { inst.Close() })
	return inst
}
