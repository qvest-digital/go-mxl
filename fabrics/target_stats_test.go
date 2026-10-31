//go:build mxl_integration

package fabrics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/qvest-digital/go-mxl/fabrics"
	"github.com/qvest-digital/go-mxl/mxl"
)

// A tcp initiator holds at most "grainsInFlight" grains per target.
// Offered more while no completion is read, it skips the rest for that
// target, counts them and reports not-ready, and once progress reads
// the completions the counters show what was delivered.
func TestInitiatorTargetStatsTCP(t *testing.T) {
	srcInst := newDomain(t)
	tgtInst := newDomain(t)

	srcWriter, _, err := srcInst.NewWriter(flowJSON)
	if err != nil {
		t.Fatalf("src NewWriter: %v", err)
	}
	t.Cleanup(func() { srcWriter.Close() })
	srcReader, err := srcInst.NewReader(flowID)
	if err != nil {
		t.Fatalf("src NewReader: %v", err)
	}
	t.Cleanup(func() { srcReader.Close() })
	tgtWriter, _, err := tgtInst.NewWriter(flowJSON)
	if err != nil {
		t.Fatalf("tgt NewWriter: %v", err)
	}
	t.Cleanup(func() { tgtWriter.Close() })

	srcFab, err := fabrics.NewInstance(srcInst)
	if err != nil {
		t.Fatalf("src fabrics.NewInstance: %v", err)
	}
	t.Cleanup(func() { srcFab.Close() })
	tgtFab, err := fabrics.NewInstance(tgtInst)
	if err != nil {
		t.Fatalf("tgt fabrics.NewInstance: %v", err)
	}
	t.Cleanup(func() { tgtFab.Close() })

	target, err := tgtFab.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	t.Cleanup(func() { target.Close() })
	info, err := target.Setup(fabrics.TargetConfig{
		Interface: fabrics.InterfaceConfig{
			Provider: fabrics.ProviderTCP,
			Address:  fabrics.EndpointAddress{Node: "127.0.0.1", Service: freePort(t)},
		},
		Writer: tgtWriter,
	})
	if err != nil {
		t.Fatalf("Target.Setup: %v", err)
	}
	t.Cleanup(func() { info.Close() })

	initiator, err := srcFab.NewInitiator()
	if err != nil {
		t.Fatalf("NewInitiator: %v", err)
	}
	t.Cleanup(func() { initiator.Close() })
	if err := initiator.Setup(fabrics.InitiatorConfig{
		Interface: fabrics.InterfaceConfig{
			Provider: fabrics.ProviderTCP,
			Address:  fabrics.EndpointAddress{Node: "127.0.0.1", Service: freePort(t)},
		},
		Reader:  srcReader,
		Options: `{"grainsInFlight": 2}`,
	}); err != nil {
		t.Fatalf("Initiator.Setup: %v", err)
	}

	if _, err := initiator.TargetStats(info); !errors.Is(err, mxl.ErrNotFound) {
		t.Fatalf("TargetStats before AddTarget: %v, want ErrNotFound", err)
	}
	if err := initiator.AddTarget(info); err != nil {
		t.Fatalf("AddTarget: %v", err)
	}

	progressUntil := func(what string, done func() bool) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for !done() {
			if time.Now().After(deadline) {
				t.Fatalf("%s within 10s", what)
			}
			if _, err := target.ReadGrainNonBlocking(); err != nil && !errors.Is(err, fabrics.ErrNotReady) {
				t.Fatalf("target ReadGrainNonBlocking: %v", err)
			}
			if err := initiator.MakeProgressNonBlocking(); err != nil &&
				!errors.Is(err, fabrics.ErrNotReady) && !errors.Is(err, fabrics.ErrInterrupted) {
				t.Fatalf("initiator MakeProgressNonBlocking: %v", err)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	progressUntil("target did not connect", func() bool {
		s, err := initiator.TargetState(info)
		return err == nil && s == fabrics.TargetConnected
	})

	const offered = 5
	for idx := uint64(0); idx < offered; idx++ {
		err := initiator.TransferGrain(idx, 0, 1080)
		if idx < 2 && err != nil {
			t.Fatalf("TransferGrain(%d): %v", idx, err)
		}
		if idx >= 2 && !errors.Is(err, fabrics.ErrNotReady) {
			t.Fatalf("TransferGrain(%d) beyond the window: %v, want ErrNotReady", idx, err)
		}
	}
	s, err := initiator.TargetStats(info)
	if err != nil {
		t.Fatalf("TargetStats: %v", err)
	}
	if s.Posted != 2 || s.InFlight != 2 || s.Skipped != offered-2 || s.LastCompletedIndex != fabrics.NoIndex {
		t.Fatalf("TargetStats with no progress: %+v; want 2 posted and in flight, %d skipped, no completed index", s, offered-2)
	}

	progressUntil("grains in flight did not complete", func() bool {
		s, err = initiator.TargetStats(info)
		return err == nil && s.InFlight == 0
	})
	if s.Completed != 2 || s.LastCompletedIndex != 1 || s.Posted != s.Completed+s.Failed+s.InFlight {
		t.Fatalf("TargetStats after progress: %+v; want 2 completed up to index 1", s)
	}
}
