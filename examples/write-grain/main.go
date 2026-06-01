// write-grain creates (or opens) a discrete MXL flow from a JSON flow
// definition file and writes a deterministic synthetic pattern into each
// grain at the flow's nominal grain rate.
//
// Usage:
//
//	write-grain -domain /dev/shm/mxl -flow-def path/to/flow.json
//
// The grain payload is filled with `(byte_index + grain_index) % 256` so a
// matching reader can verify continuity. Stops cleanly on SIGINT/SIGTERM.
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

// fillGrainRamp writes byte((i+start)&0xFF) across buf. A matching reader
// uses the same pattern to verify grain continuity.
func fillGrainRamp(buf []byte, start uint64) {
	for i := range buf {
		buf[i] = byte((uint64(i) + start) & 0xFF)
	}
}

func run(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("write-grain", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		domain  = fs.String("domain", "/dev/shm/mxl", "MXL domain directory (tmpfs)")
		flowDef = fs.String("flow-def", "", "Path to JSON flow definition")
		count   = fs.Int64("count", 0, "Stop after N grains (0 = run forever)")
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
	if !cfg.Common.Format.IsDiscrete() {
		return fmt.Errorf("flow format %s is not discrete; use write-samples instead", cfg.Common.Format)
	}
	rate := cfg.Common.GrainRate
	idx := mxl.CurrentIndex(rate)
	log.Printf("writing flow grainRate=%d/%d starting at idx=%d", rate.Num, rate.Den, idx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var written int64
	for {
		select {
		case <-stop:
			log.Printf("stopping after %d grains", written)
			return nil
		default:
		}

		ga, err := w.OpenGrain(idx)
		if err != nil {
			return fmt.Errorf("OpenGrain(%d): %w", idx, err)
		}
		fillGrainRamp(ga.Payload, idx)
		if err := ga.Commit(ga.TotalSlices, 0); err != nil {
			return fmt.Errorf("Commit(%d): %w", idx, err)
		}

		written++
		if *count > 0 && written >= *count {
			log.Printf("done: %d grains written", written)
			return nil
		}

		idx++
		// Pace ourselves to roughly the grain rate.
		mxl.SleepNs(mxl.NsUntilIndex(idx, rate))
	}
}
