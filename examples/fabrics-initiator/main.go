// fabrics-initiator transfers grains from a local MXL flow reader over a
// libmxl-fabrics fabric to a matching target. The target's serialized
// TargetInfo (produced by fabrics-target) addresses the remote end.
//
// Usage:
//
//	fabrics-initiator -domain /dev/shm/mxl-src -flow <uuid> \
//	    -provider tcp -node 127.0.0.1 -service 23457 \
//	    -target-info target.info
//
// A separate writer (e.g. write-grain) must be producing into the source
// flow before this runs. The same flow definition must be in use at the
// target end.
package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qvest-digital/go-mxl/fabrics"
	"github.com/qvest-digital/go-mxl/mxl"
)

func main() {
	var (
		domain     = flag.String("domain", "/dev/shm/mxl-src", "MXL domain directory (tmpfs)")
		flowID     = flag.String("flow", "", "Flow UUID to forward")
		providerS  = flag.String("provider", "tcp", "libmxl-fabrics provider (tcp|verbs|efa|shm)")
		node       = flag.String("node", "127.0.0.1", "Local endpoint address")
		service    = flag.String("service", "23457", "Local endpoint port/service")
		targetFile = flag.String("target-info", "target.info", "Path to read the TargetInfo from")
		count      = flag.Int64("count", 0, "Stop after N transfers (0 = run forever)")
		poll       = flag.Duration("poll", 50*time.Millisecond, "MakeProgress interval")
	)
	flag.Parse()

	if *flowID == "" {
		log.Fatalf("missing -flow <uuid>")
	}
	provider, err := fabrics.ParseProvider(*providerS)
	if err != nil {
		log.Fatalf("parse -provider %q: %v", *providerS, err)
	}
	infoBytes, err := os.ReadFile(*targetFile)
	if err != nil {
		log.Fatalf("read %s: %v", *targetFile, err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		log.Fatalf("mxl.NewInstance: %v", err)
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
	if !info.Config.Common.Format.IsDiscrete() {
		log.Fatalf("flow %s is not discrete; fabrics-initiator only forwards discrete flows", *flowID)
	}

	fi, err := fabrics.NewInstance(inst)
	if err != nil {
		log.Fatalf("fabrics.NewInstance: %v", err)
	}
	defer fi.Close()

	regs, err := fabrics.RegionsForFlowReader(r)
	if err != nil {
		log.Fatalf("RegionsForFlowReader: %v", err)
	}
	defer regs.Close()

	ti, err := fabrics.ParseTargetInfo(string(infoBytes))
	if err != nil {
		log.Fatalf("ParseTargetInfo: %v", err)
	}
	defer ti.Close()

	in, err := fi.NewInitiator()
	if err != nil {
		log.Fatalf("NewInitiator: %v", err)
	}
	defer in.Close()

	if err := in.Setup(fabrics.InitiatorConfig{
		Endpoint: fabrics.EndpointAddress{Node: *node, Service: *service},
		Provider: provider,
		Regions:  regs,
	}); err != nil {
		log.Fatalf("Initiator.Setup: %v", err)
	}
	if err := in.AddTarget(ti); err != nil {
		log.Fatalf("AddTarget: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	rate := info.Config.Common.GrainRate
	idx := mxl.CurrentIndex(rate)
	var sent int64
	for {
		select {
		case <-stop:
			log.Printf("stopping after %d transfers", sent)
			return
		default:
		}

		// Wait until the source grain at idx is committed by the upstream writer.
		g, err := r.GetGrain(idx, *poll)
		switch {
		case err == nil:
			if err := in.TransferGrain(g.Index, 0, g.TotalSlices); err != nil {
				log.Fatalf("TransferGrain(%d): %v", g.Index, err)
			}
			if err := in.MakeProgress(*poll); err != nil && !errors.Is(err, fabrics.ErrNotReady) {
				log.Fatalf("MakeProgress: %v", err)
			}
			sent++
			if *count > 0 && sent >= *count {
				log.Printf("done: %d transfers", sent)
				return
			}
			idx++
		case errors.Is(err, mxl.ErrTimeout), errors.Is(err, mxl.ErrOutOfRangeEarly):
			// Source not caught up yet; drive progress and retry.
			if err := in.MakeProgressNonBlocking(); err != nil && !errors.Is(err, fabrics.ErrNotReady) {
				log.Fatalf("MakeProgressNonBlocking: %v", err)
			}
		case errors.Is(err, mxl.ErrOutOfRangeLate):
			log.Printf("fell behind at idx=%d, resyncing", idx)
			idx = mxl.CurrentIndex(rate)
		default:
			log.Fatalf("GetGrain: %v", err)
		}
	}
}
