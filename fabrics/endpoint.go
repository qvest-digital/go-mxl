package fabrics

/*
#include <stdlib.h>
#include <mxl/fabrics.h>
*/
import "C"

import "unsafe"

// EndpointAddress is the bind address for a libmxl-fabrics endpoint,
// modelled on hostname/port pairs. The actual values vary by provider
// (for TCP this is an IP and port). Empty fields map to NULL on the
// C side.
type EndpointAddress struct {
	// Node is the address part (typically an IP).
	Node string
	// Service is the port or service identifier.
	Service string
}

// cBuf bundles a C-side mxlFabricsEndpointAddress with the C strings
// it points at, so the caller can free them as one unit.
type endpointCBuf struct {
	addr    C.mxlFabricsEndpointAddress
	node    *C.char
	service *C.char
}

func (a EndpointAddress) toC() *endpointCBuf {
	buf := &endpointCBuf{}
	if a.Node != "" {
		buf.node = C.CString(a.Node)
	}
	if a.Service != "" {
		buf.service = C.CString(a.Service)
	}
	buf.addr.node = buf.node
	buf.addr.service = buf.service
	return buf
}

func (b *endpointCBuf) free() {
	if b.node != nil {
		C.free(unsafe.Pointer(b.node))
	}
	if b.service != nil {
		C.free(unsafe.Pointer(b.service))
	}
}
