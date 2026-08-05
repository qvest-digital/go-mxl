//go:build mxl_integration

package fabrics

import (
	"encoding/json"
	"testing"
)

func TestInterfacesEnumeratesTheHost(t *testing.T) {
	_, fi, _ := newTestFabrics(t)

	ifaces, err := fi.Interfaces(nil)
	if err != nil {
		t.Fatalf("Interfaces(nil): %v", err)
	}
	if len(ifaces) == 0 {
		t.Fatal("want at least one interface on a host libmxl-fabrics can run on")
	}
	for _, iface := range ifaces {
		if iface.Caps.Flags == 0 {
			t.Errorf("interface %v reports no capabilities", iface.Provider)
		}
		if iface.Provider.String() == "" {
			t.Errorf("interface reports an unnameable provider %d", iface.Provider)
		}
	}

	// The selection a setup would make. Remote write is the only
	// transfer mode libmxl-fabrics implements, so a host that can run
	// it at all must offer one.
	if _, ok := SelectInterface(ifaces, InterfaceCapRemoteWrite); !ok {
		t.Fatal("no interface offers remote write")
	}
}

func TestInterfacesFilterByProvider(t *testing.T) {
	_, fi, _ := newTestFabrics(t)

	all, err := fi.Interfaces(nil)
	if err != nil {
		t.Fatalf("Interfaces(nil): %v", err)
	}
	if len(all) == 0 {
		t.Skip("no interfaces to filter")
	}

	want := all[0].Provider
	got, err := fi.Interfaces(&InterfaceConfig{Provider: want})
	if err != nil {
		t.Fatalf("Interfaces(provider=%v): %v", want, err)
	}
	for _, iface := range got {
		if iface.Provider != want {
			t.Fatalf("filter returned provider %v, want only %v", iface.Provider, want)
		}
	}
}

// Attr is the only source of device detail in the API: which NIC an
// interface belongs to, what state its link is in, how fast it is. A
// caller that has to tell one of a host's interfaces from another has
// nothing else to go on, so a silently dropped attr would be invisible
// until it mattered.
func TestInterfacesReportDeviceAttributes(t *testing.T) {
	_, fi, _ := newTestFabrics(t)

	ifaces, err := fi.Interfaces(nil)
	if err != nil {
		t.Fatalf("Interfaces(nil): %v", err)
	}
	if len(ifaces) == 0 {
		t.Fatal("want at least one interface on a host libmxl-fabrics can run on")
	}

	withAttr := 0
	for _, iface := range ifaces {
		t.Logf("provider %v node %q attr %s", iface.Provider, iface.Address.Node, iface.Attr)
		if iface.Attr == "" {
			continue
		}
		withAttr++
		var doc map[string]any
		if err := json.Unmarshal([]byte(iface.Attr), &doc); err != nil {
			t.Errorf("provider %v attr is not a JSON object: %v\nattr: %s",
				iface.Provider, err, iface.Attr)
			continue
		}
		// Which keys appear depends on the provider and the hardware,
		// but the tcp provider takes device_name from the domain name,
		// so it is there on any host that enumerates tcp at all. It is
		// the only name that ties an interface back to a netdev.
		if iface.Provider == ProviderTCP {
			if name, _ := doc["device_name"].(string); name == "" {
				t.Errorf("tcp interface %q carries no device_name: %s",
					iface.Address.Node, iface.Attr)
			}
		}
	}
	if withAttr == 0 {
		t.Error("no interface carried attr; the enumeration output is being dropped")
	}
}

// The C field is enumeration output that the setup functions ignore.
// Setting it on a config bound for a setup must not reach the library.
func TestInterfaceAttrIsNotSentToSetup(t *testing.T) {
	_, fi, w := newTestFabrics(t)
	tgt, err := fi.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	t.Cleanup(func() { tgt.Close() })

	info, err := tgt.Setup(TargetConfig{
		Interface: InterfaceConfig{
			Provider: ProviderTCP,
			Caps:     InterfaceCaps{Flags: InterfaceCapRemoteWrite, MaxMessageSize: 1 << 20},
			Address:  EndpointAddress{Node: "127.0.0.1", Service: "0"},
			Attr:     `{"device_name":"ignored"}`,
		},
		Writer: w,
	})
	if err != nil {
		t.Fatalf("Setup with attr set: %v", err)
	}
	if err := info.Close(); err != nil {
		t.Fatalf("TargetInfo.Close: %v", err)
	}
}

func TestInterfacesAfterCloseReportsClosed(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	if err := fi.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := fi.Interfaces(nil); err == nil {
		t.Fatal("want an error from a closed instance")
	}
}
