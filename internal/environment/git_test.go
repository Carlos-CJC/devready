package environment

import (
	"context"
	"testing"
)

func TestParseGitVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"git version 2.51.0", "2.51.0"},
		{"git version 2.39.5 (Apple Git-154)", "2.39.5"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseGitVersion(c.in); got != c.want {
			t.Errorf("parseGitVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectGitSystemPath(t *testing.T) {
	run := fakeRunner(map[string]string{
		"git --version": "git version 2.51.0",
	})
	got := detect(context.Background(), run, gitSpec, "/usr/bin/git")

	if got.Manager != "system" {
		t.Errorf("Manager = %q, want system", got.Manager)
	}
	if got.Active != "2.51.0" {
		t.Errorf("Active = %q, want 2.51.0", got.Active)
	}
	if !got.Available {
		t.Error("Available = false, want true")
	}
}

func TestDetectGitMissing(t *testing.T) {
	run := fakeRunner(map[string]string{}) // 所有命令都失败
	got := detect(context.Background(), run, gitSpec, "")

	if got.Manager != "unknown" {
		t.Errorf("Manager = %q, want unknown", got.Manager)
	}
	if got.Available || got.Active != "" {
		t.Errorf("未安装时 Available=%v Active=%q, want false/空", got.Available, got.Active)
	}
}
