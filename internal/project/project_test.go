package project

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseGoMod(t *testing.T) {
	cases := []struct {
		name       string
		content    string
		wantModule string
		wantGo     string
	}{
		{"基础", "module github.com/cjc/devscope\n\ngo 1.22.0\n", "github.com/cjc/devscope", "1.22.0"},
		{"行内注释", "module x\n\ngo 1.22 // pinned\n", "x", "1.22"},
		{"无 go 指令", "module x\n", "x", ""},
		{"空内容", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, g := parseGoMod(c.content)
			if m != c.wantModule || g != c.wantGo {
				t.Errorf("parseGoMod() = (%q, %q), want (%q, %q)", m, g, c.wantModule, c.wantGo)
			}
		})
	}
}

func TestDetectInGoMod(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module github.com/cjc/devscope\n\ngo 1.22.0\n")

	p, ok := Detect(dir)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	if p.Name != "devscope" {
		t.Errorf("Name = %q, want devscope", p.Name)
	}
	if len(p.Types) != 1 || p.Types[0] != "Go" {
		t.Errorf("Types = %#v, want [Go]", p.Types)
	}
	if len(p.Requirements) != 1 {
		t.Fatalf("Requirements = %#v, want 1 条", p.Requirements)
	}
	r := p.Requirements[0]
	if r.Kind != "go" || r.Version != "1.22.0" || r.Source != "go.mod" {
		t.Errorf("Requirement = %#v, want {go Go 1.22.0 go.mod}", r)
	}
}

func TestDetectWalksUp(t *testing.T) {
	root := t.TempDir()
	write(t, root, "go.mod", "module example/root\n\ngo 1.21.0\n")
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	p, ok := Detect(sub)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	if p.Root != root {
		t.Errorf("Root = %q, want %q", p.Root, root)
	}
}

func TestDetectNoProject(t *testing.T) {
	if _, ok := Detect(t.TempDir()); ok {
		t.Error("Detect() found=true, want false")
	}
}
