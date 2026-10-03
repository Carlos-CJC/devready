package environment

import (
	"context"
	"testing"
)

func TestParseDockerVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Docker version 29.8.1, build 4a63305", "29.8.1"},
		{"Docker version 27.0.3, build abcdef", "27.0.3"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseDockerVersion(c.in); got != c.want {
			t.Errorf("parseDockerVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectDocker(t *testing.T) {
	run := fakeRunner(map[string]string{
		"docker --version": "Docker version 29.8.1, build 4a63305",
	})
	got := detect(context.Background(), run, dockerSpec, "/usr/local/bin/docker")
	if got.Active != "29.8.1" {
		t.Errorf("Active = %q, want 29.8.1", got.Active)
	}
}
