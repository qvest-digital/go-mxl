package fabrics_test

import (
	"errors"
	"testing"

	"github.com/qvest-digital/go-mxl/fabrics"
	"github.com/qvest-digital/go-mxl/mxl"
)

func TestErrNotReadyMatchesMxlStatusErrNotReady(t *testing.T) {
	var err error = mxl.StatusErrNotReady
	if !errors.Is(err, fabrics.ErrNotReady) {
		t.Fatal("errors.Is(mxl.StatusErrNotReady, fabrics.ErrNotReady) = false")
	}
}

func TestErrNotReadyMatchesInverse(t *testing.T) {
	if !errors.Is(fabrics.ErrNotReady, mxl.StatusErrNotReady) {
		t.Fatal("errors.Is(fabrics.ErrNotReady, mxl.StatusErrNotReady) = false")
	}
}

func TestErrInterruptedMatchesMxlStatusErrInterrupted(t *testing.T) {
	var err error = mxl.StatusErrInterrupted
	if !errors.Is(err, fabrics.ErrInterrupted) {
		t.Fatal("errors.Is(mxl.StatusErrInterrupted, fabrics.ErrInterrupted) = false")
	}
}

func TestErrInterruptedMatchesInverse(t *testing.T) {
	if !errors.Is(fabrics.ErrInterrupted, mxl.StatusErrInterrupted) {
		t.Fatal("errors.Is(fabrics.ErrInterrupted, mxl.StatusErrInterrupted) = false")
	}
}

// The two sentinels describe different outcomes, so a caller branching on one
// must not match the other.
func TestErrNotReadyAndErrInterruptedAreDistinct(t *testing.T) {
	if errors.Is(fabrics.ErrInterrupted, fabrics.ErrNotReady) {
		t.Fatal("errors.Is(fabrics.ErrInterrupted, fabrics.ErrNotReady) = true")
	}
	if errors.Is(fabrics.ErrNotReady, fabrics.ErrInterrupted) {
		t.Fatal("errors.Is(fabrics.ErrNotReady, fabrics.ErrInterrupted) = true")
	}
}
