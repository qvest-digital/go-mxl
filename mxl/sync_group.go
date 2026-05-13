package mxl

/*
#include <stdint.h>
#include <mxl/mxl.h>
#include <mxl/flow.h>
*/
import "C"

import (
	"runtime"
	"sync"
	"time"
)

// SyncGroup waits for data to become available on a set of flow readers in
// parallel. For continuous flows this means the sample at the specified
// timestamp; for discrete flows it means the corresponding grain (optionally
// only a minimum number of slices). All readers in a group must belong to
// the same Instance.
//
// Methods on SyncGroup are safe for concurrent use from multiple goroutines.
type SyncGroup struct {
	mu     sync.RWMutex
	parent *Instance // pinned to keep the instance alive until Close
	handle C.mxlFlowSynchronizationGroup
}

// NewSyncGroup creates an empty synchronization group on this instance.
func (i *Instance) NewSyncGroup() (*SyncGroup, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return nil, ErrClosed
	}
	var h C.mxlFlowSynchronizationGroup
	if err := statusErr(C.mxlCreateFlowSynchronizationGroup(i.handle, &h)); err != nil {
		return nil, err
	}
	g := &SyncGroup{parent: i, handle: h}
	runtime.SetFinalizer(g, func(g *SyncGroup) { _ = g.Close() })
	return g, nil
}

// Close releases the synchronization group. Safe to call multiple times.
func (g *SyncGroup) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.handle == nil {
		return nil
	}
	g.parent.mu.RLock()
	defer g.parent.mu.RUnlock()
	var rc C.mxlStatus = C.MXL_STATUS_OK
	if g.parent.handle != nil {
		rc = C.mxlReleaseFlowSynchronizationGroup(g.parent.handle, g.handle)
	}
	g.handle = nil
	g.parent = nil
	runtime.SetFinalizer(g, nil)
	return statusErr(rc)
}

// AddReader adds a flow reader to the group. For continuous flows, WaitForDataAt
// will wait for the sample at the requested timestamp; for discrete flows it
// will wait for the full grain. Adding the same reader twice updates its
// configuration in place (last call wins).
func (g *SyncGroup) AddReader(r *Reader) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.handle == nil {
		return ErrClosed
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlFlowSynchronizationGroupAddReader(g.handle, r.handle))
}

// AddPartialGrainReader adds a discrete flow reader configured to wait for at
// least minValidSlices slices of the grain at the requested timestamp.
func (g *SyncGroup) AddPartialGrainReader(r *Reader, minValidSlices uint16) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.handle == nil {
		return ErrClosed
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlFlowSynchronizationGroupAddPartialGrainReader(g.handle, r.handle,
		C.uint16_t(minValidSlices)))
}

// RemoveReader removes a reader from the group.
func (g *SyncGroup) RemoveReader(r *Reader) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.handle == nil {
		return ErrClosed
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlFlowSynchronizationGroupRemoveReader(g.handle, r.handle))
}

// WaitForDataAt blocks until the data corresponding to the given timestamp
// (ns since the SMPTE ST 2059 epoch) is available across every reader in
// the group, or timeout elapses.
func (g *SyncGroup) WaitForDataAt(timestamp uint64, timeout time.Duration) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlFlowSynchronizationGroupWaitForDataAt(g.handle,
		C.uint64_t(timestamp), C.uint64_t(durationNs(timeout))))
}
