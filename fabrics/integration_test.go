//go:build mxl_integration

package fabrics_test

import (
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/qvest-digital/go-mxl/fabrics"
	"github.com/qvest-digital/go-mxl/mxl"
)

const flowJSON = `{
  "description": "go-mxl fabrics test, 1080p29",
  "id": "5fbec3b1-1b0f-417d-9059-8b94a47197ed",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl fabrics test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl fabrics test:Video"] },
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

const flowID = "5fbec3b1-1b0f-417d-9059-8b94a47197ed"

func newDomain(t *testing.T) *mxl.Instance {
	t.Helper()
	if _, err := os.Stat("/dev/shm"); err != nil {
		t.Skip("/dev/shm not present")
	}
	dir, err := os.MkdirTemp("/dev/shm", "go-mxl-fab-it-*")
	if err != nil {
		t.Skipf("cannot create dir in /dev/shm: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	inst, err := mxl.NewInstance(dir, "")
	if err != nil {
		t.Fatalf("NewInstance(%q): %v", dir, err)
	}
	t.Cleanup(func() { inst.Close() })
	return inst
}

// freePort asks the kernel for an unused TCP port on loopback.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer l.Close()
	_, port, _ := net.SplitHostPort(l.Addr().String())
	if _, err := strconv.Atoi(port); err != nil {
		t.Fatalf("parse port %q: %v", port, err)
	}
	return port
}

// TestFabricsGrainTransferTCP exercises the full target+initiator lifecycle
// over libfabric's TCP provider on loopback. Two MXL domains are spun up
// in the same process; a writer fills a grain, the initiator transfers it
// to the target, and the receiving reader observes the same payload.
func TestFabricsGrainTransferTCP(t *testing.T) {
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

	tgtRegs, err := fabrics.RegionsForFlowWriter(tgtWriter)
	if err != nil {
		t.Fatalf("RegionsForFlowWriter: %v", err)
	}
	t.Cleanup(func() { tgtRegs.Close() })

	srcRegs, err := fabrics.RegionsForFlowReader(srcReader)
	if err != nil {
		t.Fatalf("RegionsForFlowReader: %v", err)
	}
	t.Cleanup(func() { srcRegs.Close() })

	target, err := tgtFab.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	t.Cleanup(func() { target.Close() })

	targetPort := freePort(t)
	initiatorPort := freePort(t)

	info, err := target.Setup(fabrics.TargetConfig{
		Endpoint: fabrics.EndpointAddress{Node: "127.0.0.1", Service: targetPort},
		Provider: fabrics.ProviderTCP,
		Regions:  tgtRegs,
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
		Endpoint: fabrics.EndpointAddress{Node: "127.0.0.1", Service: initiatorPort},
		Provider: fabrics.ProviderTCP,
		Regions:  srcRegs,
	}); err != nil {
		t.Fatalf("Initiator.Setup: %v", err)
	}
	if err := initiator.AddTarget(info); err != nil {
		t.Fatalf("AddTarget: %v", err)
	}

	// Both sides must be progressed before the initiator reports the
	// connection up; the target advances its handshake state from inside
	// its non-blocking read.
	connectDeadline := time.Now().Add(10 * time.Second)
	for {
		if time.Now().After(connectDeadline) {
			t.Fatalf("initiator did not reach connected state within 10s")
		}
		_, rerr := target.ReadGrainNonBlocking()
		if rerr != nil && !errors.Is(rerr, fabrics.ErrNotReady) {
			t.Fatalf("target ReadGrainNonBlocking during setup: %v", rerr)
		}
		err := initiator.MakeProgressNonBlocking()
		if err == nil {
			break
		}
		if !errors.Is(err, fabrics.ErrNotReady) {
			t.Fatalf("initiator MakeProgressNonBlocking: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	cfg := srcWriter.Config()
	idx := mxl.CurrentIndex(cfg.Common.GrainRate)
	ga, err := srcWriter.OpenGrain(idx)
	if err != nil {
		t.Fatalf("OpenGrain: %v", err)
	}
	for i := range ga.Payload {
		ga.Payload[i] = byte((uint64(i) + idx) & 0xFF)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := initiator.TransferGrain(idx, 0, ga.TotalSlices); err != nil {
		t.Fatalf("TransferGrain: %v", err)
	}

	// Drive both sides until the grain arrives or the deadline expires.
	deadline := time.Now().Add(5 * time.Second)
	var (
		gotIdx uint64
		wg     sync.WaitGroup
		stop   = make(chan struct{})
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			err := initiator.MakeProgressNonBlocking()
			if err != nil && !errors.Is(err, fabrics.ErrNotReady) {
				t.Errorf("initiator MakeProgressNonBlocking: %v", err)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	for {
		if time.Now().After(deadline) {
			close(stop)
			wg.Wait()
			t.Fatalf("grain did not arrive at target within 5s")
		}
		idxRecv, err := target.ReadGrain(100 * time.Millisecond)
		if err == nil {
			gotIdx = idxRecv
			break
		}
		if !errors.Is(err, fabrics.ErrNotReady) {
			close(stop)
			wg.Wait()
			t.Fatalf("Target.ReadGrain: %v", err)
		}
	}
	close(stop)
	wg.Wait()

	if gotIdx != idx {
		t.Fatalf("target reported idx=%d, want %d", gotIdx, idx)
	}

	// The fabric writes both grain header and payload directly into the
	// target's flow memory. Opening the same index on the target writer
	// surfaces the populated bytes; Cancel leaves the head untouched so
	// the assertion doesn't perturb downstream state.
	ga, err2 := tgtWriter.OpenGrain(idx)
	if err2 != nil {
		t.Fatalf("tgt OpenGrain: %v", err2)
	}
	for i := 0; i < len(ga.Payload); i += 4096 {
		want := byte((uint64(i) + idx) & 0xFF)
		if ga.Payload[i] != want {
			t.Fatalf("payload mismatch at offset %d: got %d, want %d", i, ga.Payload[i], want)
		}
	}
	if err := ga.Cancel(); err != nil {
		t.Fatalf("tgt Cancel: %v", err)
	}
}
