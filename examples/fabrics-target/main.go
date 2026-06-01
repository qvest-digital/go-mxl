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
		domain     = flag.String("domain", "/dev/shm/mxl-tgt", "MXL domain directory (tmpfs)")
		flowDef    = flag.String("flow-def", "", "Path to JSON flow definition")
		providerS  = flag.String("provider", "tcp", "libmxl-fabrics provider (tcp|verbs|efa|shm)")
		node       = flag.String("node", "127.0.0.1", "Endpoint address (bind)")
		service    = flag.String("service", "23456", "Endpoint port/service")
		targetFile = flag.String("target-info", "target.info", "Path to write the serialized TargetInfo")
		timeout    = flag.Duration("timeout", 500*time.Millisecond, "Per-poll wait timeout")
	)
	flag.Parse()

	if *flowDef == "" {
		log.Fatalf("missing -flow-def <path>")
	}
	def, err := os.ReadFile(*flowDef)
	if err != nil {
		log.Fatalf("read %s: %v", *flowDef, err)
	}
	provider, err := fabrics.ParseProvider(*providerS)
	if err != nil {
		log.Fatalf("parse -provider %q: %v", *providerS, err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		log.Fatalf("mxl.NewInstance: %v", err)
	}
	defer inst.Close()

	w, _, err := inst.NewWriter(string(def))
	if err != nil {
		log.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	fi, err := fabrics.NewInstance(inst)
	if err != nil {
		log.Fatalf("fabrics.NewInstance: %v", err)
	}
	defer fi.Close()

	tgt, err := fi.NewTarget()
	if err != nil {
		log.Fatalf("NewTarget: %v", err)
	}
	defer tgt.Close()

	info, err := tgt.Setup(fabrics.TargetConfig{
		Endpoint: fabrics.EndpointAddress{Node: *node, Service: *service},
		Provider: provider,
		Writer:   w,
	})
	if err != nil {
		log.Fatalf("Target.Setup: %v", err)
	}
	defer info.Close()

	s, err := info.MarshalString()
	if err != nil {
		log.Fatalf("TargetInfo.MarshalString: %v", err)
	}
	if err := os.WriteFile(*targetFile, []byte(s), 0o644); err != nil {
		log.Fatalf("write %s: %v", *targetFile, err)
	}
	log.Printf("target bound %s:%s, wrote info to %s (%d bytes)", *node, *service, *targetFile, len(s))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			log.Printf("stopping")
			return
		default:
		}
		idx, err := tgt.ReadGrain(*timeout)
		switch {
		case err == nil:
			log.Printf("received grain idx=%d", idx)
		case errors.Is(err, fabrics.ErrNotReady):
			// Idle tick; loop.
		default:
			log.Fatalf("ReadGrain: %v", err)
		}
	}
}
