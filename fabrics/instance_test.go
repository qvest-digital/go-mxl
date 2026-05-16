package fabrics

import (
	"errors"
	"os"
	"testing"

	"github.com/qvest-digital/go-mxl/mxl"
)

// newTestMxlInstance opens a parent mxl.Instance under /dev/shm. Tests that
// need a fabrics Instance pair use this helper; on systems without /dev/shm
// the test is skipped.
func newTestMxlInstance(t *testing.T) *mxl.Instance {
	t.Helper()
	if _, err := os.Stat("/dev/shm"); err != nil {
		t.Skip("/dev/shm not present")
	}
	dir, err := os.MkdirTemp("/dev/shm", "go-mxl-fabrics-test-*")
	if err != nil {
		t.Skipf("cannot create dir in /dev/shm: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	inst, err := mxl.NewInstance(dir, "")
	if err != nil {
		t.Fatalf("mxl.NewInstance(%q): %v", dir, err)
	}
	t.Cleanup(func() { inst.Close() })
	return inst
}

func TestNewInstanceNil(t *testing.T) {
	if _, err := NewInstance(nil); !errors.Is(err, mxl.ErrInvalidArg) {
		t.Fatalf("NewInstance(nil): %v, want ErrInvalidArg", err)
	}
}

func TestNewInstanceClosedParent(t *testing.T) {
	parent := newTestMxlInstance(t)
	if err := parent.Close(); err != nil {
		t.Fatalf("parent.Close: %v", err)
	}
	if _, err := NewInstance(parent); !errors.Is(err, mxl.ErrClosed) {
		t.Fatalf("NewInstance(closed parent): %v, want ErrClosed", err)
	}
}

func TestInstanceCloseIdempotent(t *testing.T) {
	parent := newTestMxlInstance(t)
	fi, err := NewInstance(parent)
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	if err := fi.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := fi.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestNewTargetAfterClose(t *testing.T) {
	parent := newTestMxlInstance(t)
	fi, err := NewInstance(parent)
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	if err := fi.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := fi.NewTarget(); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("NewTarget after Close: %v, want ErrClosed", err)
	}
	if _, err := fi.NewInitiator(); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("NewInitiator after Close: %v, want ErrClosed", err)
	}
}
