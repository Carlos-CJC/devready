package environment

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cjc/devscope/internal/versionmanager"
)

// fakeRunner 依据完整命令行返回预设输出;值为空串视为命令失败。
func fakeRunner(responses map[string]string) versionmanager.Runner {
	return func(_ context.Context, name string, args ...string) (string, error) {
		key := strings.TrimSpace(name + " " + strings.Join(args, " "))
		out, ok := responses[key]
		if !ok || out == "" {
			return "", fmt.Errorf("命令失败: %s", key)
		}
		return out, nil
	}
}

func TestParseGoVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"go version go1.22.0 darwin/arm64", "1.22.0"},
		{"go version go1.21.5 linux/amd64", "1.21.5"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseGoVersion(c.in); got != c.want {
			t.Errorf("parseGoVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectGoWithASDFActive(t *testing.T) {
	run := fakeRunner(map[string]string{
		"asdf list golang":    " *1.22.0\n",
		"asdf current golang": "Name            Version         Source                     Installed\ngolang          1.22.0          /p/.tool-versions          true\n",
		"go version":          "go version go1.22.0 darwin/arm64",
	})
	got := detectGoWith(context.Background(), run, "/Users/x/.asdf/shims/go")

	if got.Manager != versionmanager.ManagerASDF {
		t.Errorf("Manager = %q, want asdf", got.Manager)
	}
	if got.Active != "1.22.0" || got.Source != "/p/.tool-versions" {
		t.Errorf("Active/Source = (%q, %q), want (1.22.0, /p/.tool-versions)", got.Active, got.Source)
	}
	if !got.Available {
		t.Error("Available = false, want true")
	}
	if !reflect.DeepEqual(got.Installed, []string{"1.22.0"}) {
		t.Errorf("Installed = %#v, want [1.22.0]", got.Installed)
	}
}

func TestDetectGoWithASDFNotSet(t *testing.T) {
	run := fakeRunner(map[string]string{
		"asdf list golang":    "  1.22.0\n",
		"asdf current golang": "Name            Version         Source          Installed\ngolang          ______          ______          \n",
		// "go version" 未提供 -> 模拟 shim 无版本时执行失败
	})
	got := detectGoWith(context.Background(), run, "/Users/x/.asdf/shims/go")

	if got.Manager != versionmanager.ManagerASDF {
		t.Errorf("Manager = %q, want asdf", got.Manager)
	}
	if got.Active != "" {
		t.Errorf("Active = %q, want empty", got.Active)
	}
	if got.Available {
		t.Error("Available = true, want false")
	}
	if !reflect.DeepEqual(got.Installed, []string{"1.22.0"}) {
		t.Errorf("Installed = %#v, want [1.22.0]", got.Installed)
	}
}

func TestDetectGoSystemPath(t *testing.T) {
	run := fakeRunner(map[string]string{
		"go version": "go version go1.22.0 darwin/arm64",
	})
	got := detectGoWith(context.Background(), run, "/usr/local/go/bin/go")

	if got.Manager != "system" {
		t.Errorf("Manager = %q, want system", got.Manager)
	}
	if got.Active != "1.22.0" {
		t.Errorf("Active = %q, want 1.22.0", got.Active)
	}
}
