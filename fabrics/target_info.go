package fabrics

/*
#include <stdlib.h>
#include <mxl/fabrics.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

// TargetInfo is the libmxl-fabrics target descriptor that an initiator
// needs in order to send to a target. It is opaque and only meaningful
// when round-tripped via MarshalString/ParseTargetInfo across the
// wire.
type TargetInfo struct {
	mu     sync.Mutex
	handle C.mxlFabricsTargetInfo
}

// MarshalString serializes the target info into a string form suitable
// for transport to remote initiators.
func (t *TargetInfo) MarshalString() (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return "", errors.New("mxl/fabrics: target info closed")
	}
	var size C.size_t
	rc := C.mxlFabricsTargetInfoToString(t.handle, nil, &size)
	if err := fabricsStatusErr(rc); err != nil {
		return "", err
	}
	if size == 0 {
		return "", nil
	}
	buf := C.malloc(size)
	defer C.free(buf)
	if err := fabricsStatusErr(C.mxlFabricsTargetInfoToString(
		t.handle, (*C.char)(buf), &size)); err != nil {
		return "", err
	}
	n := int(size)
	if n > 0 && (*[1 << 30]byte)(buf)[n-1] == 0 {
		n--
	}
	return string((*[1 << 30]byte)(buf)[:n:n]), nil
}

// ParseTargetInfo reconstructs a TargetInfo from the string form
// produced by MarshalString.
func ParseTargetInfo(s string) (*TargetInfo, error) {
	c := C.CString(s)
	defer C.free(unsafe.Pointer(c))
	var h C.mxlFabricsTargetInfo
	if err := fabricsStatusErr(C.mxlFabricsTargetInfoFromString(c, &h)); err != nil {
		return nil, err
	}
	t := &TargetInfo{handle: h}
	runtime.SetFinalizer(t, func(x *TargetInfo) { _ = x.Close() })
	return t, nil
}

// Close frees the TargetInfo.
func (t *TargetInfo) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handle == nil {
		return nil
	}
	rc := C.mxlFabricsFreeTargetInfo(t.handle)
	t.handle = nil
	runtime.SetFinalizer(t, nil)
	return fabricsStatusErr(rc)
}

// rawHandle returns the underlying C handle. Caller must hold t.mu.
func (t *TargetInfo) rawHandle() C.mxlFabricsTargetInfo {
	return t.handle
}
