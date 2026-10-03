// Package project 识别当前目录所属的项目,并解析其环境要求。
//
// 第一版仅支持 Go(go.mod);后续按设计文档 §12 的规则表扩展。
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

// detectIn 检查单个目录是否为一个项目根。
func detectIn(dir string) (Project, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return Project{}, false
	}
	module, goVer := parseGoMod(string(data))

	name := path.Base(module)
	if name == "" || name == "." || name == "/" {
		name = filepath.Base(dir)
	}

	p := Project{Root: dir, Name: name, Types: []string{"Go"}}
	if goVer != "" {
		p.Requirements = append(p.Requirements, Requirement{
			Kind: "go", Name: "Go", Version: goVer, Source: "go.mod",
		})
	}
	return p, true
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
