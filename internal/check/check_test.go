package check

import (
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
