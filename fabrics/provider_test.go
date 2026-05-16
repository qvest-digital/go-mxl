package fabrics

import (
	"testing"
)

func TestProviderStringNonEmpty(t *testing.T) {
	cases := []Provider{ProviderAuto, ProviderTCP, ProviderVerbs, ProviderEFA, ProviderSHM}
	seen := map[string]Provider{}
	for _, p := range cases {
		s := p.String()
		if s == "" {
			t.Errorf("Provider(%d).String() returned empty", p)
			continue
		}
		if prev, dup := seen[s]; dup {
			t.Errorf("Provider(%d).String() = %q clashes with Provider(%d)", p, s, prev)
		}
		seen[s] = p
	}
}

func TestProviderRoundTrip(t *testing.T) {
	// libmxl-fabrics' string parser only accepts the concrete providers;
	// the AUTO sentinel stringifies to "auto" but does not parse back.
	cases := []Provider{ProviderTCP, ProviderVerbs, ProviderEFA, ProviderSHM}
	for _, p := range cases {
		s := p.String()
		got, err := ParseProvider(s)
		if err != nil {
			t.Errorf("ParseProvider(%q): %v", s, err)
			continue
		}
		if got != p {
			t.Errorf("ParseProvider(%q) = %d, want %d", s, got, p)
		}
	}
}

func TestParseProviderAuto(t *testing.T) {
	if _, err := ParseProvider("auto"); err == nil {
		t.Fatal("libmxl-fabrics started accepting \"auto\" — update Provider docs")
	}
}

func TestParseProviderEmpty(t *testing.T) {
	if _, err := ParseProvider(""); err == nil {
		t.Fatal("ParseProvider(\"\") returned nil error")
	}
}

func TestParseProviderUnknown(t *testing.T) {
	if _, err := ParseProvider("not-a-real-provider"); err == nil {
		t.Fatal("ParseProvider(unknown) returned nil error")
	}
}
