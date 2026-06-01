package fabrics

/*
#include <stdint.h>
#include <stdlib.h>
#include <mxl/fabrics.h>
*/
import "C"

import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/qvest-digital/go-mxl/mxl"
)

// TargetConfig configures the local end of a libmxl-fabrics target —
// the receiver of grain transfers from one or more initiators.
type TargetConfig struct {
	// Endpoint is the bind address for the local target.
	Endpoint EndpointAddress

	// Provider selects the libmxl-fabrics provider.
	Provider Provider

	// Writer is the flow writer whose backing memory receives incoming
	// transfers. The Target pins it for the Target lifetime after Setup.
	Writer *mxl.Writer

	// Options is a JSON-formatted options string passed through to
	// libmxl-fabrics. Leave empty for the default options.
	Options string
}

// Target is a libmxl-fabrics receiver. One target accepts grain
// writes from any number of initiators that have called AddTarget on
// the matching TargetInfo.
type Target struct {
	mu     sync.Mutex
	parent *Instance // pinned for lifetime
	writer *mxl.Writer
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

// Setup binds the endpoint, registers the writer's memory, and returns
// the TargetInfo descriptor that should be transported to remote
// initiators. The returned TargetInfo is owned by the caller and
// must be Close'd separately.
func (t *Target) Setup(cfg TargetConfig) (*TargetInfo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return nil, ErrClosed()
	}
	if cfg.Writer == nil {
		return nil, ErrInvalidArg()
	}
	h := cfg.Writer.Handle()
	if h == nil {
		return nil, ErrClosed()
	}

	cbuf := cfg.Endpoint.toC()
	defer cbuf.free()
	var copts *C.char
	if cfg.Options != "" {
		copts = C.CString(cfg.Options)
		defer C.free(unsafe.Pointer(copts))
	}

	cTargetCfg := C.mxlFabricsTargetConfig{
		version:         C.MXL_FABRICS_API_VERSION,
		endpointAddress: cbuf.addr,
		provider:        C.mxlFabricsProvider(cfg.Provider),
		writer:          C.mxlFlowWriter(h),
	}

	var info C.mxlFabricsTargetInfo
	if err := fabricsStatusErr(C.mxlFabricsTargetSetup(t.handle, &cTargetCfg, copts, &info)); err != nil {
		return nil, err
	}
	t.writer = cfg.Writer
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

// ReadSamples blocks until samples have been written by an initiator (or
// timeout elapses) and returns the head index of the received range and the
// per-channel sample count. Returns ErrNotReady if no samples arrived before
// the timeout. Continuous (audio) flows only.
func (t *Target) ReadSamples(timeout time.Duration) (headIndex uint64, count int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return 0, 0, ErrClosed()
	}
	var head C.uint64_t
	var n C.size_t
	rc := C.mxlFabricsTargetReadSamples(t.handle, C.uint16_t(timeoutMs(timeout)), &head, &n)
	if err := fabricsStatusErr(rc); err != nil {
		return 0, 0, err
	}
	return uint64(head), int(n), nil
}

// ReadSamplesNonBlocking is the non-blocking variant of ReadSamples.
// Returns ErrNotReady if no samples are currently available.
func (t *Target) ReadSamplesNonBlocking() (headIndex uint64, count int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return 0, 0, ErrClosed()
	}
	var head C.uint64_t
	var n C.size_t
	rc := C.mxlFabricsTargetReadSamplesNonBlocking(t.handle, &head, &n)
	if err := fabricsStatusErr(rc); err != nil {
		return 0, 0, err
	}
	return uint64(head), int(n), nil
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
	t.writer = nil
	runtime.SetFinalizer(t, nil)
	return fabricsStatusErr(rc)
}
