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
	ProviderAuto  Provider = C.MXL_FABRICS_PROVIDER_AUTO
	ProviderTCP   Provider = C.MXL_FABRICS_PROVIDER_TCP
	ProviderVerbs Provider = C.MXL_FABRICS_PROVIDER_VERBS
	ProviderEFA   Provider = C.MXL_FABRICS_PROVIDER_EFA
	ProviderSHM   Provider = C.MXL_FABRICS_PROVIDER_SHM
)

// String returns the canonical name of the provider as accepted by
// mxlFabricsProviderFromString. Unknown values return an empty string.
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

// ParseProvider converts the canonical string form ("auto", "tcp",
// "verbs", "efa", "shm") into a Provider.
func ParseProvider(name string) (Provider, error) {
	if name == "" {
		return ProviderAuto, errors.New("mxl/fabrics: empty provider name")
	}
	c := C.CString(name)
	defer C.free(unsafe.Pointer(c))
	var p C.mxlFabricsProvider
	if err := fabricsStatusErr(C.mxlFabricsProviderFromString(c, &p)); err != nil {
		return ProviderAuto, err
	}
	return Provider(p), nil
}
