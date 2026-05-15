package fabrics

/*
#include <stdint.h>
#include <mxl/fabrics.h>
*/
import "C"

import (
	"runtime"
	"sync"
	"time"
)

// TargetConfig configures the local end of a libmxl-fabrics target —
// the receiver of grain transfers from one or more initiators.
type TargetConfig struct {
	// Endpoint is the bind address for the local target.
	Endpoint EndpointAddress

	// Provider selects the libmxl-fabrics provider.
	Provider Provider

	// Regions describes the local shared-memory backing of the flow
	// being received. Typically obtained from RegionsForFlowWriter.
	Regions *Regions

	// DeviceSupport requests support for transfers involving device
	// memory (e.g. GPU-direct). Most callers leave this false.
	DeviceSupport bool
}

// Target is a libmxl-fabrics receiver. One target accepts grain
// writes from any number of initiators that have called AddTarget on
// the matching TargetInfo.
type Target struct {
	mu     sync.Mutex
	parent *Instance // pinned for lifetime
	handle C.mxlFabricsTarget
}

// NewTarget creates a Target on the given fabrics Instance. The
// Target is not active until Setup is called.
func (i *Instance) NewTarget() (*Target, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return nil, ErrClosed()
	}
	var h C.mxlFabricsTarget
	if err := fabricsStatusErr(C.mxlFabricsCreateTarget(i.handle, &h)); err != nil {
		return nil, err
	}
	t := &Target{parent: i, handle: h}
	runtime.SetFinalizer(t, func(x *Target) { _ = x.Close() })
	return t, nil
}

// Setup binds the endpoint, registers the memory region, and returns
// the TargetInfo descriptor that should be transported to remote
// initiators. The returned TargetInfo is owned by the caller and
// must be Close'd separately.
func (t *Target) Setup(cfg TargetConfig) (*TargetInfo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return nil, ErrClosed()
	}
	if cfg.Regions == nil {
		return nil, ErrInvalidArg()
	}

	cfg.Regions.mu.Lock()
	defer cfg.Regions.mu.Unlock()
	if cfg.Regions.handle == nil {
		return nil, ErrClosed()
	}

	cbuf := cfg.Endpoint.toC()
	defer cbuf.free()

	cTargetCfg := C.mxlFabricsTargetConfig{
		endpointAddress: cbuf.addr,
		provider:        C.mxlFabricsProvider(cfg.Provider),
		regions:         cfg.Regions.handle,
		deviceSupport:   C.bool(cfg.DeviceSupport),
	}

	var info C.mxlFabricsTargetInfo
	if err := fabricsStatusErr(C.mxlFabricsTargetSetup(t.handle, &cTargetCfg, &info)); err != nil {
		return nil, err
	}
	ti := &TargetInfo{handle: info}
	runtime.SetFinalizer(ti, func(x *TargetInfo) { _ = x.Close() })
	return ti, nil
}

// ReadGrain blocks until a grain has been written by an initiator (or
// timeout elapses) and returns its index. Returns ErrNotReady if no
// grain arrived before the timeout.
func (t *Target) ReadGrain(timeout time.Duration) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return 0, ErrClosed()
	}
	var idx C.uint64_t
	rc := C.mxlFabricsTargetReadGrain(t.handle, C.uint16_t(timeoutMs(timeout)), &idx)
	if err := fabricsStatusErr(rc); err != nil {
		return 0, err
	}
	return uint64(idx), nil
}

// ReadGrainNonBlocking is the non-blocking variant of ReadGrain.
// Returns ErrNotReady if no grain is currently available.
func (t *Target) ReadGrainNonBlocking() (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return 0, ErrClosed()
	}
	var idx C.uint64_t
	rc := C.mxlFabricsTargetReadGrainNonBlocking(t.handle, &idx)
	if err := fabricsStatusErr(rc); err != nil {
		return 0, err
	}
	return uint64(idx), nil
}

// Close destroys the Target. Safe to call multiple times.
func (t *Target) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return nil
	}
	t.parent.mu.RLock()
	defer t.parent.mu.RUnlock()
	var rc C.mxlStatus = C.MXL_STATUS_OK
	if t.parent.handle != nil {
		rc = C.mxlFabricsDestroyTarget(t.parent.handle, t.handle)
	}
	t.handle = nil
	t.parent = nil
	runtime.SetFinalizer(t, nil)
	return fabricsStatusErr(rc)
}
