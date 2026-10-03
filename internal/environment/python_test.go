package environment

import (
	"context"
	"reflect"
	"testing"

	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

func TestParsePythonVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Python 3.12.0", "3.12.0"},
		{"Python 3.9.6", "3.9.6"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parsePythonVersion(c.in); got != c.want {
			t.Errorf("parsePythonVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectPythonPyenvActive(t *testing.T) {
	run := fakeRunner(map[string]string{
		"pyenv versions --bare": "3.12.0\n",
		"pyenv version":         "3.12.0 (set by /p/.python-version)\n",
		"python3 --version":     "Python 3.12.0",
	})
	got := detect(context.Background(), run, pythonSpec, "/Users/x/.pyenv/shims/python3")

	if got.Manager != versionmanager.ManagerPyenv {
		t.Errorf("Manager = %q, want pyenv", got.Manager)
	}
	if got.Active != "3.12.0" || got.Source != "/p/.python-version" {
		t.Errorf("Active/Source = (%q, %q), want (3.12.0, /p/.python-version)", got.Active, got.Source)
	}
	if !got.Available {
		t.Error("Available = false, want true")
	}
	if !reflect.DeepEqual(got.Installed, []string{"3.12.0"}) {
		t.Errorf("Installed = %#v, want [3.12.0]", got.Installed)
	}
}

func TestDetectPythonPyenvFallsBackToSystem(t *testing.T) {
	// pyenv 已装但未激活具体版本:shim 回落到系统 Python,Manager 记为 system;
	// Installed 仍如实列出 pyenv 中已安装的版本,供 check 提示。
	run := fakeRunner(map[string]string{
		"pyenv versions --bare": "3.12.0\n",
		"pyenv version":         "system (set by /Users/x/.pyenv/version)\n",
		"python3 --version":     "Python 3.9.6",
	})
	got := detect(context.Background(), run, pythonSpec, "/Users/x/.pyenv/shims/python3")

	if got.Manager != "system" {
		t.Errorf("Manager = %q, want system", got.Manager)
	}
	if got.Active != "3.9.6" {
		t.Errorf("Active = %q, want 3.9.6", got.Active)
	}
	if !reflect.DeepEqual(got.Installed, []string{"3.12.0"}) {
		t.Errorf("Installed = %#v, want [3.12.0]", got.Installed)
	}
}
