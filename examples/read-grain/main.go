// read-grain subscribes to a discrete MXL flow and prints a summary of
// each grain it observes.
//
// Usage:
//
//	read-grain -domain /dev/shm/mxl -flow <uuid>
//
// A writer must be producing into the same domain before this runs.
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
		timeout = flag.Duration("timeout", 200*time.Millisecond, "Per-grain read timeout")
		count   = flag.Int("count", 0, "Stop after N grains (0 = run forever)")
	)
	flag.Parse()

	if *flowID == "" {
		log.Fatalf("missing -flow <uuid>")
	}

	ver, err := mxl.LibVersion()
	if err == nil {
		log.Printf("libmxl %d.%d.%d (%s)", ver.Major, ver.Minor, ver.Bugfix, ver.Full)
	}

	ok, _ := mxl.IsTmpFs(*domain)
	if !ok {
		log.Printf("warning: %s is not on tmpfs; MXL expects a RAM-backed mount", *domain)
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
	log.Printf("flow format=%s grainRate=%d/%d grainCount=%d sliceSizes=%v",
		info.Config.Common.Format,
		info.Config.Common.GrainRate.Num, info.Config.Common.GrainRate.Den,
		info.Config.Discrete.GrainCount,
		info.Config.Discrete.SliceSizes)

	if !info.Config.Common.Format.IsDiscrete() {
		log.Fatalf("flow %s is not a discrete flow (samples not yet supported)", *flowID)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	rate := info.Config.Common.GrainRate
	idx := mxl.CurrentIndex(rate)
	if idx == mxl.UndefinedIndex {
		log.Fatalf("invalid edit rate")
	}

	seen := 0
	for {
		select {
		case <-stop:
			log.Printf("stopping: %d grains read", seen)
			return
		default:
		}

		g, err := r.GetGrain(idx, *timeout)
		switch {
		case err == nil:
			fmt.Printf("idx=%d size=%d slices=%d/%d flags=0x%x invalid=%v\n",
				g.Index, g.GrainSize, g.ValidSlices, g.TotalSlices, g.Flags, g.Invalid())
			idx++
			seen++
			if *count > 0 && seen >= *count {
				log.Printf("done: %d grains read", seen)
				return
			}
		case errors.Is(err, mxl.ErrTimeout):
			// Writer is slow or stopped; resync to current.
			idx = mxl.CurrentIndex(rate)
		case errors.Is(err, mxl.ErrOutOfRangeEarly):
			// We got ahead of the writer; back off briefly.
			time.Sleep(10 * time.Millisecond)
		case errors.Is(err, mxl.ErrOutOfRangeLate):
			// We fell behind; jump to current.
			log.Printf("fell behind, resyncing")
			idx = mxl.CurrentIndex(rate)
		default:
			log.Fatalf("GetGrain: %v", err)
		}
	}
}
