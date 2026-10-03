package versionmanager

import (
	"reflect"
	"testing"
)

func TestIsASDFShim(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/Users/cjc/.asdf/shims/go", true},
		{"/usr/local/go/bin/go", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsASDFShim(c.path); got != c.want {
			t.Errorf("IsASDFShim(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestParseList(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want []string
	}{
		{"激活版本带星号", " *1.22.0\n", []string{"1.22.0"}},
		{"多版本", " *1.22.0\n  1.20.0\n", []string{"1.22.0", "1.20.0"}},
		{"无内容", "", nil},
		{"无星号", "1.22.0\n", []string{"1.22.0"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ParseList(c.out); !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseList() = %#v, want %#v", got, c.want)
			}
		})
	}
}

func TestParseCurrent(t *testing.T) {
	const withVersion = "Name            Version         Source                                      Installed\n" +
		"golang          1.22.0          /Volumes/Data/projects/go_project/devready/.tool-versions true\n"
	const unset = "Name            Version         Source          Installed\n" +
		"golang          ______          ______          \n"

	cases := []struct {
		name       string
		out        string
		wantVer    string
		wantSource string
	}{
		{"已设置", withVersion, "1.22.0", "/Volumes/Data/projects/go_project/devready/.tool-versions"},
		{"未设置", unset, "", ""},
		{"空输出", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ver, src := ParseCurrent(c.out)
			if ver != c.wantVer || src != c.wantSource {
				t.Errorf("ParseCurrent() = (%q, %q), want (%q, %q)", ver, src, c.wantVer, c.wantSource)
			}
		})
	}
}

func TestParseToolVersions(t *testing.T) {
	cases := []struct {
		name    string
		content string
		plugin  string
		want    string
	}{
		{"命中的插件", "golang 1.22.0\n", "golang", "1.22.0"},
		{"多插件", "nodejs 20.0.0\ngolang 1.21.0\n", "golang", "1.21.0"},
		{"行尾注释", "golang 1.21.0 # pinned\n", "golang", "1.21.0"},
		{"整行注释", "# golang 1.0.0\ngolang 1.22.0\n", "golang", "1.22.0"},
		{"未命中", "nodejs 20.0.0\n", "golang", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ParseToolVersions(c.content, c.plugin); got != c.want {
				t.Errorf("ParseToolVersions() = %q, want %q", got, c.want)
			}
		})
	}
}
