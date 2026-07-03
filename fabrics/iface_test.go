package fabrics

import (
	"testing"
)

// The C header defines the cap flags via MXL_FABRICS_FLAG(n) = 1<<n.
// Pin the bit positions so an upstream renumbering is caught here
// instead of surfacing as silent capability mismatches.
func TestInterfaceCapFlagValues(t *testing.T) {
	cases := []struct {
		name string
		got  InterfaceCapFlags
		want InterfaceCapFlags
	}{
		{"InterfaceCapBlockingOperations", InterfaceCapBlockingOperations, 1 << 0},
		{"InterfaceCapRemoteWrite", InterfaceCapRemoteWrite, 1 << 1},
		{"InterfaceCapSendReceive", InterfaceCapSendReceive, 1 << 2},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestInterfaceCapFlagsDistinct(t *testing.T) {
	all := InterfaceCapBlockingOperations | InterfaceCapRemoteWrite | InterfaceCapSendReceive
	n := 0
	for f := all; f != 0; f &= f - 1 {
		n++
	}
	if n != 3 {
		t.Fatalf("cap flags overlap: union %b has %d bits, want 3", all, n)
	}
}

// TestTargetSetupWithCaps drives a full Setup with every config field
// populated, including Caps, so the Go-to-C conversion of the nested
// interface struct is exercised against the linked library.
func TestTargetSetupWithCaps(t *testing.T) {
	_, fi, w := newTestFabrics(t)
	tgt, err := fi.NewTarget()
	if err != nil {
		t.Fatalf("NewTarget: %v", err)
	}
	t.Cleanup(func() { tgt.Close() })

	info, err := tgt.Setup(TargetConfig{
		Interface: InterfaceConfig{
			Provider: ProviderTCP,
			Caps: InterfaceCaps{
				Flags:          InterfaceCapBlockingOperations | InterfaceCapRemoteWrite | InterfaceCapSendReceive,
				MaxMessageSize: 1 << 20,
			},
			Address: EndpointAddress{Node: "127.0.0.1", Service: "0"},
		},
		Writer: w,
	})
	if err != nil {
		t.Fatalf("Setup with caps: %v", err)
	}
	if err := info.Close(); err != nil {
		t.Fatalf("TargetInfo.Close: %v", err)
	}
}
