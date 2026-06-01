// fabrics-target wraps an MXL flow writer with a libmxl-fabrics target so
// that grains transferred over the fabric by a matching initiator land in
// the local flow and are consumable by read-grain on the same domain.
//
// Usage:
//
//	fabrics-target -domain /dev/shm/mxl-tgt -flow-def flow.json \
//	    -provider tcp -node 127.0.0.1 -service 23456 \
//	    -target-info target.info
//
// The target binds the given endpoint, then writes its serialized TargetInfo
// to -target-info for the initiator side to read. The same flow.json must
// be used at both ends so the ring buffer layout matches.
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
	fs := flag.NewFlagSet("fabrics-target", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		domain     = fs.String("domain", "/dev/shm/mxl-tgt", "MXL domain directory (tmpfs)")
		flowDef    = fs.String("flow-def", "", "Path to JSON flow definition")
		providerS  = fs.String("provider", "tcp", "libmxl-fabrics provider (tcp|verbs|efa|shm)")
		node       = fs.String("node", "127.0.0.1", "Endpoint address (bind)")
		service    = fs.String("service", "23456", "Endpoint port/service")
		targetFile = fs.String("target-info", "target.info", "Path to write the serialized TargetInfo")
		timeout    = fs.Duration("timeout", 500*time.Millisecond, "Per-poll wait timeout")
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
	provider, err := fabrics.ParseProvider(*providerS)
	if err != nil {
		return fmt.Errorf("parse -provider %q: %w", *providerS, err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		return fmt.Errorf("mxl.NewInstance: %w", err)
	}
	defer inst.Close()

	w, _, err := inst.NewWriter(string(def))
	if err != nil {
		return fmt.Errorf("NewWriter: %w", err)
	}
	defer w.Close()

	fi, err := fabrics.NewInstance(inst)
	if err != nil {
		return fmt.Errorf("fabrics.NewInstance: %w", err)
	}
	defer fi.Close()

	tgt, err := fi.NewTarget()
	if err != nil {
		return fmt.Errorf("NewTarget: %w", err)
	}
	defer tgt.Close()

	info, err := tgt.Setup(fabrics.TargetConfig{
		Endpoint: fabrics.EndpointAddress{Node: *node, Service: *service},
		Provider: provider,
		Writer:   w,
	})
	if err != nil {
		return fmt.Errorf("Target.Setup: %w", err)
	}
	defer info.Close()

	s, err := info.MarshalString()
	if err != nil {
		return fmt.Errorf("TargetInfo.MarshalString: %w", err)
	}
	if err := os.WriteFile(*targetFile, []byte(s), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", *targetFile, err)
	}
	log.Printf("target bound %s:%s, wrote info to %s (%d bytes)", *node, *service, *targetFile, len(s))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			log.Printf("stopping")
			return nil
		default:
		}
		idx, err := tgt.ReadGrain(*timeout)
		switch {
		case err == nil:
			log.Printf("received grain idx=%d", idx)
		case errors.Is(err, fabrics.ErrNotReady):
			// Idle tick; loop.
		default:
			return fmt.Errorf("ReadGrain: %w", err)
		}
	}
}
