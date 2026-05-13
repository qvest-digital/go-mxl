package mxl

/*
#include <stdint.h>
#include <mxl/mxl.h>
#include <mxl/flow.h>
#include <mxl/flowinfo.h>
#include "internal_helpers.h"
*/
import "C"

import (
	"time"
	"unsafe"
)

// SamplesView is an immutable, per-channel view onto a contiguous range of
// samples returned by Reader.GetSamples / GetSamplesNonBlocking. The view
// spans all channels of the flow and may straddle the ring-buffer
// wraparound; ChannelFragments exposes the (up to) two contiguous fragments
// per channel.
//
// The fragment byte slices alias libmxl's shared memory directly. They are
// only valid for as long as the underlying ring buffer entries remain
// undisturbed by the writer; libmxl makes no formal guarantee about how
// long this is. Callers that need to retain the data must copy it out.
type SamplesView struct {
	slice C.mxlWrappedMultiBufferSlice
	// ChannelCount is the number of channels in this view.
	ChannelCount uint64
	// Stride is the byte distance between the same offset in two adjacent
	// channels' ring buffers.
	Stride uint64
}

// ChannelFragments returns the (up to two) byte fragments that make up the
// requested sample range for channel `ch`. The second fragment is non-empty
// only when the range straddles the ring buffer wraparound. Slices alias
// shared memory; copy them out if you need to retain.
func (s *SamplesView) ChannelFragments(ch uint64) (frag1, frag2 []byte, err error) {
	if ch >= s.ChannelCount {
		return nil, nil, ErrInvalidArg
	}
	for i := 0; i < 2; i++ {
		basePtr := C.mxl_go_wrapped_fragment_ptr_ro(&s.slice, C.size_t(i))
		size := uint64(C.mxl_go_wrapped_fragment_size_ro(&s.slice, C.size_t(i)))
		if basePtr == nil || size == 0 {
			continue
		}
		chPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(ch)*uintptr(s.Stride))
		buf := unsafe.Slice((*byte)(chPtr), int(size))
		if i == 0 {
			frag1 = buf
		} else {
			frag2 = buf
		}
	}
	return frag1, frag2, nil
}

// CopyChannel returns a single heap-allocated slice with the entire range of
// samples for channel `ch`, concatenating both fragments. Use this when you
// need to retain the data past the next read or hand it off to another
// goroutine.
func (s *SamplesView) CopyChannel(ch uint64) ([]byte, error) {
	f1, f2, err := s.ChannelFragments(ch)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(f1)+len(f2))
	out = append(out, f1...)
	out = append(out, f2...)
	return out, nil
}

// GetMaxReadLengthSamples returns the absolute maximum number of samples a
// single GetSamples call may request on this reader. Only valid on
// continuous flows.
func (r *Reader) GetMaxReadLengthSamples() (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	if err := statusErr(C.mxlFlowReaderGetMaxReadLengthSamples(r.handle, &n)); err != nil {
		return 0, err
	}
	return uint64(n), nil
}

// GetSamples blocks until count samples ending at the given index are
// available (or timeout elapses) across all channels of a continuous flow,
// and returns a view onto them.
//
// Per the C API: this call never returns ErrTimeout; the actual failure
// mode when waiting for unavailable data is ErrOutOfRangeEarly.
func (r *Reader) GetSamples(index uint64, count int, timeout time.Duration) (*SamplesView, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return nil, ErrClosed
	}
	if count <= 0 {
		return nil, ErrInvalidArg
	}
	var slice C.mxlWrappedMultiBufferSlice
	rc := C.mxlFlowReaderGetSamples(r.handle, C.uint64_t(index), C.size_t(count),
		C.uint64_t(durationNs(timeout)), &slice)
	if err := statusErr(rc); err != nil {
		return nil, err
	}
	return makeSamplesView(&slice), nil
}

// GetSamplesNonBlocking is the non-blocking variant of GetSamples.
func (r *Reader) GetSamplesNonBlocking(index uint64, count int) (*SamplesView, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return nil, ErrClosed
	}
	if count <= 0 {
		return nil, ErrInvalidArg
	}
	var slice C.mxlWrappedMultiBufferSlice
	rc := C.mxlFlowReaderGetSamplesNonBlocking(r.handle, C.uint64_t(index), C.size_t(count), &slice)
	if err := statusErr(rc); err != nil {
		return nil, err
	}
	return makeSamplesView(&slice), nil
}

func makeSamplesView(slice *C.mxlWrappedMultiBufferSlice) *SamplesView {
	return &SamplesView{
		slice:        *slice,
		ChannelCount: uint64(C.mxl_go_wrapped_count_ro(slice)),
		Stride:       uint64(C.mxl_go_wrapped_stride_ro(slice)),
	}
}
