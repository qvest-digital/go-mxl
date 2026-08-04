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
	cases := []Provider{ProviderAny, ProviderTCP, ProviderVerbs, ProviderEFA, ProviderSHM}
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
	// "any" gained a parseable form in libmxl-fabrics after
	// v1.1.0-beta-1. It names every provider rather than selecting one,
	// so callers that treat a parse failure as "this is not a concrete
	// provider" need their own check.
	got, err := ParseProvider("any")
	if err != nil {
		t.Fatalf("ParseProvider(\"any\"): %v", err)
	}
	if got != ProviderAny {
		t.Fatalf("ParseProvider(\"any\") = %d, want ProviderAny (%d)", got, ProviderAny)
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
