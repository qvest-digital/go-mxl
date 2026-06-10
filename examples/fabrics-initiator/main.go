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
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qvest-digital/go-mxl/fabrics"
	"github.com/qvest-digital/go-mxl/mxl"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		log.Fatal(err)
	}
}

func run(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("fabrics-initiator", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		domain     = fs.String("domain", "/dev/shm/mxl-src", "MXL domain directory (tmpfs)")
		flowID     = fs.String("flow", "", "Flow UUID to forward")
		providerS  = fs.String("provider", "tcp", "libmxl-fabrics provider (tcp|verbs|efa|shm)")
		node       = fs.String("node", "127.0.0.1", "Local endpoint address")
		service    = fs.String("service", "23457", "Local endpoint port/service")
		targetFile = fs.String("target-info", "target.info", "Path to read the TargetInfo from")
		count      = fs.Int64("count", 0, "Stop after N transfers (0 = run forever)")
		poll       = fs.Duration("poll", 50*time.Millisecond, "MakeProgress interval")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *flowID == "" {
		return errors.New("missing -flow <uuid>")
	}
	provider, err := fabrics.ParseProvider(*providerS)
	if err != nil {
		return fmt.Errorf("parse -provider %q: %w", *providerS, err)
	}
	infoBytes, err := os.ReadFile(*targetFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", *targetFile, err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		return fmt.Errorf("mxl.NewInstance: %w", err)
	}
	defer inst.Close()

	r, err := inst.NewReader(*flowID)
	if err != nil {
		return fmt.Errorf("NewReader: %w", err)
	}
	defer r.Close()

	info, err := r.Info()
	if err != nil {
		return fmt.Errorf("Info: %w", err)
	}
	if !info.Config.Common.Format.IsDiscrete() {
		return fmt.Errorf("flow %s is not discrete; fabrics-initiator only forwards discrete flows", *flowID)
	}

	fi, err := fabrics.NewInstance(inst)
	if err != nil {
		return fmt.Errorf("fabrics.NewInstance: %w", err)
	}
	defer fi.Close()

	ti, err := fabrics.ParseTargetInfo(string(infoBytes))
	if err != nil {
		return fmt.Errorf("ParseTargetInfo: %w", err)
	}
	defer ti.Close()

	in, err := fi.NewInitiator()
	if err != nil {
		return fmt.Errorf("NewInitiator: %w", err)
	}
	defer in.Close()

	if err := in.Setup(fabrics.InitiatorConfig{
		Interface: fabrics.InterfaceConfig{
			Provider: provider,
			Address:  fabrics.EndpointAddress{Node: *node, Service: *service},
		},
		Reader: r,
	}); err != nil {
		return fmt.Errorf("Initiator.Setup: %w", err)
	}
	if err := in.AddTarget(ti); err != nil {
		return fmt.Errorf("AddTarget: %w", err)
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
			return nil
		default:
		}

		// Wait until the source grain at idx is committed by the upstream writer.
		g, err := r.GetGrain(idx, *poll)
		switch {
		case err == nil:
			if err := in.TransferGrain(g.Index, 0, g.TotalSlices); err != nil {
				return fmt.Errorf("TransferGrain(%d): %w", g.Index, err)
			}
			if err := in.MakeProgress(*poll); err != nil && !errors.Is(err, fabrics.ErrNotReady) {
				return fmt.Errorf("MakeProgress: %w", err)
			}
			sent++
			if *count > 0 && sent >= *count {
				log.Printf("done: %d transfers", sent)
				return nil
			}
			idx++
		case errors.Is(err, mxl.ErrTimeout), errors.Is(err, mxl.ErrOutOfRangeEarly):
			// Source not caught up yet; drive progress and retry.
			if err := in.MakeProgressNonBlocking(); err != nil && !errors.Is(err, fabrics.ErrNotReady) {
				return fmt.Errorf("MakeProgressNonBlocking: %w", err)
			}
		case errors.Is(err, mxl.ErrOutOfRangeLate):
			log.Printf("fell behind at idx=%d, resyncing", idx)
			idx = mxl.CurrentIndex(rate)
		default:
			return fmt.Errorf("GetGrain: %w", err)
		}
	}
}
