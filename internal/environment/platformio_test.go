package environment

import (
	"context"
	"testing"
)

func TestParsePlatformIOVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"PlatformIO Core, version 6.1.15", "6.1.15"},
		{"PlatformIO Core, version 6.1.0", "6.1.0"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parsePlatformIOVersion(c.in); got != c.want {
			t.Errorf("parsePlatformIOVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectPlatformIO(t *testing.T) {
	run := fakeRunner(map[string]string{
		"pio --version": "PlatformIO Core, version 6.1.15",
	})
	got := detect(context.Background(), run, platformioSpec, "/usr/local/bin/pio")
	if got.Active != "6.1.15" || !got.Available {
		t.Errorf("Active/Available = (%q, %v), want (6.1.15, true)", got.Active, got.Available)
	}
}
