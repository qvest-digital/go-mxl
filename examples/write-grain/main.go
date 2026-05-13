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
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/qvest-digital/go-mxl/mxl"
)

func main() {
	var (
		domain  = flag.String("domain", "/dev/shm/mxl", "MXL domain directory (tmpfs)")
		flowDef = flag.String("flow-def", "", "Path to JSON flow definition")
		count   = flag.Int64("count", 0, "Stop after N grains (0 = run forever)")
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
	if !cfg.Common.Format.IsDiscrete() {
		log.Fatalf("flow format %s is not discrete; use write-samples instead", cfg.Common.Format)
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
			return
		default:
		}

		ga, err := w.OpenGrain(idx)
		if err != nil {
			log.Fatalf("OpenGrain(%d): %v", idx, err)
		}
		for i := range ga.Payload {
			ga.Payload[i] = byte((uint64(i) + idx) & 0xFF)
		}
		if err := ga.Commit(ga.TotalSlices, 0); err != nil {
			log.Fatalf("Commit(%d): %v", idx, err)
		}

		written++
		if *count > 0 && written >= *count {
			log.Printf("done: %d grains written", written)
			return
		}

		idx++
		// Pace ourselves to roughly the grain rate.
		mxl.SleepNs(mxl.NsUntilIndex(idx, rate))
	}
}
