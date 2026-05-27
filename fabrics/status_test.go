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
