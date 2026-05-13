package mxl

/*
#include <stdlib.h>
#include <stdint.h>
#include <mxl/mxl.h>
#include <mxl/flow.h>
#include <mxl/flowinfo.h>
*/
import "C"

import (
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// Reader subscribes to a flow and exposes grain-level reads for discrete
// (video/data) flows. A single Reader is not safe for concurrent use; create
// one per goroutine.
//
// The reader pins its parent Instance against garbage collection so that
// callers don't have to keep both references alive.
type Reader struct {
	mu     sync.Mutex
	parent *Instance // pinned to keep the instance alive until Close
	handle C.mxlFlowReader
}

// NewReader creates a flow reader. options is currently unused by libmxl and
// may be empty.
func (i *Instance) NewReader(flowID string) (*Reader, error) {
	return i.NewReaderOpts(flowID, "")
}

// NewReaderOpts creates a flow reader with options (currently unused).
func (i *Instance) NewReaderOpts(flowID, options string) (*Reader, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return nil, ErrClosed
	}
	cid := C.CString(flowID)
	defer C.free(unsafe.Pointer(cid))

	var copts *C.char
	if options != "" {
		copts = C.CString(options)
		defer C.free(unsafe.Pointer(copts))
	}

	var h C.mxlFlowReader
	if err := statusErr(C.mxlCreateFlowReader(i.handle, cid, copts, &h)); err != nil {
		return nil, err
	}

	r := &Reader{parent: i, handle: h}
	runtime.SetFinalizer(r, func(r *Reader) { _ = r.Close() })
	return r, nil
}

// Close releases the underlying flow reader. Safe to call multiple times.
// After Close, the parent Instance becomes eligible for GC again (assuming
// no other references).
func (r *Reader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return nil
	}
	r.parent.mu.RLock()
	defer r.parent.mu.RUnlock()
	var rc C.mxlStatus = C.MXL_STATUS_OK
	if r.parent.handle != nil {
		rc = C.mxlReleaseFlowReader(r.parent.handle, r.handle)
	}
	r.handle = nil
	r.parent = nil
	runtime.SetFinalizer(r, nil)
	return statusErr(rc)
}

// Info returns a snapshot of the flow's config + runtime metadata.
func (r *Reader) Info() (FlowInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return FlowInfo{}, ErrClosed
	}
	var fi C.mxlFlowInfo
	if err := statusErr(C.mxlFlowReaderGetInfo(r.handle, &fi)); err != nil {
		return FlowInfo{}, err
	}
	return goFlowInfo(&fi), nil
}

// Config returns just the immutable configuration of the flow.
func (r *Reader) Config() (FlowConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return FlowConfig{}, ErrClosed
	}
	var c C.mxlFlowConfigInfo
	if err := statusErr(C.mxlFlowReaderGetConfigInfo(r.handle, &c)); err != nil {
		return FlowConfig{}, err
	}
	return goFlowConfig(&c), nil
}

// Runtime returns just the mutable runtime info (head index, last write/read).
func (r *Reader) Runtime() (FlowRuntime, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return FlowRuntime{}, ErrClosed
	}
	var rt C.mxlFlowRuntimeInfo
	if err := statusErr(C.mxlFlowReaderGetRuntimeInfo(r.handle, &rt)); err != nil {
		return FlowRuntime{}, err
	}
	return goFlowRuntime(&rt), nil
}

// GetGrain blocks until the full grain at index is available or timeout
// elapses. Use timeout=0 for a true non-blocking attempt; for the
// fully-non-blocking variant use GetGrainNonBlocking which avoids the
// futex wait setup entirely.
//
// The returned Grain.Payload aliases shared memory and is only valid until
// the next read on this Reader. Copy() it if you need to retain it.
func (r *Reader) GetGrain(index uint64, timeout time.Duration) (Grain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return Grain{}, ErrClosed
	}
	var info C.mxlGrainInfo
	var payload *C.uint8_t
	rc := C.mxlFlowReaderGetGrain(r.handle, C.uint64_t(index),
		C.uint64_t(durationNs(timeout)), &info, &payload)
	if err := statusErr(rc); err != nil {
		return Grain{}, err
	}
	return makeGrain(&info, payload), nil
}

// GetGrainSlice blocks for a grain that has at least minValidSlices committed
// slices, or returns the current partial state if it does. Use
// GrainValidSlicesAny to accept any number (i.e. return as soon as the grain
// exists at all), or GrainValidSlicesAll for a fully committed grain.
func (r *Reader) GetGrainSlice(index uint64, minValidSlices uint16, timeout time.Duration) (Grain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return Grain{}, ErrClosed
	}
	var info C.mxlGrainInfo
	var payload *C.uint8_t
	rc := C.mxlFlowReaderGetGrainSlice(r.handle, C.uint64_t(index),
		C.uint16_t(minValidSlices), C.uint64_t(durationNs(timeout)),
		&info, &payload)
	if err := statusErr(rc); err != nil {
		return Grain{}, err
	}
	return makeGrain(&info, payload), nil
}

// GetGrainNonBlocking returns immediately. Most likely error for "not ready
// yet" is ErrOutOfRangeEarly.
func (r *Reader) GetGrainNonBlocking(index uint64) (Grain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return Grain{}, ErrClosed
	}
	var info C.mxlGrainInfo
	var payload *C.uint8_t
	rc := C.mxlFlowReaderGetGrainNonBlocking(r.handle, C.uint64_t(index), &info, &payload)
	if err := statusErr(rc); err != nil {
		return Grain{}, err
	}
	return makeGrain(&info, payload), nil
}

// GetGrainSliceNonBlocking is the non-blocking variant of GetGrainSlice.
func (r *Reader) GetGrainSliceNonBlocking(index uint64, minValidSlices uint16) (Grain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return Grain{}, ErrClosed
	}
	var info C.mxlGrainInfo
	var payload *C.uint8_t
	rc := C.mxlFlowReaderGetGrainSliceNonBlocking(r.handle, C.uint64_t(index),
		C.uint16_t(minValidSlices), &info, &payload)
	if err := statusErr(rc); err != nil {
		return Grain{}, err
	}
	return makeGrain(&info, payload), nil
}

// makeGrain copies scalar metadata and aliases the payload into a Go slice.
func makeGrain(info *C.mxlGrainInfo, payload *C.uint8_t) Grain {
	size := int(info.grainSize)
	var data []byte
	if payload != nil && size > 0 {
		// Alias C-owned shared memory. Caller must not retain past next read.
		data = unsafe.Slice((*byte)(unsafe.Pointer(payload)), size)
	}
	return Grain{
		Index:       uint64(info.index),
		Flags:       uint32(info.flags),
		GrainSize:   uint32(info.grainSize),
		TotalSlices: uint16(info.totalSlices),
		ValidSlices: uint16(info.validSlices),
		Payload:     data,
	}
}

func durationNs(d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return d.Nanoseconds()
}
