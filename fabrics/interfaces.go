package fabrics

/*
#include <stdlib.h>
#include <mxl/fabrics.h>
*/
import "C"

// Interfaces reports the fabric interfaces libmxl-fabrics can offer on
// this host.
//
// A nil query lists every interface of every provider. A non-nil query
// filters: only its Provider and Address fields are read, and a zero
// Provider (ProviderAny) means "do not filter by provider".
// Capability flags in a query are ignored, so an interface that comes
// back is not thereby known to satisfy any particular requirement --
// read Caps on the result and decide.
//
// The returned Caps and Address are what a setup expects to be handed
// back: they carry the provider's own view of the interface, including
// the MaxMessageSize a caller has no other way to learn. Attr carries
// the provider's description of the device behind the interface and is
// only ever populated here.
func (i *Instance) Interfaces(query *InterfaceConfig) ([]InterfaceConfig, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.handle == nil {
		return nil, ErrClosed()
	}

	var cQuery *C.mxlFabricsInterfaceConfig
	if query != nil {
		buf := query.toC()
		defer buf.free()
		// Copied out of the buffer before it is handed over: the
		// buffer also holds the Go pointer that owns the address
		// strings, and cgo refuses a pointer into an allocation
		// containing one. The copy is a bare C struct.
		q := buf.iface
		cQuery = &q
	}

	var list *C.mxlFabricsInterfaceList
	if err := fabricsStatusErr(C.mxlFabricsGetInterfaces(
		C.mxlFabricsInstance(i.handle), cQuery, &list)); err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	defer C.mxlFabricsFreeInterfaceList(list)

	var out []InterfaceConfig
	for entry := list; entry != nil; entry = entry.next {
		out = append(out, interfaceConfigFromC(&entry._interface))
	}
	return out, nil
}

// interfaceConfigFromC copies a C interface config into Go memory. The
// strings are copied rather than referenced because the list they live
// in is freed before this returns.
func interfaceConfigFromC(c *C.mxlFabricsInterfaceConfig) InterfaceConfig {
	return InterfaceConfig{
		Provider: Provider(c.provider),
		Caps: InterfaceCaps{
			Flags:          InterfaceCapFlags(c.caps.flags),
			MaxMessageSize: uint64(c.caps.maxMessageSize),
		},
		Address: EndpointAddress{
			Node:    goStringOrEmpty(c.address.node),
			Service: goStringOrEmpty(c.address.service),
		},
		Attr: goStringOrEmpty(c.attr),
	}
}

func goStringOrEmpty(s *C.char) string {
	if s == nil {
		return ""
	}
	return C.GoString(s)
}

// SelectInterface returns the interface from ifaces that carries every
// capability in want, preferring the fastest provider available.
//
// The preference order is the one libmxl-fabrics' own demo applies:
// EFA, then verbs, then TCP, then shared memory. Returns false when
// nothing in ifaces carries the requested capabilities.
func SelectInterface(ifaces []InterfaceConfig, want InterfaceCapFlags) (InterfaceConfig, bool) {
	var best InterfaceConfig
	found := false
	for _, iface := range ifaces {
		if iface.Caps.Flags&want != want {
			continue
		}
		if !found || providerPriority(iface.Provider) > providerPriority(best.Provider) {
			best = iface
			found = true
		}
	}
	return best, found
}

// providerPriority ranks providers by transfer performance. Unknown
// providers sort last so a provider added to libmxl-fabrics is never
// picked ahead of one this build knows how to drive.
func providerPriority(p Provider) int {
	switch p {
	case ProviderEFA:
		return 4
	case ProviderVerbs:
		return 3
	case ProviderTCP:
		return 2
	case ProviderSHM:
		return 1
	default:
		return 0
	}
}
