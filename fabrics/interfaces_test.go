package fabrics

import "testing"

func TestSelectInterfacePrefersTheFastestProvider(t *testing.T) {
	ifaces := []InterfaceConfig{
		{Provider: ProviderSHM, Caps: InterfaceCaps{Flags: InterfaceCapRemoteWrite}},
		{Provider: ProviderTCP, Caps: InterfaceCaps{Flags: InterfaceCapRemoteWrite}},
		{Provider: ProviderVerbs, Caps: InterfaceCaps{Flags: InterfaceCapRemoteWrite}},
	}
	got, ok := SelectInterface(ifaces, InterfaceCapRemoteWrite)
	if !ok {
		t.Fatal("want a selection")
	}
	if got.Provider != ProviderVerbs {
		t.Fatalf("Provider = %v, want verbs", got.Provider)
	}
}

func TestSelectInterfaceSkipsInterfacesMissingTheCap(t *testing.T) {
	// The faster provider here cannot do remote write, so the slower
	// one is the only usable answer.
	ifaces := []InterfaceConfig{
		{Provider: ProviderVerbs, Caps: InterfaceCaps{Flags: InterfaceCapSendReceive}},
		{Provider: ProviderTCP, Caps: InterfaceCaps{Flags: InterfaceCapRemoteWrite}},
	}
	got, ok := SelectInterface(ifaces, InterfaceCapRemoteWrite)
	if !ok {
		t.Fatal("want a selection")
	}
	if got.Provider != ProviderTCP {
		t.Fatalf("Provider = %v, want tcp", got.Provider)
	}
}

func TestSelectInterfaceRequiresEveryRequestedCap(t *testing.T) {
	ifaces := []InterfaceConfig{
		{Provider: ProviderVerbs, Caps: InterfaceCaps{Flags: InterfaceCapRemoteWrite}},
	}
	if _, ok := SelectInterface(ifaces, InterfaceCapRemoteWrite|InterfaceCapBlockingOperations); ok {
		t.Fatal("an interface missing one requested cap must not be selected")
	}
}

func TestSelectInterfaceEmpty(t *testing.T) {
	if _, ok := SelectInterface(nil, InterfaceCapRemoteWrite); ok {
		t.Fatal("no interfaces cannot yield a selection")
	}
}

func TestSelectInterfaceCarriesAddressAndMaxMessageSize(t *testing.T) {
	// The point of selecting rather than hand-building a config: the
	// address and message size come from the provider, and a caller
	// has no other way to learn them.
	ifaces := []InterfaceConfig{{
		Provider: ProviderTCP,
		Caps:     InterfaceCaps{Flags: InterfaceCapRemoteWrite, MaxMessageSize: 4096},
		Address:  EndpointAddress{Node: "10.0.0.7"},
	}}
	got, ok := SelectInterface(ifaces, InterfaceCapRemoteWrite)
	if !ok {
		t.Fatal("want a selection")
	}
	if got.Address.Node != "10.0.0.7" || got.Caps.MaxMessageSize != 4096 {
		t.Fatalf("selection dropped provider-supplied fields: %+v", got)
	}
}

func TestProviderPriorityRanksUnknownLast(t *testing.T) {
	if providerPriority(Provider(99)) != 0 {
		t.Fatal("an unknown provider must never outrank a known one")
	}
	if providerPriority(ProviderEFA) <= providerPriority(ProviderVerbs) {
		t.Fatal("efa must outrank verbs")
	}
	if providerPriority(ProviderTCP) <= providerPriority(ProviderSHM) {
		t.Fatal("tcp must outrank shm")
	}
}
