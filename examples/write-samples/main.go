// write-samples creates (or opens) a continuous (audio) MXL flow and writes
// a deterministic synthetic pattern into batches of samples at the flow's
// nominal sample rate.
//
// Usage:
//
//	write-samples -domain /dev/shm/mxl -flow-def path/to/audio-flow.json
//
// Each sample byte is set to `(running_sample_index % 256)` so a matching
// reader can verify continuity. Batch size defaults to ~10 ms of audio.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qvest-digital/go-mxl/mxl"
)

func main() {
	var (
		domain    = flag.String("domain", "/dev/shm/mxl", "MXL domain directory (tmpfs)")
		flowDef   = flag.String("flow-def", "", "Path to JSON flow definition")
		batchSize = flag.Int("batch", 0, "Samples per batch (0 = ~10 ms)")
		count     = flag.Int64("count", 0, "Stop after N samples (0 = run forever)")
	)
	flag.Parse()

	if *flowDef == "" {
		log.Fatalf("missing -flow-def <path>")
	}
	def, err := os.ReadFile(*flowDef)
	if err != nil {
		log.Fatalf("read %s: %v", *flowDef, err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		log.Fatalf("NewInstance: %v", err)
	}
	defer inst.Close()

	w, created, err := inst.NewWriter(string(def))
	if err != nil {
		log.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()
	if !created {
		log.Printf("reusing existing flow")
	}

	cfg := w.Config()
	if cfg.Common.Format.IsDiscrete() {
		log.Fatalf("flow format %s is discrete; use write-grain instead", cfg.Common.Format)
	}

	rate := cfg.Common.GrainRate
	batch := uint64(*batchSize)
	if batch == 0 {
		// ~10 ms worth of samples.
		batch = uint64(rate.Num / (100 * rate.Den))
		if batch == 0 {
			batch = 1
		}
	}
	idx := mxl.CurrentIndex(rate)
	log.Printf("writing flow sampleRate=%d/%d channels=%d batch=%d first-idx=%d",
		rate.Num, rate.Den, cfg.Continuous.ChannelCount, batch, idx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var written uint64
	for {
		select {
		case <-stop:
			log.Printf("stopping after %d samples", written)
			return
		default:
		}

		sa, err := w.OpenSamples(idx, int(batch))
		if err != nil {
			log.Fatalf("OpenSamples(%d, %d): %v", idx, batch, err)
		}
		// Fill every channel with a continuous byte ramp.
		samp := written
		for ch := uint64(0); ch < sa.ChannelCount; ch++ {
			f1, f2, err := sa.ChannelFragments(ch)
			if err != nil {
				log.Fatalf("ChannelFragments(%d): %v", ch, err)
			}
			s := samp
			for i := range f1 {
				f1[i] = byte(s & 0xFF)
				s++
			}
			for i := range f2 {
				f2[i] = byte(s & 0xFF)
				s++
			}
		}
		if err := sa.Commit(); err != nil {
			log.Fatalf("Commit: %v", err)
		}

		written += batch
		if *count > 0 && written >= uint64(*count) {
			log.Printf("done: %d samples written", written)
			return
		}

		idx += batch
		mxl.SleepNs(mxl.NsUntilIndex(idx, rate))
	}
}
