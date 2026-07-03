package fabrics

/*
#include <mxl/fabrics.h>
*/
import "C"

// InterfaceCapFlags is a bitset of fabric interface capabilities.
// Values mirror the C mxlFabricsInterfaceCapFlags enum.
type InterfaceCapFlags uint64

const (
	// InterfaceCapBlockingOperations: the interface supports blocking
	// operations.
	InterfaceCapBlockingOperations InterfaceCapFlags = C.MXL_FABRICS_IFACE_CAP_BLOCKING_OPERATIONS
	// InterfaceCapRemoteWrite: the interface supports remote-write
	// operations.
	InterfaceCapRemoteWrite InterfaceCapFlags = C.MXL_FABRICS_IFACE_CAP_REMOTE_WRITE
	// InterfaceCapSendReceive: the interface supports send/receive
	// type operations.
	InterfaceCapSendReceive InterfaceCapFlags = C.MXL_FABRICS_IFACE_CAP_SEND_RECEIVE
)

// InterfaceCaps describes the capabilities of a fabric interface,
// mirroring the C mxlFabricsInterfaceCaps struct.
type InterfaceCaps struct {
	// Flags is a bitset of InterfaceCapFlags values.
	Flags InterfaceCapFlags
	// MaxMessageSize is the maximum message size supported on the
	// interface.
	MaxMessageSize uint64
}

// InterfaceConfig identifies the local fabric interface for a Target
// or Initiator setup, mirroring the C mxlFabricsInterfaceConfig
// struct. The C struct's attr field is filled in by libmxl-fabrics
// when it enumerates interfaces and ignored by the setup functions;
// it has no Go counterpart.
type InterfaceConfig struct {
	// Provider selects the libmxl-fabrics provider.
	Provider Provider
	// Caps are the interface capabilities.
	Caps InterfaceCaps
	// Address is the node/service address of the interface.
	Address EndpointAddress
}

// ifaceCBuf bundles a C-side mxlFabricsInterfaceConfig with the C
// strings its address points at, so the caller can free them as one
// unit.
type ifaceCBuf struct {
	iface C.mxlFabricsInterfaceConfig
	addr  *endpointCBuf
}

func (c InterfaceConfig) toC() *ifaceCBuf {
	buf := &ifaceCBuf{addr: c.Address.toC()}
	buf.iface.version = C.MXL_FABRICS_API_VERSION
	buf.iface.provider = C.mxlFabricsProvider(c.Provider)
	buf.iface.caps.version = C.MXL_FABRICS_API_VERSION
	buf.iface.caps.flags = C.uint64_t(c.Caps.Flags)
	buf.iface.caps.maxMessageSize = C.uint64_t(c.Caps.MaxMessageSize)
	buf.iface.address = buf.addr.addr
	buf.iface.attr = nil
	return buf
}

func (b *ifaceCBuf) free() {
	b.addr.free()
}
