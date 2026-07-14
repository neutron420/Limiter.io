package http

import "testing"

func TestParseCORSOrigins(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{"*"}},
		{"   ", []string{"*"}},
		{"*", []string{"*"}},
		{"https://a.com", []string{"https://a.com"}},
		{"https://a.com, https://b.com ", []string{"https://a.com", "https://b.com"}},
		{" , ,", []string{"*"}},
	}
	for _, c := range cases {
		got := parseCORSOrigins(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("parseCORSOrigins(%q) len = %d, want %d (%v)", c.in, len(got), len(c.want), got)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("parseCORSOrigins(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestOriginAllowed(t *testing.T) {
	allow := []string{"https://app.limiter.io", "https://limiter.io"}
	if !originAllowed("https://app.limiter.io", allow) {
		t.Error("expected listed origin to be allowed")
	}
	if originAllowed("https://evil.com", allow) {
		t.Error("expected unlisted origin to be rejected")
	}
	if originAllowed("", allow) {
		t.Error("empty origin should not be allowed")
	}
}
