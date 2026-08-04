package mxl

/*
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>
#include <mxl/mxl.h>
#include <mxl/flow.h>
#include <mxl/flowinfo.h>
#include "internal_helpers.h"
*/
import "C"

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

// Writer is a flow producer. A Writer is created from an Instance via
// NewWriter and routes per-grain or per-sample work to its underlying
// mxlFlowWriter. Discrete flows (video, data) use the grain API
// (OpenGrain/Commit/Cancel); continuous flows (audio) use the sample API
// (OpenSamples/CommitSamples/CancelSamples). The cached Config tells callers
// which side of the API is valid for this Writer.
//
// A Writer is not safe for concurrent use; create one per goroutine.
type Writer struct {
	mu     sync.Mutex
	parent *Instance // pinned to keep the instance alive until Close
	handle C.mxlFlowWriter
	config FlowConfig

	// dataPath is <domain>/<id>.mxl-flow/data, the file libmxl deletes
	// the flow by. guard holds a shared lock on it; see flowGuard.
	// Both are empty/nil when the file could not be opened, which
	// leaves the writer behaving as it did before guards existed.
	dataPath string
	guard    *flowGuard
}

// NewWriter creates or opens a flow writer for the given JSON flow definition.
// If a flow with the matching id already exists in this domain, it is opened
// and 'created' is false; otherwise the flow is created and 'created' is true.
// In the existing-flow case the on-disk flow's config wins and may differ
// from flowDef; use w.Config() to inspect it.
func (i *Instance) NewWriter(flowDef string) (*Writer, bool, error) {
	return i.NewWriterOpts(flowDef, "")
}

// NewWriterOpts is the option-taking variant of NewWriter. options is
// currently unused by libmxl and may be empty.
func (i *Instance) NewWriterOpts(flowDef, options string) (*Writer, bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return nil, false, ErrClosed
	}

	cdef := C.CString(flowDef)
	defer C.free(unsafe.Pointer(cdef))

	var copts *C.char
	if options != "" {
		copts = C.CString(options)
		defer C.free(unsafe.Pointer(copts))
	}

	var h C.mxlFlowWriter
	var cfg C.mxlFlowConfigInfo
	var created C.bool
	rc := C.mxlCreateFlowWriter(i.handle, cdef, copts, &h, &cfg, &created)
	if err := statusErr(rc); err != nil {
		return nil, false, err
	}

	w := &Writer{
		parent: i,
		handle: h,
		config: goFlowConfig(&cfg),
	}
	// Taken now rather than at release time so the guard is pinned to
	// the file this writer opened. Acquiring it later would pin
	// whatever the path resolves to then, which is the wrong file in
	// precisely the case the guard exists for. A failure here is not
	// fatal: it costs Detach, not the writer.
	w.dataPath = flowDataPath(i.domain, uuidString(w.config.Common.ID))
	if g, err := newFlowGuard(w.dataPath); err == nil {
		w.guard = g
		i.addGuard(g)
	}
	runtime.SetFinalizer(w, func(w *Writer) { _ = w.Close() })
	return w, bool(created), nil
}

// Close releases the underlying flow writer. Any in-progress OpenGrain or
// OpenSamples is silently dropped by libmxl; callers should call Cancel or
// Commit first to keep grain state predictable. Safe to call multiple times.
//
// libmxl removes the flow from the domain when the writer being released
// is its last holder. Close keeps that behaviour, with one exception: it
// does not offer the flow for deletion when the flow directory no longer
// holds the file this writer opened. libmxl decides on the writer's own
// descriptor but deletes by path, so once the directory has been
// replaced the two name different flows and the delete would fall on
// whichever one holds the path now.
//
// Use Detach to release a writer without offering the flow at all.
func (w *Writer) Close() error {
	return w.release(w.guard != nil && !w.guard.pathStillOurs(w.dataPath))
}

// Detach releases the underlying flow writer and leaves the flow in the
// domain, whether or not this writer was its last holder. Use it where
// something other than this writer owns the flow's lifetime.
//
// Returns ErrNoFlowGuard without releasing anything when the writer has
// no guard on its data file, because then the release cannot be kept
// from deleting the flow. Safe to call multiple times.
func (w *Writer) Detach() error {
	w.mu.Lock()
	held := w.handle != nil
	missing := w.guard == nil
	w.mu.Unlock()
	if held && missing {
		return ErrNoFlowGuard
	}
	return w.release(true)
}

// release hands the writer back to libmxl. With keep set, the guard is
// still held when libmxl tries to upgrade the writer's descriptor to an
// exclusive lock, so the upgrade fails and the flow survives; without
// it, the guard is dropped first and libmxl deletes the flow if this
// writer was the last holder.
func (w *Writer) release(keep bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.handle == nil {
		return nil
	}
	inst := w.parent
	if !keep {
		inst.dropGuard(w.guard)
		w.guard = nil
	}

	inst.mu.RLock()
	var rc C.mxlStatus = C.MXL_STATUS_OK
	if inst.handle != nil {
		rc = C.mxlReleaseFlowWriter(inst.handle, w.handle)
	}
	inst.mu.RUnlock()

	// Dropped only now on the keeping path: the lock has to outlive
	// libmxl's attempt to upgrade it, which is what the release above
	// makes.
	inst.dropGuard(w.guard)
	w.guard = nil
	w.handle = nil
	w.parent = nil
	runtime.SetFinalizer(w, nil)
	return statusErr(rc)
}

// Config returns the immutable flow configuration captured at create time.
// This is the source of truth for whether to use grain or sample APIs.
func (w *Writer) Config() FlowConfig {
	return w.config
}

// Handle returns the underlying C handle as an opaque pointer for use
// by the sister fabrics sub-package. Callers must not retain the
// pointer past Close; the fabrics types pin the Writer.
func (w *Writer) Handle() unsafe.Pointer {
	return unsafe.Pointer(w.handle)
}

// GrainInfo returns the current libmxl-side metadata for the grain at the
// given index without opening it for mutation. Only valid on discrete flows.
func (w *Writer) GrainInfo(index uint64) (Grain, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.handle == nil {
		return Grain{}, ErrClosed
	}
	var info C.mxlGrainInfo
	if err := statusErr(C.mxlFlowWriterGetGrainInfo(w.handle, C.uint64_t(index), &info)); err != nil {
		return Grain{}, err
	}
	return Grain{
		Index:       uint64(info.index),
		Flags:       uint32(info.flags),
		GrainSize:   uint32(info.grainSize),
		TotalSlices: uint16(info.totalSlices),
		ValidSlices: uint16(info.validSlices),
	}, nil
}

// GrainWriteAccess is a mutable view onto a single grain opened by
// Writer.OpenGrain. Mutate Payload in place, then call Commit (passing the
// number of valid slices) or Cancel exactly once. After Commit/Cancel the
// Payload slice must not be touched.
//
// The Payload aliases the underlying shared memory directly; callers must
// not retain it past Commit/Cancel.
type GrainWriteAccess struct {
	writer  *Writer
	info    C.mxlGrainInfo
	payload *C.uint8_t
	closed  bool
	// Payload is the mutable grain payload buffer.
	Payload []byte
	// TotalSlices is the number of slices that make up a complete grain.
	TotalSlices uint16
	// GrainSize is the total payload size in bytes for a complete grain.
	GrainSize uint32
}

// OpenGrain opens the grain at the given index for mutation. libmxl tracks
// the currently-open grain on the writer, so OpenGrain must be matched with
// either Commit or Cancel before opening another grain.
func (w *Writer) OpenGrain(index uint64) (*GrainWriteAccess, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.handle == nil {
		return nil, ErrClosed
	}
	var info C.mxlGrainInfo
	var payload *C.uint8_t
	if err := statusErr(C.mxlFlowWriterOpenGrain(w.handle, C.uint64_t(index), &info, &payload)); err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, errors.New("mxl: mxlFlowWriterOpenGrain returned NULL payload")
	}
	size := int(info.grainSize)
	ga := &GrainWriteAccess{
		writer:      w,
		info:        info,
		payload:     payload,
		TotalSlices: uint16(info.totalSlices),
		GrainSize:   uint32(info.grainSize),
	}
	if size > 0 {
		ga.Payload = unsafe.Slice((*byte)(unsafe.Pointer(payload)), size)
	}
	return ga, nil
}

// Commit publishes the open grain. validSlices is the number of slices that
// have been written; pass TotalSlices for a complete grain. flags is the
// bitmask written into the on-disk grain header (use GrainFlagInvalid to
// signal an invalid grain).
func (g *GrainWriteAccess) Commit(validSlices uint16, flags uint32) error {
	if g.closed {
		return errors.New("mxl: GrainWriteAccess already committed or cancelled")
	}
	g.closed = true
	g.info.validSlices = C.uint16_t(validSlices)
	g.info.flags = C.uint32_t(flags)
	g.writer.mu.Lock()
	defer g.writer.mu.Unlock()
	if g.writer.handle == nil {
		return ErrClosed
	}
	rc := C.mxlFlowWriterCommitGrain(g.writer.handle, &g.info)
	// Defensively clear the Payload so further use blows up loudly.
	g.Payload = nil
	return statusErr(rc)
}

// Cancel discards the open grain without making it visible to readers.
func (g *GrainWriteAccess) Cancel() error {
	if g.closed {
		return nil
	}
	g.closed = true
	g.writer.mu.Lock()
	defer g.writer.mu.Unlock()
	if g.writer.handle == nil {
		return ErrClosed
	}
	rc := C.mxlFlowWriterCancelGrain(g.writer.handle)
	g.Payload = nil
	return statusErr(rc)
}

// GetMaxWriteLengthSamples returns the absolute maximum number of samples a
// single OpenSamples call may request on this writer. Only valid on
// continuous flows.
func (w *Writer) GetMaxWriteLengthSamples() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	if err := statusErr(C.mxlFlowWriterGetMaxWriteLengthSamples(w.handle, &n)); err != nil {
		return 0, err
	}
	return uint64(n), nil
}

// SamplesWriteAccess is a mutable, per-channel view onto a contiguous range
// of samples opened by Writer.OpenSamples. The view spans all channels of
// the flow and may straddle the ring-buffer wraparound; ChannelFragments
// exposes the (up to) two contiguous fragments per channel.
//
// Mutate the fragment slices in place, then call Commit or Cancel exactly
// once. The slices alias shared memory and must not be retained past then.
type SamplesWriteAccess struct {
	writer *Writer
	slice  C.mxlMutableWrappedMultiBufferSlice
	closed bool
	// ChannelCount is the number of channels in this view.
	ChannelCount uint64
	// Stride is the byte distance between the same offset in two adjacent
	// channels' ring buffers.
	Stride uint64
}

// OpenSamples opens count samples ending at the given index (i.e. samples
// (index-count+1) .. index) for mutation across all channels. Only valid on
// continuous flows.
func (w *Writer) OpenSamples(index uint64, count int) (*SamplesWriteAccess, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.handle == nil {
		return nil, ErrClosed
	}
	if count <= 0 {
		return nil, ErrInvalidArg
	}
	var slice C.mxlMutableWrappedMultiBufferSlice
	rc := C.mxlFlowWriterOpenSamples(w.handle, C.uint64_t(index), C.size_t(count), &slice)
	if err := statusErr(rc); err != nil {
		return nil, err
	}
	sa := &SamplesWriteAccess{
		writer:       w,
		slice:        slice,
		ChannelCount: uint64(C.mxl_go_wrapped_count_rw(&slice)),
		Stride:       uint64(C.mxl_go_wrapped_stride_rw(&slice)),
	}
	return sa, nil
}

// ChannelFragments returns the (up to two) mutable byte fragments that make
// up the opened sample range for channel `ch`. The second fragment is
// non-empty only when the open range straddles the ring buffer wraparound.
// Slices alias shared memory and are valid until Commit or Cancel.
func (s *SamplesWriteAccess) ChannelFragments(ch uint64) (frag1, frag2 []byte, err error) {
	if s.closed {
		return nil, nil, errors.New("mxl: SamplesWriteAccess already committed or cancelled")
	}
	if ch >= s.ChannelCount {
		return nil, nil, ErrInvalidArg
	}
	for i := 0; i < 2; i++ {
		basePtr := C.mxl_go_wrapped_fragment_ptr_rw(&s.slice, C.size_t(i))
		size := uint64(C.mxl_go_wrapped_fragment_size_rw(&s.slice, C.size_t(i)))
		if basePtr == nil || size == 0 {
			continue
		}
		// Offset into channel ch's ring buffer at this fragment's position.
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

// Commit publishes the opened sample range to readers.
func (s *SamplesWriteAccess) Commit() error {
	if s.closed {
		return errors.New("mxl: SamplesWriteAccess already committed or cancelled")
	}
	s.closed = true
	s.writer.mu.Lock()
	defer s.writer.mu.Unlock()
	if s.writer.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlFlowWriterCommitSamples(s.writer.handle))
}

// Cancel discards the opened sample range.
func (s *SamplesWriteAccess) Cancel() error {
	if s.closed {
		return nil
	}
	s.closed = true
	s.writer.mu.Lock()
	defer s.writer.mu.Unlock()
	if s.writer.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlFlowWriterCancelSamples(s.writer.handle))
}
