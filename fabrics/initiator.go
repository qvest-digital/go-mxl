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

// InitiatorConfig configures the local end of a libmxl-fabrics
// initiator — the sender of grain transfers to one or more targets.
type InitiatorConfig struct {
	// Endpoint is the bind address for the local initiator.
	Endpoint EndpointAddress

	// Provider selects the libmxl-fabrics provider. Must match the
	// provider used by the targets the initiator will connect to.
	Provider Provider

	// Regions describes the local shared-memory backing of the flow
	// being sent. Typically obtained from RegionsForFlowReader.
	Regions *Regions

	// DeviceSupport requests support for transfers involving device
	// memory.
	DeviceSupport bool
}

// Initiator is a libmxl-fabrics sender. One initiator fans out grain
// writes to every target added via AddTarget.
type Initiator struct {
	mu     sync.Mutex
	parent *Instance
	handle C.mxlFabricsInitiator
}

// NewInitiator creates an Initiator on the given fabrics Instance.
// The Initiator is not active until Setup is called.
func (i *Instance) NewInitiator() (*Initiator, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return nil, ErrClosed()
	}
	var h C.mxlFabricsInitiator
	if err := fabricsStatusErr(C.mxlFabricsCreateInitiator(i.handle, &h)); err != nil {
		return nil, err
	}
	in := &Initiator{parent: i, handle: h}
	runtime.SetFinalizer(in, func(x *Initiator) { _ = x.Close() })
	return in, nil
}

// Setup binds the endpoint and registers the memory region.
func (in *Initiator) Setup(cfg InitiatorConfig) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return ErrClosed()
	}
	if cfg.Regions == nil {
		return ErrInvalidArg()
	}

	cfg.Regions.mu.Lock()
	defer cfg.Regions.mu.Unlock()
	if cfg.Regions.handle == nil {
		return ErrClosed()
	}

	cbuf := cfg.Endpoint.toC()
	defer cbuf.free()

	cInitCfg := C.mxlFabricsInitiatorConfig{
		endpointAddress: cbuf.addr,
		provider:        C.mxlFabricsProvider(cfg.Provider),
		regions:         cfg.Regions.handle,
		deviceSupport:   C.bool(cfg.DeviceSupport),
	}

	rc := C.mxlFabricsInitiatorSetup(in.handle, &cInitCfg)
	return fabricsStatusErr(rc)
}

// AddTarget registers a remote target as a destination for subsequent
// TransferGrain calls. Connection establishment is deferred to
// MakeProgress.
func (in *Initiator) AddTarget(info *TargetInfo) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return ErrClosed()
	}
	if info == nil {
		return ErrInvalidArg()
	}
	info.mu.Lock()
	defer info.mu.Unlock()
	if info.handle == nil {
		return ErrClosed()
	}
	return fabricsStatusErr(C.mxlFabricsInitiatorAddTarget(in.handle, info.handle))
}

// RemoveTarget de-registers a remote target. New transfers will skip
// it; pending transfers complete or fail on the next MakeProgress.
func (in *Initiator) RemoveTarget(info *TargetInfo) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return ErrClosed()
	}
	if info == nil {
		return ErrInvalidArg()
	}
	info.mu.Lock()
	defer info.mu.Unlock()
	if info.handle == nil {
		return ErrClosed()
	}
	return fabricsStatusErr(C.mxlFabricsInitiatorRemoveTarget(in.handle, info.handle))
}

// TransferGrain enqueues a transfer of the slice range
// [startSlice, endSlice) of grain at grainIdx to every added target.
// The call is non-blocking; completion is driven by MakeProgress.
func (in *Initiator) TransferGrain(grainIdx uint64, startSlice, endSlice uint16) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return ErrClosed()
	}
	return fabricsStatusErr(C.mxlFabricsInitiatorTransferGrain(
		in.handle, C.uint64_t(grainIdx),
		C.uint16_t(startSlice), C.uint16_t(endSlice),
	))
}

// MakeProgress drives queued transfers, connection setup, and
// connection shutdown for up to timeout. Returns ErrNotReady if there
// is still progress to be made when the timeout elapses.
func (in *Initiator) MakeProgress(timeout time.Duration) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return ErrClosed()
	}
	return fabricsStatusErr(C.mxlFabricsInitiatorMakeProgressBlocking(
		in.handle, C.uint16_t(timeoutMs(timeout))))
}

// MakeProgressNonBlocking is the non-blocking variant of MakeProgress.
// Returns ErrNotReady if there is still progress to be made.
func (in *Initiator) MakeProgressNonBlocking() error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return ErrClosed()
	}
	return fabricsStatusErr(C.mxlFabricsInitiatorMakeProgressNonBlocking(in.handle))
}

// Close destroys the Initiator. Safe to call multiple times.
func (in *Initiator) Close() error {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.handle == nil {
		return nil
	}
	in.parent.mu.RLock()
	defer in.parent.mu.RUnlock()
	var rc C.mxlStatus = C.MXL_STATUS_OK
	if in.parent.handle != nil {
		rc = C.mxlFabricsDestroyInitiator(in.parent.handle, in.handle)
	}
	in.handle = nil
	in.parent = nil
	runtime.SetFinalizer(in, nil)
	return fabricsStatusErr(rc)
}
