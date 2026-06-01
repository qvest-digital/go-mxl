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
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qvest-digital/go-mxl/mxl"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		log.Fatal(err)
	}
}

// defaultBatch returns the per-batch sample count: override when > 0, else
// ~10 ms of audio at rate, clamped to at least one sample.
func defaultBatch(rate mxl.Rational, override int) uint64 {
	batch := uint64(override)
	if batch == 0 {
		batch = uint64(rate.Num / (100 * rate.Den))
		if batch == 0 {
			batch = 1
		}
	}
	return batch
}

// fillSampleRamp writes a continuous byte ramp starting at start across f1
// then f2 (the two ring-buffer fragments of one channel) and returns the
// next ramp value. A matching reader recomputes the same sequence.
func fillSampleRamp(f1, f2 []byte, start uint64) uint64 {
	s := start
	for i := range f1 {
		f1[i] = byte(s & 0xFF)
		s++
	}
	for i := range f2 {
		f2[i] = byte(s & 0xFF)
		s++
	}
	return s
}

func run(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("write-samples", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		domain    = fs.String("domain", "/dev/shm/mxl", "MXL domain directory (tmpfs)")
		flowDef   = fs.String("flow-def", "", "Path to JSON flow definition")
		batchSize = fs.Int("batch", 0, "Samples per batch (0 = ~10 ms)")
		count     = fs.Int64("count", 0, "Stop after N samples (0 = run forever)")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *flowDef == "" {
		return errors.New("missing -flow-def <path>")
	}
	def, err := os.ReadFile(*flowDef)
	if err != nil {
		return fmt.Errorf("read %s: %w", *flowDef, err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		return fmt.Errorf("NewInstance: %w", err)
	}
	defer inst.Close()

	w, created, err := inst.NewWriter(string(def))
	if err != nil {
		return fmt.Errorf("NewWriter: %w", err)
	}
	defer w.Close()
	if !created {
		log.Printf("reusing existing flow")
	}

	cfg := w.Config()
	if cfg.Common.Format.IsDiscrete() {
		return fmt.Errorf("flow format %s is discrete; use write-grain instead", cfg.Common.Format)
	}

	rate := cfg.Common.GrainRate
	batch := defaultBatch(rate, *batchSize)
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
			return nil
		default:
		}

		sa, err := w.OpenSamples(idx, int(batch))
		if err != nil {
			return fmt.Errorf("OpenSamples(%d, %d): %w", idx, batch, err)
		}
		// Fill every channel with a continuous byte ramp.
		for ch := uint64(0); ch < sa.ChannelCount; ch++ {
			f1, f2, err := sa.ChannelFragments(ch)
			if err != nil {
				return fmt.Errorf("ChannelFragments(%d): %w", ch, err)
			}
			fillSampleRamp(f1, f2, written)
		}
		if err := sa.Commit(); err != nil {
			return fmt.Errorf("Commit: %w", err)
		}

		written += batch
		if *count > 0 && written >= uint64(*count) {
			log.Printf("done: %d samples written", written)
			return nil
		}

		idx += batch
		mxl.SleepNs(mxl.NsUntilIndex(idx, rate))
	}
}
