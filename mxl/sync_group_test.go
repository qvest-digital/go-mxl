package mxl

import (
	"errors"
	"testing"
	"time"
)

func newTestSyncGroup(t *testing.T, inst *Instance) *SyncGroup {
	t.Helper()
	g, err := inst.NewSyncGroup()
	if err != nil {
		t.Fatalf("NewSyncGroup: %v", err)
	}
	t.Cleanup(func() { g.Close() })
	return g
}

func TestSyncGroupCloseIdempotent(t *testing.T) {
	inst := newTestInstance(t)
	g := newTestSyncGroup(t, inst)
	if err := g.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := g.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestSyncGroupAddRemoveReader(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	g := newTestSyncGroup(t, inst)

	if err := g.AddReader(r); err != nil {
		t.Fatalf("AddReader: %v", err)
	}
	// Re-adding the same reader is documented as updating its config in
	// place; libmxl should not surface an error.
	if err := g.AddReader(r); err != nil {
		t.Fatalf("re-AddReader: %v", err)
	}
	if err := g.RemoveReader(r); err != nil {
		t.Fatalf("RemoveReader: %v", err)
	}
}

func TestSyncGroupAddPartialGrainReader(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	g := newTestSyncGroup(t, inst)

	if err := g.AddPartialGrainReader(r, GrainValidSlicesAny); err != nil {
		t.Fatalf("AddPartialGrainReader(any): %v", err)
	}
	if err := g.AddPartialGrainReader(r, GrainValidSlicesAll); err != nil {
		t.Fatalf("AddPartialGrainReader(all): %v", err)
	}
}

func TestSyncGroupAddClosedReader(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	if err := r.Close(); err != nil {
		t.Fatalf("Reader.Close: %v", err)
	}
	g := newTestSyncGroup(t, inst)
	if err := g.AddReader(r); !errors.Is(err, ErrClosed) {
		t.Errorf("AddReader(closed): %v, want ErrClosed", err)
	}
	if err := g.AddPartialGrainReader(r, GrainValidSlicesAll); !errors.Is(err, ErrClosed) {
		t.Errorf("AddPartialGrainReader(closed): %v, want ErrClosed", err)
	}
	if err := g.RemoveReader(r); !errors.Is(err, ErrClosed) {
		t.Errorf("RemoveReader(closed): %v, want ErrClosed", err)
	}
}

func TestSyncGroupMethodsAfterClose(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	g := newTestSyncGroup(t, inst)
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := g.AddReader(r); !errors.Is(err, ErrClosed) {
		t.Errorf("AddReader after Close: %v, want ErrClosed", err)
	}
	if err := g.AddPartialGrainReader(r, GrainValidSlicesAll); !errors.Is(err, ErrClosed) {
		t.Errorf("AddPartialGrainReader after Close: %v, want ErrClosed", err)
	}
	if err := g.RemoveReader(r); !errors.Is(err, ErrClosed) {
		t.Errorf("RemoveReader after Close: %v, want ErrClosed", err)
	}
	if err := g.WaitForDataAt(Now(), time.Millisecond); !errors.Is(err, ErrClosed) {
		t.Errorf("WaitForDataAt after Close: %v, want ErrClosed", err)
	}
}
