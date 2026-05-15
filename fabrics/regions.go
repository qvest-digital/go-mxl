package fabrics

/*
#include <mxl/flow.h>
#include <mxl/fabrics.h>
*/
import "C"

import (
	"runtime"
	"sync"

	"github.com/qvest-digital/go-mxl/mxl"
)

// Regions is a collection of memory regions that back a flow. Pass
// one to TargetConfig or InitiatorConfig to register the flow's
// shared memory with the chosen libmxl-fabrics provider.
//
// The Regions handle may be freed (Close) once the Target or
// Initiator has been set up.
type Regions struct {
	mu     sync.Mutex
	parent any // *mxl.Reader or *mxl.Writer, pinned for lifetime
	handle C.mxlFabricsRegions
}

// RegionsForFlowReader returns the regions describing the shared
// memory backing the given FlowReader. The Reader is pinned for the
// lifetime of the Regions.
func RegionsForFlowReader(r *mxl.Reader) (*Regions, error) {
	if r == nil {
		return nil, mxl.ErrInvalidArg
	}
	h := r.Handle()
	if h == nil {
		return nil, mxl.ErrClosed
	}
	var rh C.mxlFabricsRegions
	if err := fabricsStatusErr(C.mxlFabricsRegionsForFlowReader(C.mxlFlowReader(h), &rh)); err != nil {
		return nil, err
	}
	out := &Regions{parent: r, handle: rh}
	runtime.SetFinalizer(out, func(x *Regions) { _ = x.Close() })
	return out, nil
}

// RegionsForFlowWriter returns the regions describing the shared
// memory backing the given FlowWriter.
func RegionsForFlowWriter(w *mxl.Writer) (*Regions, error) {
	if w == nil {
		return nil, mxl.ErrInvalidArg
	}
	h := w.Handle()
	if h == nil {
		return nil, mxl.ErrClosed
	}
	var rh C.mxlFabricsRegions
	if err := fabricsStatusErr(C.mxlFabricsRegionsForFlowWriter(C.mxlFlowWriter(h), &rh)); err != nil {
		return nil, err
	}
	out := &Regions{parent: w, handle: rh}
	runtime.SetFinalizer(out, func(x *Regions) { _ = x.Close() })
	return out, nil
}

// Close frees the Regions handle.
func (r *Regions) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle == nil {
		return nil
	}
	rc := C.mxlFabricsRegionsFree(r.handle)
	r.handle = nil
	r.parent = nil
	runtime.SetFinalizer(r, nil)
	return fabricsStatusErr(rc)
}

// rawHandle returns the underlying C handle. Caller must hold r.mu.
func (r *Regions) rawHandle() C.mxlFabricsRegions {
	return r.handle
}
