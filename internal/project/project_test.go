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
		{"基础", "module github.com/Carlos-CJC/devscope\n\ngo 1.22.0\n", "github.com/Carlos-CJC/devscope", "1.22.0"},
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
	write(t, dir, "go.mod", "module github.com/Carlos-CJC/devscope\n\ngo 1.22.0\n")

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

func TestParseRequiresPython(t *testing.T) {
	cases := []struct{ in, want string }{
		{"[project]\nrequires-python = \">=3.11\"\n", "3.11"},
		{"requires-python = \"~=3.11.0\"", "3.11.0"},
		{"requires-python=\"==3.12.*\"", "3.12"},
		{"requires-python = '>=3.9'  # 注释", "3.9"},
		{"[project]\nname = \"x\"", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseRequiresPython(c.in); got != c.want {
			t.Errorf("parseRequiresPython(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectInPyproject(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pyproject.toml", "[project]\nrequires-python = \">=3.11\"\n")

	p, ok := Detect(dir)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	if len(p.Types) != 1 || p.Types[0] != "Python" {
		t.Errorf("Types = %#v, want [Python]", p.Types)
	}
	if len(p.Requirements) != 1 {
		t.Fatalf("Requirements = %#v, want 1 条", p.Requirements)
	}
	r := p.Requirements[0]
	if r.Kind != "python" || r.Version != "3.11" || r.Source != "pyproject.toml" {
		t.Errorf("Requirement = %#v, want {python Python 3.11 pyproject.toml}", r)
	}
}

func TestDetectInRequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "requirements.txt", "requests\n")

	p, ok := Detect(dir)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	r := p.Requirements[0]
	if r.Kind != "python" || r.Version != "" || r.Source != "requirements.txt" {
		t.Errorf("Requirement = %#v, want {python Python \"\" requirements.txt}", r)
	}
}

func TestDetectInDocker(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Dockerfile", "FROM alpine\n")
	write(t, dir, "compose.yml", "services: {}\n")

	p, ok := Detect(dir)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	kinds := map[string]string{}
	for _, r := range p.Requirements {
		kinds[r.Kind] = r.Source
	}
	if kinds["docker"] != "Dockerfile" || kinds["docker-compose"] != "compose.yml" {
		t.Errorf("Requirements = %#v, want docker=Dockerfile & docker-compose=compose.yml", p.Requirements)
	}
}

func TestDetectInPlatformIO(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "platformio.ini", "[env:esp32]\n")

	p, ok := Detect(dir)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	r := p.Requirements[0]
	if r.Kind != "platformio" || r.Source != "platformio.ini" {
		t.Errorf("Requirement = %#v, want platformio from platformio.ini", r)
	}
}

func TestDetectPythonEnv(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "requirements.txt", "")
	if err := os.MkdirAll(filepath.Join(dir, ".venv"), 0o755); err != nil {
		t.Fatal(err)
	}

	p, ok := Detect(dir)
	if !ok {
		t.Fatal("Detect() found=false, want true")
	}
	if p.PythonEnv != ".venv" {
		t.Errorf("PythonEnv = %q, want .venv", p.PythonEnv)
	}
}
