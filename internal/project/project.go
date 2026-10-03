// Package project 识别当前目录所属的项目,并解析其环境要求。
//
// 按设计文档 §12 的规则表推断:Go(go.mod)、Python(pyproject.toml /
// requirements.txt)、Docker(Dockerfile / compose)、PlatformIO(platformio.ini)。
package project

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Requirement 是项目对某项环境的要求。
type Requirement struct {
	Kind    string // 需求类型,如 "go"
	Name    string // 展示名,如 "Go"
	Version string // 最低版本,如 "1.22.0"
	Source  string // 需求来源文件,如 "go.mod"
}

// Project 描述一个被识别出的项目。
type Project struct {
	Root         string // 项目根目录
	Name         string // 项目名
	Types        []string
	Requirements []Requirement
	PythonEnv    string // 识别到的 Python 环境标记(.venv / poetry.lock / …),"" 表示未识别
}

// Detect 自 dir 向上逐级查找项目标识文件,返回最近的项目根。
// found=false 表示未识别到项目。
func Detect(dir string) (Project, bool) {
	for d := dir; ; {
		if p, ok := detectIn(d); ok {
			return p, true
		}
		parent := filepath.Dir(d)
		if parent == d { // 已到文件系统根
			return Project{}, false
		}
		d = parent
	}
}

// detectIn 检查单个目录是否为一个项目根,按 §12 规则表推断类型与需求。
func detectIn(dir string) (Project, bool) {
	var types []string
	var reqs []Requirement
	module := ""

	if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		types = append(types, "Go")
		var goVer string
		module, goVer = parseGoMod(string(data))
		if goVer != "" {
			reqs = append(reqs, Requirement{Kind: "go", Name: "Go", Version: goVer, Source: "go.mod"})
		}
	}

	if req, ok := pythonRequirement(dir); ok {
		types = append(types, "Python")
		reqs = append(reqs, req)
	}

	if fileExists(dir, "Dockerfile") {
		types = append(types, "Docker")
		reqs = append(reqs, Requirement{Kind: "docker", Name: "Docker", Source: "Dockerfile"})
	}
	if name := composeFile(dir); name != "" {
		types = append(types, "Docker Compose")
		reqs = append(reqs, Requirement{Kind: "docker-compose", Name: "Docker Compose", Source: name})
	}

	if fileExists(dir, "platformio.ini") {
		types = append(types, "PlatformIO")
		reqs = append(reqs, Requirement{Kind: "platformio", Name: "PlatformIO", Source: "platformio.ini"})
	}

	if len(types) == 0 {
		return Project{}, false
	}

	name := path.Base(module)
	if name == "" || name == "." || name == "/" {
		name = filepath.Base(dir)
	}

	return Project{
		Root:         dir,
		Name:         name,
		Types:        types,
		Requirements: reqs,
		PythonEnv:    detectPythonEnv(dir),
	}, true
}

// pythonRequirement 依据 pyproject.toml / requirements.txt 推断 Python 需求。
// pyproject 能给出 requires-python 版本;requirements.txt 只能说明"需要 Python"。
func pythonRequirement(dir string) (Requirement, bool) {
	if data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml")); err == nil {
		return Requirement{
			Kind: "python", Name: "Python", Version: parseRequiresPython(string(data)), Source: "pyproject.toml",
		}, true
	}
	if fileExists(dir, "requirements.txt") {
		return Requirement{Kind: "python", Name: "Python", Source: "requirements.txt"}, true
	}
	return Requirement{}, false
}

// composeFile 返回目录中 compose 文件的名称,未找到时返回空串。
func composeFile(dir string) string {
	for _, name := range []string{"compose.yml", "compose.yaml", "docker-compose.yml", "docker-compose.yaml"} {
		if fileExists(dir, name) {
			return name
		}
	}
	return ""
}

// detectPythonEnv 识别项目使用的 Python 环境管理器,返回命中的标记名。
func detectPythonEnv(dir string) string {
	for _, m := range []string{".venv", "poetry.lock", "uv.lock", ".python-version", "environment.yml"} {
		if fileExists(dir, m) {
			return m
		}
	}
	return ""
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// parseRequiresPython 从 pyproject.toml 的 requires-python 取最低版本。
// 形如 requires-python = ">=3.11" -> "3.11";取不到返回空串。
func parseRequiresPython(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		rest, ok := strings.CutPrefix(line, "requires-python")
		if !ok {
			continue
		}
		rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest), "="))
		return leadingVersion(strings.Trim(rest, `"'`))
	}
	return ""
}

// leadingVersion 截取版本约束中从第一个数字起的版本号。
// 如 ">=3.11" -> "3.11";"~=3.11.0" -> "3.11.0";"==3.12.*" -> "3.12"。
func leadingVersion(s string) string {
	i := strings.IndexFunc(s, func(r rune) bool { return r >= '0' && r <= '9' })
	if i < 0 {
		return ""
	}
	s = s[i:]
	if end := strings.IndexFunc(s, func(r rune) bool {
		return !(r >= '0' && r <= '9' || r == '.')
	}); end >= 0 {
		s = s[:end]
	}
	return strings.TrimSuffix(s, ".")
}

// parseGoMod 从 go.mod 内容中取出 module 路径与 go 指令版本。
func parseGoMod(content string) (module, goVersion string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if rest, ok := strings.CutPrefix(line, "module "); ok {
			module = strings.TrimSpace(rest)
		}
		if rest, ok := strings.CutPrefix(line, "go "); ok {
			goVersion = strings.TrimSpace(rest)
		}
	}
	return module, goVersion
}
