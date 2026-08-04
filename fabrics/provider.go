package fabrics

/*
#include <stdlib.h>
#include <mxl/fabrics.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Provider selects the libmxl-fabrics provider for a Target or
// Initiator setup. Values mirror the C mxlFabricsProvider enum.
type Provider int

const (
	// ProviderAny names every provider rather than selecting one. The
	// libmxl-fabrics header gives it as an input to interface
	// enumeration, where it means "do not filter by provider"; it
	// carried an auto-select reading that fell back to TCP until
	// v1.1.0-beta-1, and no longer does.
	ProviderAny   Provider = C.MXL_FABRICS_PROVIDER_ANY
	ProviderTCP   Provider = C.MXL_FABRICS_PROVIDER_TCP
	ProviderVerbs Provider = C.MXL_FABRICS_PROVIDER_VERBS
	ProviderEFA   Provider = C.MXL_FABRICS_PROVIDER_EFA
	ProviderSHM   Provider = C.MXL_FABRICS_PROVIDER_SHM
)

// String returns the canonical name of the provider as reported by
// mxlFabricsProviderToString. Unknown values return an empty string.
func (p Provider) String() string {
	var size C.size_t
	rc := C.mxlFabricsProviderToString(C.mxlFabricsProvider(p), nil, &size)
	if err := fabricsStatusErr(rc); err != nil || size == 0 {
		return ""
	}
	buf := C.malloc(size)
	defer C.free(buf)
	if err := fabricsStatusErr(C.mxlFabricsProviderToString(
		C.mxlFabricsProvider(p), (*C.char)(buf), &size)); err != nil {
		return ""
	}
	n := int(size)
	if n > 0 && (*[1 << 30]byte)(buf)[n-1] == 0 {
		n--
	}
	return string((*[1 << 30]byte)(buf)[:n:n])
}

// ParseProvider converts the canonical string form ("any", "tcp",
// "verbs", "efa", "shm") into a Provider. "any" parses to ProviderAny,
// which names no single provider -- see its documentation before
// passing the result to a setup function.
func ParseProvider(name string) (Provider, error) {
	if name == "" {
		return ProviderAny, errors.New("mxl/fabrics: empty provider name")
	}
	c := C.CString(name)
	defer C.free(unsafe.Pointer(c))
	var p C.mxlFabricsProvider
	if err := fabricsStatusErr(C.mxlFabricsProviderFromString(c, &p)); err != nil {
		return ProviderAny, err
	}
	return Provider(p), nil
}
