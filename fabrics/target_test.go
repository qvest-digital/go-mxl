package fabrics

import (
	"errors"
	"testing"
	"time"

	"github.com/qvest-digital/go-mxl/mxl"
)

func TestTargetMethodsAfterClose(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	tgt, err := fi.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	if err := tgt.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := tgt.ReadGrain(time.Millisecond); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("ReadGrain after Close: %v, want ErrClosed", err)
	}
	if _, err := tgt.ReadGrainNonBlocking(); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("ReadGrainNonBlocking after Close: %v, want ErrClosed", err)
	}
	if _, _, err := tgt.ReadSamples(time.Millisecond); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("ReadSamples after Close: %v, want ErrClosed", err)
	}
	if _, _, err := tgt.ReadSamplesNonBlocking(); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("ReadSamplesNonBlocking after Close: %v, want ErrClosed", err)
	}
}
