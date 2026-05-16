package mxl

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLibVersion(t *testing.T) {
	v, err := LibVersion()
	if err != nil {
		t.Fatalf("LibVersion: %v", err)
	}
	if v.Full == "" {
		t.Fatal("LibVersion returned empty Full string")
	}
	if v.Major == 0 && v.Minor == 0 && v.Bugfix == 0 && v.Build == 0 {
		t.Fatalf("LibVersion returned zero version: %+v", v)
	}
}

func TestIsTmpFsMissingPath(t *testing.T) {
	if _, err := IsTmpFs(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatal("IsTmpFs on missing path returned nil error")
	}
}

func TestIsTmpFsDevShm(t *testing.T) {
	if _, err := os.Stat("/dev/shm"); err != nil {
		t.Skip("/dev/shm not present")
	}
	ok, err := IsTmpFs("/dev/shm")
	if err != nil {
		t.Fatalf("IsTmpFs(/dev/shm): %v", err)
	}
	if !ok {
		t.Fatal("/dev/shm reported as non-tmpfs")
	}
}

func TestNewInstanceMissingDomain(t *testing.T) {
	if _, err := NewInstance(filepath.Join(t.TempDir(), "missing-domain"), ""); err == nil {
		t.Fatal("NewInstance on missing domain returned nil error")
	}
}

// newTestInstance opens a domain under /dev/shm so libmxl's tmpfs check is
// satisfied. Skips when /dev/shm is unavailable.
func newTestInstance(t *testing.T) *Instance {
	t.Helper()
	if _, err := os.Stat("/dev/shm"); err != nil {
		t.Skip("/dev/shm not present")
	}
	dir, err := os.MkdirTemp("/dev/shm", "go-mxl-test-*")
	if err != nil {
		t.Skipf("cannot create dir in /dev/shm: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	inst, err := NewInstance(dir, "")
	if err != nil {
		t.Fatalf("NewInstance(%q): %v", dir, err)
	}
	t.Cleanup(func() { inst.Close() })
	return inst
}

func TestInstanceCloseIdempotent(t *testing.T) {
	inst := newTestInstance(t)
	if err := inst.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := inst.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestInstanceMethodsAfterClose(t *testing.T) {
	inst := newTestInstance(t)
	if err := inst.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := inst.GarbageCollect(); !errors.Is(err, ErrClosed) {
		t.Errorf("GarbageCollect after Close: %v, want ErrClosed", err)
	}
	if _, err := inst.IsFlowActive("00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrClosed) {
		t.Errorf("IsFlowActive after Close: %v, want ErrClosed", err)
	}
	if _, err := inst.FlowDef("00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrClosed) {
		t.Errorf("FlowDef after Close: %v, want ErrClosed", err)
	}
	if _, err := inst.NewReader("00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrClosed) {
		t.Errorf("NewReader after Close: %v, want ErrClosed", err)
	}
	if _, _, err := inst.NewWriter("{}"); !errors.Is(err, ErrClosed) {
		t.Errorf("NewWriter after Close: %v, want ErrClosed", err)
	}
	if _, err := inst.NewSyncGroup(); !errors.Is(err, ErrClosed) {
		t.Errorf("NewSyncGroup after Close: %v, want ErrClosed", err)
	}
}

func TestInstanceIsFlowActiveMissing(t *testing.T) {
	inst := newTestInstance(t)
	_, err := inst.IsFlowActive("00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("IsFlowActive on missing flow: %v, want ErrFlowNotFound", err)
	}
}

func TestInstanceFlowDefMissing(t *testing.T) {
	inst := newTestInstance(t)
	if _, err := inst.FlowDef("00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("FlowDef on missing flow: %v, want ErrFlowNotFound", err)
	}
}

func TestInstanceGarbageCollect(t *testing.T) {
	inst := newTestInstance(t)
	if err := inst.GarbageCollect(); err != nil {
		t.Fatalf("GarbageCollect on fresh instance: %v", err)
	}
}
