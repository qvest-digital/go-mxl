package fabrics

import (
	"errors"
	"testing"
	"time"

	"github.com/qvest-digital/go-mxl/mxl"
)

const testFlowJSON = `{
  "description": "go-mxl fabrics unit test, 1080p29",
  "id": "5fbec3b1-1b0f-417d-9059-8b94a47197ed",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl fabrics unit test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl fabrics unit test:Video"] },
  "parents": [],
  "media_type": "video/v210",
  "grain_rate": { "numerator": 30000, "denominator": 1001 },
  "frame_width": 1920,
  "frame_height": 1080,
  "interlace_mode": "progressive",
  "colorspace": "BT709",
  "components": [
    { "name": "Y",  "width": 1920, "height": 1080, "bit_depth": 10 },
    { "name": "Cb", "width": 960,  "height": 1080, "bit_depth": 10 },
    { "name": "Cr", "width": 960,  "height": 1080, "bit_depth": 10 }
  ]
}`

// newTestFabrics returns a parent mxl.Instance plus a fabrics.Instance and
// a writer on the test flow, all cleaned up at test end. Skips when
// /dev/shm is unavailable (via newTestMxlInstance).
func newTestFabrics(t *testing.T) (*mxl.Instance, *Instance, *mxl.Writer) {
	t.Helper()
	parent := newTestMxlInstance(t)
	w, _, err := parent.NewWriter(testFlowJSON)
	if err != nil {
		t.Fatalf("parent.NewWriter: %v", err)
	}
	t.Cleanup(func() { w.Close() })
	fi, err := NewInstance(parent)
	if err != nil {
		t.Fatalf("fabrics.NewInstance: %v", err)
	}
	t.Cleanup(func() { fi.Close() })
	return parent, fi, w
}

func TestInitiatorCloseIdempotent(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	in, err := fi.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	if err := in.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := in.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestInitiatorSetupNilRegions(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	in, err := fi.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	t.Cleanup(func() { in.Close() })
	err = in.Setup(InitiatorConfig{
		Endpoint: EndpointAddress{Node: "127.0.0.1", Service: "0"},
		Provider: ProviderTCP,
		Regions:  nil,
	})
	if !errors.Is(err, mxl.ErrInvalidArg) {
		t.Fatalf("Setup(nil Regions): %v, want ErrInvalidArg", err)
	}
}

func TestInitiatorSetupClosedRegions(t *testing.T) {
	_, fi, w := newTestFabrics(t)
	regs, err := RegionsForFlowWriter(w)
	if err != nil {
		t.Fatalf("RegionsForFlowWriter: %v", err)
	}
	if err := regs.Close(); err != nil {
		t.Fatalf("Regions.Close: %v", err)
	}
	in, err := fi.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	t.Cleanup(func() { in.Close() })
	err = in.Setup(InitiatorConfig{
		Endpoint: EndpointAddress{Node: "127.0.0.1", Service: "0"},
		Provider: ProviderTCP,
		Regions:  regs,
	})
	if !errors.Is(err, mxl.ErrClosed) {
		t.Fatalf("Setup(closed Regions): %v, want ErrClosed", err)
	}
}

func TestInitiatorAddRemoveTargetNil(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	in, err := fi.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	t.Cleanup(func() { in.Close() })
	if err := in.AddTarget(nil); !errors.Is(err, mxl.ErrInvalidArg) {
		t.Errorf("AddTarget(nil): %v, want ErrInvalidArg", err)
	}
	if err := in.RemoveTarget(nil); !errors.Is(err, mxl.ErrInvalidArg) {
		t.Errorf("RemoveTarget(nil): %v, want ErrInvalidArg", err)
	}
}

func TestInitiatorAddTargetClosedInfo(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	in, err := fi.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	t.Cleanup(func() { in.Close() })

	// Build a TargetInfo via Setup on a separate Target, then close it.
	parent, fi2, w := newTestFabrics(t)
	_ = parent
	regs, err := RegionsForFlowWriter(w)
	if err != nil {
		t.Fatalf("RegionsForFlowWriter: %v", err)
	}
	t.Cleanup(func() { regs.Close() })
	tgt, err := fi2.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	t.Cleanup(func() { tgt.Close() })
	info, err := tgt.Setup(TargetConfig{
		Endpoint: EndpointAddress{Node: "127.0.0.1", Service: "0"},
		Provider: ProviderTCP,
		Regions:  regs,
	})
	if err != nil {
		t.Fatalf("Target.Setup: %v", err)
	}
	if err := info.Close(); err != nil {
		t.Fatalf("TargetInfo.Close: %v", err)
	}

	if err := in.AddTarget(info); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("AddTarget(closed): %v, want ErrClosed", err)
	}
	if err := in.RemoveTarget(info); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("RemoveTarget(closed): %v, want ErrClosed", err)
	}
}

func TestInitiatorMethodsAfterClose(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	in, err := fi.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	if err := in.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := in.Setup(InitiatorConfig{
		Endpoint: EndpointAddress{Node: "127.0.0.1", Service: "0"},
		Provider: ProviderTCP,
	}); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("Setup after Close: %v, want ErrClosed", err)
	}
	if err := in.AddTarget(&TargetInfo{}); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("AddTarget after Close: %v, want ErrClosed", err)
	}
	if err := in.RemoveTarget(&TargetInfo{}); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("RemoveTarget after Close: %v, want ErrClosed", err)
	}
	if err := in.TransferGrain(0, 0, 1); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("TransferGrain after Close: %v, want ErrClosed", err)
	}
	if err := in.MakeProgress(time.Millisecond); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("MakeProgress after Close: %v, want ErrClosed", err)
	}
	if err := in.MakeProgressNonBlocking(); !errors.Is(err, mxl.ErrClosed) {
		t.Errorf("MakeProgressNonBlocking after Close: %v, want ErrClosed", err)
	}
}
