package fabrics

import (
	"errors"
	"testing"

	"github.com/qvest-digital/go-mxl/mxl"
)

func TestRegionsForFlowReaderNil(t *testing.T) {
	if _, err := RegionsForFlowReader(nil); !errors.Is(err, mxl.ErrInvalidArg) {
		t.Fatalf("RegionsForFlowReader(nil): %v, want ErrInvalidArg", err)
	}
}

func TestRegionsForFlowWriterNil(t *testing.T) {
	if _, err := RegionsForFlowWriter(nil); !errors.Is(err, mxl.ErrInvalidArg) {
		t.Fatalf("RegionsForFlowWriter(nil): %v, want ErrInvalidArg", err)
	}
}
