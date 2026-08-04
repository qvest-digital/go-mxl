package fabrics

import "testing"

// libmxl-fabrics rejects a setup that requests no transfer capability.
// It used to default one, and every caller that relied on that would
// otherwise start failing at setup on the version that stopped.

func TestRequireTransferCapDefaultsToRemoteWrite(t *testing.T) {
	got := InterfaceConfig{}.requireTransferCap()
	if got.Caps.Flags != InterfaceCapRemoteWrite {
		t.Fatalf("Caps.Flags = %#x, want %#x", got.Caps.Flags, InterfaceCapRemoteWrite)
	}
}

func TestRequireTransferCapKeepsAnExplicitChoice(t *testing.T) {
	tests := []struct {
		name string
		in   InterfaceCapFlags
	}{
		{"remote write", InterfaceCapRemoteWrite},
		{"send receive", InterfaceCapSendReceive},
		{"both", InterfaceCapRemoteWrite | InterfaceCapSendReceive},
		{"remote write with blocking", InterfaceCapRemoteWrite | InterfaceCapBlockingOperations},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InterfaceConfig{Caps: InterfaceCaps{Flags: tc.in}}.requireTransferCap()
			if got.Caps.Flags != tc.in {
				t.Fatalf("Caps.Flags = %#x, want %#x unchanged", got.Caps.Flags, tc.in)
			}
		})
	}
}

func TestRequireTransferCapAddsToNonTransferFlags(t *testing.T) {
	// Blocking operations is not a transfer capability, so a config
	// carrying only that one is still missing what setup requires.
	got := InterfaceConfig{
		Caps: InterfaceCaps{Flags: InterfaceCapBlockingOperations},
	}.requireTransferCap()
	want := InterfaceCapBlockingOperations | InterfaceCapRemoteWrite
	if got.Caps.Flags != want {
		t.Fatalf("Caps.Flags = %#x, want %#x", got.Caps.Flags, want)
	}
}

func TestRequireTransferCapLeavesOtherFieldsAlone(t *testing.T) {
	in := InterfaceConfig{
		Provider: ProviderTCP,
		Caps:     InterfaceCaps{MaxMessageSize: 4096},
		Address:  EndpointAddress{Node: "10.0.0.1", Service: "9000"},
	}
	got := in.requireTransferCap()
	if got.Provider != in.Provider || got.Address != in.Address ||
		got.Caps.MaxMessageSize != in.Caps.MaxMessageSize {
		t.Fatalf("requireTransferCap altered more than Caps.Flags: %+v", got)
	}
}
