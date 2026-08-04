//go:build mxl_integration

package fabrics

import "testing"

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

func TestInterfacesAfterCloseReportsClosed(t *testing.T) {
	_, fi, _ := newTestFabrics(t)
	if err := fi.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := fi.Interfaces(nil); err == nil {
		t.Fatal("want an error from a closed instance")
	}
}
