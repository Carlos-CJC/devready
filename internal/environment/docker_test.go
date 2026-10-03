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

func TestParseComposeVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Docker Compose version v2.24.0", "2.24.0"},
		{"Docker Compose version v5.5.1", "5.5.1"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseComposeVersion(c.in); got != c.want {
			t.Errorf("parseComposeVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectDockerCompose(t *testing.T) {
	run := fakeRunner(map[string]string{
		"docker compose version": "Docker Compose version v2.24.0",
	})
	got := detect(context.Background(), run, composeSpec, "/usr/local/bin/docker")
	if got.Active != "2.24.0" || !got.Available {
		t.Errorf("Active/Available = (%q, %v), want (2.24.0, true)", got.Active, got.Available)
	}
}
