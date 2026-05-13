// sync-group demonstrates SyncGroup by waiting for data on a set of flows
// in lock-step and then reading the per-flow grain or sample range that has
// become available at the timestamp it waited for.
//
// Usage:
//
//	sync-group -domain /dev/shm/mxl -rate 30000/1001 -flow <uuid> -flow <uuid> ...
//
// Pass -flow once per flow you want to synchronize on. -rate is the grain
// rate used to convert "now" into a target timestamp.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/qvest-digital/go-mxl/mxl"
)

type flowList []string

func (f *flowList) String() string     { return strings.Join(*f, ",") }
func (f *flowList) Set(v string) error { *f = append(*f, v); return nil }

func parseRate(s string) (mxl.Rational, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return mxl.Rational{}, fmt.Errorf("rate %q must be N/D", s)
	}
	num, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return mxl.Rational{}, err
	}
	den, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return mxl.Rational{}, err
	}
	return mxl.Rational{Num: num, Den: den}, nil
}

func main() {
	var (
		domain  = flag.String("domain", "/dev/shm/mxl", "MXL domain directory (tmpfs)")
		rateStr = flag.String("rate", "30000/1001", "Grain/sample rate as N/D")
		timeout = flag.Duration("timeout", 500*time.Millisecond, "Per-tick wait timeout")
		count   = flag.Int64("count", 0, "Stop after N successful ticks (0 = forever)")
	)
	var flows flowList
	flag.Var(&flows, "flow", "Flow UUID to synchronize on (repeat)")
	flag.Parse()

	if len(flows) == 0 {
		log.Fatalf("at least one -flow <uuid> is required")
	}
	rate, err := parseRate(*rateStr)
	if err != nil {
		log.Fatalf("parse -rate: %v", err)
	}

	inst, err := mxl.NewInstance(*domain, "")
	if err != nil {
		log.Fatalf("NewInstance: %v", err)
	}
	defer inst.Close()

	group, err := inst.NewSyncGroup()
	if err != nil {
		log.Fatalf("NewSyncGroup: %v", err)
	}
	defer group.Close()

	readers := make([]*mxl.Reader, 0, len(flows))
	for _, id := range flows {
		r, err := inst.NewReader(id)
		if err != nil {
			log.Fatalf("NewReader(%s): %v", id, err)
		}
		defer r.Close()
		if err := group.AddReader(r); err != nil {
			log.Fatalf("AddReader(%s): %v", id, err)
		}
		readers = append(readers, r)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	idx := mxl.CurrentIndex(rate)
	var ticks int64
	for {
		select {
		case <-stop:
			log.Printf("stopping after %d ticks", ticks)
			return
		default:
		}

		ts := mxl.IndexToTimestamp(rate, idx)
		err := group.WaitForDataAt(ts, *timeout)
		switch {
		case err == nil:
			fmt.Printf("tick idx=%d ts=%d readers=%d\n", idx, ts, len(readers))
			ticks++
			if *count > 0 && ticks >= *count {
				log.Printf("done: %d ticks", ticks)
				return
			}
			idx++
		case errors.Is(err, mxl.ErrTimeout), errors.Is(err, mxl.ErrOutOfRangeEarly):
			// At least one reader isn't at ts yet. Brief sleep and retry —
			// re-syncing to CurrentIndex would skip ahead unnecessarily.
			time.Sleep(5 * time.Millisecond)
		case errors.Is(err, mxl.ErrOutOfRangeLate):
			log.Printf("fell behind at idx=%d, resyncing", idx)
			idx = mxl.CurrentIndex(rate)
		default:
			log.Fatalf("WaitForDataAt: %v", err)
		}
	}
}
