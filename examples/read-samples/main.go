// read-samples subscribes to a continuous (audio) MXL flow and prints a
// summary of each sample batch it observes. Mirror of read-grain for the
// continuous side of the API.
//
// Usage:
//
//	read-samples -domain /dev/shm/mxl -flow <uuid>
//
// Run write-samples (or any other producer) in another shell first.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qvest-digital/go-mxl"
)

func main() {
	var (
		domain  = flag.String("domain", "/dev/shm/mxl", "MXL domain directory (tmpfs)")
		flowID  = flag.String("flow", "", "Flow UUID to read")
		// Keep the timeout small relative to the ring buffer's live window
		// (bufferLength/2 - count). For 48 kHz with a 480-sample batch and
		// a 19456-sample buffer the window is ~190 ms; if GetSamples waits
		// longer than that, head can race past our requested index and
		// the post-wait re-evaluation returns ErrOutOfRangeLate instead
		// of OK. ~2× the batch period is a safe choice.
		timeout = flag.Duration("timeout", 20*time.Millisecond, "Per-batch read timeout")
		count   = flag.Int64("count", 0, "Stop after N batches (0 = run forever)")
		batch   = flag.Int("batch", 0, "Samples per batch (0 = ~10 ms)")
	)
	flag.Parse()

	if *flowID == "" {
		log.Fatalf("missing -flow <uuid>")
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		log.Fatalf("NewInstance: %v", err)
	}
	defer inst.Close()

	r, err := inst.NewReader(*flowID)
	if err != nil {
		log.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	info, err := r.Info()
	if err != nil {
		log.Fatalf("Info: %v", err)
	}
	if info.Config.Common.Format.IsDiscrete() {
		log.Fatalf("flow %s is discrete; use read-grain", *flowID)
	}
	cfg := info.Config
	rate := cfg.Common.GrainRate
	bs := uint64(*batch)
	if bs == 0 {
		bs = uint64(rate.Num / (100 * rate.Den))
		if bs == 0 {
			bs = 1
		}
	}
	max, _ := r.GetMaxReadLengthSamples()
	log.Printf("flow channels=%d bufferLen=%d sampleRate=%d/%d batch=%d max=%d",
		cfg.Continuous.ChannelCount, cfg.Continuous.BufferLength,
		rate.Num, rate.Den, bs, max)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Anchor to the writer's current head rather than wall-clock time. The
	// audio ring buffer is short (~400 ms here); using CurrentIndex(rate)
	// can place the first read just ahead of head, then the blocking wait
	// inside GetSamples lets head race past us by enough to push our range
	// off the back of the ring — yielding TOO_LATE in a livelock loop.
	idx := info.Runtime.HeadIndex
	if idx == 0 {
		log.Fatalf("flow has no head yet (no producer?)")
	}

	var seen int64
	for {
		select {
		case <-stop:
			log.Printf("stopping after %d batches", seen)
			return
		default:
		}

		v, err := r.GetSamples(idx, int(bs), *timeout)
		switch {
		case err == nil:
			f1, f2, _ := v.ChannelFragments(0)
			fmt.Printf("idx=%d ch=%d frags=[%d,%d]\n", idx, v.ChannelCount, len(f1), len(f2))
			idx += bs
			seen++
			if *count > 0 && seen >= *count {
				log.Printf("done: %d batches", seen)
				return
			}
		case errors.Is(err, mxl.ErrOutOfRangeEarly):
			// Writer hasn't reached idx yet; the blocking GetSamples already
			// waited up to -timeout. Brief sleep then retry the same idx.
			time.Sleep(10 * time.Millisecond)
		case errors.Is(err, mxl.ErrOutOfRangeLate):
			// We've fallen out of the buffer's live window. Snap back to
			// the current head to resume.
			rt, rerr := r.Runtime()
			if rerr != nil {
				log.Fatalf("Runtime: %v", rerr)
			}
			log.Printf("fell behind (idx=%d head=%d), resyncing", idx, rt.HeadIndex)
			idx = rt.HeadIndex
		default:
			log.Fatalf("GetSamples: %v", err)
		}
	}
}
