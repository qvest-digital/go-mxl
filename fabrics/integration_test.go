//go:build mxl_integration

package fabrics_test

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

const audioFlowJSON = `{
  "description": "go-mxl fabrics test audio, 48k stereo",
  "id": "b3bb5be7-9fe9-4324-a5bb-4c70e1084449",
  "format": "urn:x-nmos:format:audio",
  "label": "go-mxl fabrics test audio",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl fabrics test:Audio"] },
  "parents": [],
  "media_type": "audio/float32",
  "sample_rate": { "numerator": 48000 },
  "channel_count": 2,
  "bit_depth": 32
}`

const audioFlowID = "b3bb5be7-9fe9-4324-a5bb-4c70e1084449"

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
		Writer:   tgtWriter,
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
		Reader:   srcReader,
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

// TestFabricsSampleTransferTCP exercises the continuous (audio) data path over
// libfabric's TCP provider on loopback, mirroring TestFabricsGrainTransferTCP.
// A writer fills a batch of samples, the initiator transfers it to the target,
// and the bytes land in the receiving writer's flow memory.
func TestFabricsSampleTransferTCP(t *testing.T) {
	srcInst := newDomain(t)
	tgtInst := newDomain(t)

	srcWriter, _, err := srcInst.NewWriter(audioFlowJSON)
	if err != nil {
		t.Fatalf("src NewWriter: %v", err)
	}
	t.Cleanup(func() { srcWriter.Close() })

	srcReader, err := srcInst.NewReader(audioFlowID)
	if err != nil {
		t.Fatalf("src NewReader: %v", err)
	}
	t.Cleanup(func() { srcReader.Close() })

	tgtWriter, _, err := tgtInst.NewWriter(audioFlowJSON)
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

	targetPort := freePort(t)
	initiatorPort := freePort(t)

	info, err := target.Setup(fabrics.TargetConfig{
		Endpoint: fabrics.EndpointAddress{Node: "127.0.0.1", Service: targetPort},
		Provider: fabrics.ProviderTCP,
		Writer:   tgtWriter,
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
		Reader:   srcReader,
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
		_, _, rerr := target.ReadSamplesNonBlocking()
		if rerr != nil && !errors.Is(rerr, fabrics.ErrNotReady) {
			t.Fatalf("target ReadSamplesNonBlocking during setup: %v", rerr)
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

	// Fill one batch of samples on the source writer with a per-channel byte
	// ramp keyed to absolute position so the receiving side can verify it.
	const batch = 480 // 10 ms at 48 kHz
	rate := srcWriter.Config().Common.GrainRate
	idx := mxl.CurrentIndex(rate)
	sa, err := srcWriter.OpenSamples(idx, batch)
	if err != nil {
		t.Fatalf("OpenSamples: %v", err)
	}
	for ch := uint64(0); ch < sa.ChannelCount; ch++ {
		f1, f2, err := sa.ChannelFragments(ch)
		if err != nil {
			t.Fatalf("ChannelFragments(%d): %v", ch, err)
		}
		var pos uint64
		for _, frag := range [][]byte{f1, f2} {
			for i := range frag {
				frag[i] = byte((pos + idx + ch) & 0xFF)
				pos++
			}
		}
	}
	if err := sa.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := initiator.TransferSamples(idx, batch); err != nil {
		t.Fatalf("TransferSamples: %v", err)
	}

	// Drive both sides until the samples arrive at the target.
	deadline := time.Now().Add(5 * time.Second)
	var (
		gotHead  uint64
		gotCount int
		wg       sync.WaitGroup
		stop     = make(chan struct{})
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
			t.Fatalf("samples did not arrive at target within 5s")
		}
		head, count, err := target.ReadSamples(100 * time.Millisecond)
		if err == nil {
			gotHead = head
			gotCount = count
			break
		}
		if !errors.Is(err, fabrics.ErrNotReady) {
			close(stop)
			wg.Wait()
			t.Fatalf("Target.ReadSamples: %v", err)
		}
	}
	close(stop)
	wg.Wait()

	if gotHead != idx {
		t.Fatalf("target reported head=%d, want %d", gotHead, idx)
	}
	if gotCount != batch {
		t.Fatalf("target reported count=%d, want %d", gotCount, batch)
	}

	// The fabric writes the samples directly into the target's flow memory.
	// Opening the same range on the target writer surfaces the populated
	// bytes; Cancel leaves the head untouched.
	sa2, err := tgtWriter.OpenSamples(idx, batch)
	if err != nil {
		t.Fatalf("tgt OpenSamples: %v", err)
	}
	for ch := uint64(0); ch < sa2.ChannelCount; ch++ {
		f1, f2, err := sa2.ChannelFragments(ch)
		if err != nil {
			t.Fatalf("tgt ChannelFragments(%d): %v", ch, err)
		}
		var pos uint64
		for _, frag := range [][]byte{f1, f2} {
			for i := range frag {
				want := byte((pos + idx + ch) & 0xFF)
				if frag[i] != want {
					t.Fatalf("ch%d sample byte at pos %d: got %d, want %d", ch, pos, frag[i], want)
				}
				pos++
			}
		}
	}
	if err := sa2.Cancel(); err != nil {
		t.Fatalf("tgt Cancel: %v", err)
	}
}

// concFlowJSON is a tiny (192x108) video flow for the concurrency test so many
// writers fit in /dev/shm. The setup race exercised here is in the fabric layer
// (provider init + endpoint id) and is independent of flow size.
const concFlowJSON = `{
  "description": "go-mxl fabrics concurrency test, 192x108",
  "id": "ffffffff-1b0f-417d-9059-8b94a47197ed",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl fabrics concurrency test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl fabrics conc test:Video"] },
  "parents": [],
  "media_type": "video/v210",
  "grain_rate": { "numerator": 30000, "denominator": 1001 },
  "frame_width": 192,
  "frame_height": 108,
  "interlace_mode": "progressive",
  "colorspace": "BT709",
  "components": [
    { "name": "Y",  "width": 192, "height": 108, "bit_depth": 10 },
    { "name": "Cb", "width": 96,  "height": 108, "bit_depth": 10 },
    { "name": "Cr", "width": 96,  "height": 108, "bit_depth": 10 }
  ]
}`

const concFlowID = "ffffffff-1b0f-417d-9059-8b94a47197ed"

// TestFabricsConcurrentTargetSetupTCP is a regression test for the
// concurrent-setup wedge (qvest-digital/mxl-dmf-terraform#226). Many targets set
// up at the same time in one process must ALL succeed. Before serializing setup,
// concurrent libfabric provider init (the first fi_getinfo) and a weakly-seeded
// endpoint-id RNG could make a setup fail or two endpoints collide, leaving one
// mirror stuck forever -- an intermittent failure that scaled with the number of
// concurrent setups (never with 1, ~1/3 of the time with ~9). With setup
// serialized (fabricSetupMu) all N succeed deterministically. Run on main
// (without the mutex) to observe the intermittent failure this guards against.
func TestFabricsConcurrentTargetSetupTCP(t *testing.T) {
	const (
		perRound = 16 // concurrent setups per round
		rounds   = 10 // repeat so the intermittent setup race is exposed reliably
	)
	inst := newDomain(t)
	fab, err := fabrics.NewInstance(inst)
	if err != nil {
		t.Fatalf("fabrics.NewInstance: %v", err)
	}
	t.Cleanup(func() { fab.Close() })

	for round := 0; round < rounds; round++ {
		concurrentSetupRound(t, inst, fab, perRound, round)
	}
}

// concurrentSetupRound fires perRound target setups from a single release point
// and fails if any does not succeed. Writers, targets and infos are released
// before it returns so /dev/shm and ephemeral ports are freed between rounds.
func concurrentSetupRound(t *testing.T, inst *mxl.Instance, fab *fabrics.Instance, perRound, round int) {
	t.Helper()

	writers := make([]*mxl.Writer, perRound)
	ports := make([]string, perRound)
	for i := 0; i < perRound; i++ {
		id := fmt.Sprintf("%08d-0000-0000-0000-%012d", round, i)
		w, _, err := inst.NewWriter(strings.ReplaceAll(concFlowJSON, concFlowID, id))
		if err != nil {
			t.Fatalf("round %d NewWriter[%d]: %v", round, i, err)
		}
		writers[i] = w
		ports[i] = freePort(t)
	}
	defer func() {
		for _, w := range writers {
			if w != nil {
				w.Close()
			}
		}
	}()

	// Fire every setup from a single release point to maximize overlap.
	targets := make([]*fabrics.Target, perRound)
	infos := make([]*fabrics.TargetInfo, perRound)
	errs := make([]error, perRound)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < perRound; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			tg, err := fab.NewTarget()
			if err != nil {
				errs[i] = err
				return
			}
			targets[i] = tg
			infos[i], errs[i] = tg.Setup(fabrics.TargetConfig{
				Endpoint: fabrics.EndpointAddress{Node: "127.0.0.1", Service: ports[i]},
				Provider: fabrics.ProviderTCP,
				Writer:   writers[i],
			})
		}(i)
	}
	close(start)
	wg.Wait()
	defer func() {
		for _, info := range infos {
			if info != nil {
				info.Close()
			}
		}
		for _, tg := range targets {
			if tg != nil {
				tg.Close()
			}
		}
	}()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("round %d: concurrent target[%d] setup failed: %v", round, i, err)
		}
	}
}
