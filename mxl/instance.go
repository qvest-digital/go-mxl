package mxl

/*
#include <stdlib.h>
#include <stdbool.h>
#include <mxl/mxl.h>
#include <mxl/flow.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

// Version describes the loaded libmxl version.
type Version struct {
	Major  uint16
	Minor  uint16
	Bugfix uint16
	Build  uint16
	Full   string
}

// LibVersion returns the version of the linked libmxl SDK.
func LibVersion() (Version, error) {
	var v C.mxlVersionType
	if err := statusErr(C.mxlGetVersion(&v)); err != nil {
		return Version{}, err
	}
	return Version{
		Major:  uint16(v.major),
		Minor:  uint16(v.minor),
		Bugfix: uint16(v.bugfix),
		Build:  uint16(v.build),
		Full:   C.GoString(v.full),
	}, nil
}

// IsTmpFs reports whether the given path resides on a RAM-backed filesystem
// (tmpfs/ramfs on Linux). MXL domains must live on tmpfs for correctness.
func IsTmpFs(path string) (bool, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var out C.bool
	if err := statusErr(C.mxlIsTmpFs(cpath, &out)); err != nil {
		return false, err
	}
	return bool(out), nil
}

// Instance is a handle to an MXL domain. One instance maps to one tmpfs
// directory; create as many Readers and Writers from it as you need. Methods
// on Instance are safe for concurrent use from multiple goroutines.
type Instance struct {
	mu     sync.RWMutex
	handle C.mxlInstance // nil after Close
}

// NewInstance opens an MXL domain at the given filesystem path. The path must
// exist and (on Linux) should be a tmpfs mount; this is enforced by libmxl
// itself for some operations. options is currently unused by libmxl and may
// be empty.
func NewInstance(domain, options string) (*Instance, error) {
	cdomain := C.CString(domain)
	defer C.free(unsafe.Pointer(cdomain))

	var copts *C.char
	if options != "" {
		copts = C.CString(options)
		defer C.free(unsafe.Pointer(copts))
	}

	h := C.mxlCreateInstance(cdomain, copts)
	if h == nil {
		return nil, errors.New("mxl: mxlCreateInstance returned NULL")
	}

	inst := &Instance{handle: h}
	runtime.SetFinalizer(inst, func(i *Instance) { _ = i.Close() })
	return inst, nil
}

// Close releases the underlying MXL instance and any flow readers or writers
// still attached to it. Safe to call multiple times.
func (i *Instance) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.handle == nil {
		return nil
	}
	rc := C.mxlDestroyInstance(i.handle)
	i.handle = nil
	runtime.SetFinalizer(i, nil)
	return statusErr(rc)
}

// GarbageCollect removes stale flow files left behind by writers that crashed
// without releasing them. libmxl runs this automatically on Create; long-
// lived processes should call it periodically.
func (i *Instance) GarbageCollect() error {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return ErrClosed
	}
	return statusErr(C.mxlGarbageCollectFlows(i.handle))
}

// IsFlowActive reports whether the given flow has a live writer.
func (i *Instance) IsFlowActive(flowID string) (bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return false, ErrClosed
	}
	cid := C.CString(flowID)
	defer C.free(unsafe.Pointer(cid))
	var out C.bool
	if err := statusErr(C.mxlIsFlowActive(i.handle, cid, &out)); err != nil {
		return false, err
	}
	return bool(out), nil
}

// FlowDef returns the JSON flow definition for a given flow ID, as written
// by its creator.
func (i *Instance) FlowDef(flowID string) (string, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return "", ErrClosed
	}
	cid := C.CString(flowID)
	defer C.free(unsafe.Pointer(cid))

	// First call with NULL buffer to learn the required size.
	var size C.size_t
	rc := C.mxlGetFlowDef(i.handle, cid, nil, &size)
	// MXL_ERR_INVALID_ARG is the documented signal that size was written.
	if rc != C.MXL_STATUS_OK && rc != C.MXL_ERR_INVALID_ARG {
		return "", statusErr(rc)
	}
	if size == 0 {
		return "", nil
	}

	buf := C.malloc(size)
	defer C.free(buf)
	if err := statusErr(C.mxlGetFlowDef(i.handle, cid, (*C.char)(buf), &size)); err != nil {
		return "", err
	}
	// size includes the trailing NUL.
	n := int(size)
	if n > 0 && (*[1 << 30]byte)(buf)[n-1] == 0 {
		n--
	}
	return string((*[1 << 30]byte)(buf)[:n:n]), nil
}

// rawHandle returns the underlying C handle if the instance is still open.
// Callers must hold i.mu (R or W).
func (i *Instance) rawHandle() C.mxlInstance {
	return i.handle
}

// Handle returns the underlying C handle as an opaque pointer for use
// by the sister fabrics sub-package, which crosses a cgo package
// boundary and therefore cannot see C.mxlInstance directly. Callers
// must not retain the pointer past Close; the fabrics types pin the
// Instance so practical use stays safe.
func (i *Instance) Handle() unsafe.Pointer {
	return unsafe.Pointer(i.handle)
}
