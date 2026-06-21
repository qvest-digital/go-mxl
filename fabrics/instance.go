package fabrics

/*
#include <mxl/mxl.h>
#include <mxl/fabrics.h>
*/
import "C"

import (
	"runtime"
	"sync"

	"github.com/qvest-digital/go-mxl/mxl"
)

// fabricSetupMu serializes libmxl-fabrics target/initiator Setup across the
// whole process. The underlying libmxl-fabrics setup is not safe to run
// concurrently: it draws a per-endpoint id from a freshly, weakly-seeded RNG
// (under a concurrent setup burst two endpoints can draw the same id, so an
// initiator's id-keyed dispatch map collapses them and one mirror never
// connects), and the first fi_getinfo triggers libfabric's lazy, process-global
// provider init. Setup is not on the steady-state hot path, so serializing it is
// free in normal operation. Root fix: dmf-mxl/mxl#574; this is the immediate
// mitigation that needs no libmxl rebuild.
var fabricSetupMu sync.Mutex

// Instance is a fabrics-level handle bound to an mxl.Instance. Targets
// and initiators created from this instance can access the flows in
// the parent MXL domain.
type Instance struct {
	mu     sync.RWMutex
	parent *mxl.Instance // pinned to keep the underlying MXL instance alive
	handle C.mxlFabricsInstance
}

// NewInstance wraps the given mxl.Instance with a libmxl-fabrics
// instance. The mxl.Instance is pinned for the lifetime of the
// fabrics instance.
func NewInstance(in *mxl.Instance) (*Instance, error) {
	if in == nil {
		return nil, mxl.ErrInvalidArg
	}
	parentHandle := in.Handle()
	if parentHandle == nil {
		return nil, mxl.ErrClosed
	}

	var fh C.mxlFabricsInstance
	rc := C.mxlFabricsCreateInstance(C.mxlInstance(parentHandle), nil, &fh)
	if err := fabricsStatusErr(rc); err != nil {
		return nil, err
	}

	inst := &Instance{parent: in, handle: fh}
	runtime.SetFinalizer(inst, func(i *Instance) { _ = i.Close() })
	return inst, nil
}

// Close releases the underlying libmxl-fabrics instance. Targets and
// initiators created from this instance must be closed first. Safe to
// call multiple times.
func (i *Instance) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.handle == nil {
		return nil
	}
	rc := C.mxlFabricsDestroyInstance(i.handle)
	i.handle = nil
	i.parent = nil
	runtime.SetFinalizer(i, nil)
	return fabricsStatusErr(rc)
}

// rawHandle returns the underlying C handle. Caller must hold i.mu.
func (i *Instance) rawHandle() C.mxlFabricsInstance {
	return i.handle
}
