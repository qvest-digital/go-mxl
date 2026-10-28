//go:build mxl_integration

package fabrics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/qvest-digital/go-mxl/fabrics"
	"github.com/qvest-digital/go-mxl/mxl"
)

// The state of a target across its lifetime on a tcp initiator: not
// held before it is added, connecting until progress sets the
// connection up, connected, closing once removed, and not held again
// once the shutdown completes. An initiator reports nothing else about
// a target it no longer holds.
func TestInitiatorTargetStateTCP(t *testing.T) {
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
		Reader: srcReader,
	}); err != nil {
		t.Fatalf("Initiator.Setup: %v", err)
	}

	if _, err := initiator.TargetState(info); !errors.Is(err, mxl.ErrNotFound) {
		t.Fatalf("TargetState before AddTarget: %v, want ErrNotFound", err)
	}
	if err := initiator.AddTarget(info); err != nil {
		t.Fatalf("AddTarget: %v", err)
	}
	if s, err := initiator.TargetState(info); err != nil || s != fabrics.TargetConnecting {
		t.Fatalf("TargetState after AddTarget: %v, %v; want connecting", s, err)
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

	if err := initiator.RemoveTarget(info); err != nil {
		t.Fatalf("RemoveTarget: %v", err)
	}
	if s, err := initiator.TargetState(info); err != nil || s != fabrics.TargetClosing {
		t.Fatalf("TargetState after RemoveTarget: %v, %v; want closing", s, err)
	}
	progressUntil("removed target was still held", func() bool {
		_, err := initiator.TargetState(info)
		return errors.Is(err, mxl.ErrNotFound)
	})
}
