//go:build mxl_integration

package mxl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qvest-digital/go-mxl/mxl"
)

// abandonedFlow leaves smallFlowJSON in inst's domain with no writer
// holding it, the state a writer that died without releasing leaves.
func abandonedFlow(t *testing.T, inst *mxl.Instance) string {
	t.Helper()
	w, _, err := inst.NewWriter(smallFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if err := w.Detach(); err != nil {
		t.Fatalf("Detach: %v", err)
	}
	return flowDir(inst)
}

func TestGarbageCollectRunsOnlyWhenCalled(t *testing.T) {
	inst := newDomain(t)
	dir := abandonedFlow(t, inst)

	// Neither a new instance nor a new writer on the domain collects it.
	other, err := mxl.NewInstance(inst.Domain(), "")
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	t.Cleanup(func() { other.Close() })
	held, _, err := other.NewWriter(strings.ReplaceAll(smallFlowJSON, smallFlowID, otherFlowID))
	if err != nil {
		t.Fatalf("NewWriter for a second flow: %v", err)
	}
	t.Cleanup(func() { held.Close() })
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("the abandoned flow is gone without a collection: %v", err)
	}

	if err := inst.GarbageCollect(); err != nil {
		t.Fatalf("GarbageCollect: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("want the abandoned flow collected, stat gave %v", err)
	}
	if _, err := os.Stat(filepath.Join(inst.Domain(), otherFlowID+".mxl-flow")); err != nil {
		t.Fatalf("a flow its writer still holds was collected: %v", err)
	}
}

func TestGarbageCollectIgnoresReaders(t *testing.T) {
	inst := newDomain(t)
	dir := abandonedFlow(t, inst)
	r, err := inst.NewReader(smallFlowID)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	t.Cleanup(func() { r.Close() })

	if err := inst.GarbageCollect(); err != nil {
		t.Fatalf("GarbageCollect: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("want a flow open only for reading collected, stat gave %v", err)
	}
}

const otherFlowID = "1d7b6c40-0f2a-4a1e-9c33-6a2b5e4d0012"
