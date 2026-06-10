package fabrics

import (
	"testing"
)

func TestProviderStringNonEmpty(t *testing.T) {
	cases := []Provider{ProviderAny, ProviderTCP, ProviderVerbs, ProviderEFA, ProviderSHM}
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

func TestProviderAnyString(t *testing.T) {
	if got := ProviderAny.String(); got != "any" {
		t.Fatalf("ProviderAny.String() = %q, want %q", got, "any")
	}
}

func TestProviderRoundTrip(t *testing.T) {
	// libmxl-fabrics' string parser only accepts the concrete providers;
	// the ANY sentinel stringifies to "any" but does not parse back.
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

func TestParseProviderAny(t *testing.T) {
	if _, err := ParseProvider("any"); err == nil {
		t.Fatal("libmxl-fabrics started accepting \"any\" -- update Provider docs")
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
