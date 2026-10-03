package versionmanager

import (
	"context"
	"errors"
	"testing"
)

func TestIsPyenvShim(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/Users/me/.pyenv/shims/python3", true},
		{"/Users/me/.asdf/shims/python3", false},
		{"/usr/bin/python3", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsPyenvShim(c.path); got != c.want {
			t.Errorf("IsPyenvShim(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestParsePyenvVersion(t *testing.T) {
	cases := []struct {
		name        string
		out         string
		wantVersion string
		wantSource  string
	}{
		{"激活版本", "3.12.0 (set by /Users/me/.python-version)", "3.12.0", "/Users/me/.python-version"},
		{"回落系统", "system (set by /Users/me/.pyenv/version)", "", "/Users/me/.pyenv/version"},
		{"无来源", "3.12.0", "3.12.0", ""},
		{"空输出", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v, src := ParsePyenvVersion(c.out)
			if v != c.wantVersion || src != c.wantSource {
				t.Errorf("ParsePyenvVersion(%q) = (%q, %q), want (%q, %q)", c.out, v, src, c.wantVersion, c.wantSource)
			}
		})
	}
}

func TestParsePyenvVersions(t *testing.T) {
	got := ParsePyenvVersions("3.12.0\n\n3.11.4\n")
	if len(got) != 2 || got[0] != "3.12.0" || got[1] != "3.11.4" {
		t.Errorf("ParsePyenvVersions() = %v, want [3.12.0 3.11.4]", got)
	}
}

func TestPyenvActive(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) (string, error) {
		if name != "pyenv" || len(args) != 1 || args[0] != "version" {
			return "", errors.New("unexpected command")
		}
		return "3.12.0 (set by /Users/me/.python-version)\n", nil
	}
	v, src, err := PyenvActive(context.Background(), run)
	if err != nil || v != "3.12.0" || src != "/Users/me/.python-version" {
		t.Fatalf("PyenvActive() = (%q, %q, %v)", v, src, err)
	}
}
