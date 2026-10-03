package check

import (
	"strings"
	"testing"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/project"
)

func TestGoVersion(t *testing.T) {
	cases := []struct {
		name     string
		active   string
		required string
		want     domain.CheckStatus
	}{
		{"满足", "1.22.0", "1.22.0", domain.StatusPass},
		{"高于要求", "1.22.8", "1.22.0", domain.StatusPass},
		{"低于要求", "1.21.0", "1.22.0", domain.StatusFail},
		{"未激活", "", "1.22.0", domain.StatusFail},
		{"无要求", "1.22.0", "", domain.StatusSkip},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GoVersion(c.active, c.required); got.Status != c.want {
				t.Errorf("GoVersion(%q, %q).Status = %v, want %v", c.active, c.required, got.Status, c.want)
			}
		})
	}
}

func TestGoVersionRequiredLabel(t *testing.T) {
	got := GoVersion("1.22.0", "1.22.0")
	if got.Required != ">=1.22.0" {
		t.Errorf("Required = %q, want >=1.22.0", got.Required)
	}
}

func TestRun(t *testing.T) {
	proj := project.Project{
		Requirements: []project.Requirement{
			{Kind: "go", Name: "Go", Version: "1.22.0", Source: "go.mod"},
		},
	}
	results := Run(proj, Tools{Go: domain.ToolVersion{Active: "1.22.0"}})
	if len(results) != 1 || results[0].Status != domain.StatusPass {
		t.Fatalf("Run() = %#v, want 1 条 PASS", results)
	}
}

func TestTally(t *testing.T) {
	results := []domain.CheckResult{
		{Status: domain.StatusPass},
		{Status: domain.StatusPass},
		{Status: domain.StatusFail},
		{Status: domain.StatusWarn},
		{Status: domain.StatusSkip},
	}
	passed, failed, warned, skipped := Tally(results)
	if passed != 2 || failed != 1 || warned != 1 || skipped != 1 {
		t.Errorf("Tally() = (%d,%d,%d,%d), want (2,1,1,1)", passed, failed, warned, skipped)
	}
}

func TestPythonVersion(t *testing.T) {
	cases := []struct {
		name     string
		tool     domain.ToolVersion
		required string
		want     domain.CheckStatus
	}{
		{"满足", domain.ToolVersion{Active: "3.12.0"}, "3.11", domain.StatusPass},
		{"低于要求", domain.ToolVersion{Active: "3.9.6", Installed: []string{"3.12.0"}}, "3.11", domain.StatusFail},
		{"版本未知", domain.ToolVersion{Active: "3.12.0"}, "", domain.StatusWarn},
		{"不可用", domain.ToolVersion{}, "3.11", domain.StatusFail},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := PythonVersion(c.tool, c.required, ""); got.Status != c.want {
				t.Errorf("PythonVersion().Status = %v, want %v", got.Status, c.want)
			}
		})
	}
}

func TestPythonVersionHintsInstalled(t *testing.T) {
	got := PythonVersion(domain.ToolVersion{Active: "3.9.6", Installed: []string{"3.12.0"}}, "3.11", "")
	if !strings.Contains(got.Message, "3.12.0") {
		t.Errorf("Message = %q, want 提及已装版本 3.12.0", got.Message)
	}
}

func TestPythonVersionCurrentWithEnv(t *testing.T) {
	got := PythonVersion(domain.ToolVersion{Active: "3.12.0"}, "", ".venv")
	if got.Current != "3.12.0 (.venv)" {
		t.Errorf("Current = %q, want 3.12.0 (.venv)", got.Current)
	}
}

func TestPresence(t *testing.T) {
	ok := Presence("Docker", domain.ToolVersion{Active: "29.8.1", Available: true})
	if ok.Status != domain.StatusPass || ok.Current != "29.8.1" {
		t.Errorf("Presence(可用) = %#v, want PASS/29.8.1", ok)
	}
	bad := Presence("Docker", domain.ToolVersion{})
	if bad.Status != domain.StatusFail {
		t.Errorf("Presence(不可用).Status = %v, want FAIL", bad.Status)
	}
}

func TestRunDispatchesAllKinds(t *testing.T) {
	proj := project.Project{
		PythonEnv: ".venv",
		Requirements: []project.Requirement{
			{Kind: "python", Version: "3.11", Source: "pyproject.toml"},
			{Kind: "docker", Source: "Dockerfile"},
			{Kind: "platformio", Source: "platformio.ini"},
		},
	}
	results := Run(proj, Tools{
		Python: domain.ToolVersion{Active: "3.12.0", Available: true},
		Docker: domain.ToolVersion{Active: "29.8.1", Available: true},
	})
	if len(results) != 3 {
		t.Fatalf("Run() = %d 条, want 3", len(results))
	}
	if results[0].Name != "Python" || results[0].Status != domain.StatusPass || results[0].Current != "3.12.0 (.venv)" {
		t.Errorf("results[0] = %#v", results[0])
	}
	if results[1].Name != "Docker" || results[1].Status != domain.StatusPass {
		t.Errorf("results[1] = %#v", results[1])
	}
	if results[2].Name != "PlatformIO" || results[2].Status != domain.StatusFail {
		t.Errorf("results[2] = %#v", results[2])
	}
}
