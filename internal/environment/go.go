// Package environment 侦测开发工具的运行环境与版本管理状态。
package environment

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cjc/devscope/internal/domain"
	"github.com/cjc/devscope/internal/versionmanager"
)

// GoInfo 是 Go 工具的完整侦测结果。
type GoInfo struct {
	domain.ToolVersion
	Required string // 项目要求的最低版本(取自 go.mod),"" 表示无要求
}

// DetectGo 侦测 Go 的安装、版本管理与当前项目的版本要求。
// dir 为项目目录,用于读取 go.mod;传 "" 表示只看机器级状态。
func DetectGo(ctx context.Context, run versionmanager.Runner, dir string) GoInfo {
	path, _ := exec.LookPath("go")
	return detectGoWith(ctx, run, path, dir)
}

// detectGoWith 是 DetectGo 的核心实现,path 由调用方解析以便测试。
func detectGoWith(ctx context.Context, run versionmanager.Runner, path, dir string) GoInfo {
	info := GoInfo{ToolVersion: domain.ToolVersion{
		Name:    "Go",
		Command: "go",
		Manager: "unknown",
		Path:    path,
	}}

	switch {
	case versionmanager.IsASDFShim(path):
		info.Manager = versionmanager.ManagerASDF
		if versions, err := versionmanager.Installed(ctx, run, versionmanager.PluginGo); err == nil {
			info.Installed = versions
		}
		if v, src, err := versionmanager.ActiveVersion(ctx, run, versionmanager.PluginGo); err == nil {
			info.Active, info.Source = v, src
		}
	case path != "":
		info.Manager = "system"
	}

	// 命令可实际执行时,读取运行时版本。
	if out, err := run(ctx, "go", "version"); err == nil {
		info.Available = true
		if info.Active == "" {
			info.Active = parseGoVersion(out)
		}
	}

	if req, err := readGoModRequirement(dir); err == nil {
		info.Required = req
	}
	return info
}

// parseGoVersion 从 `go version` 输出中取出语义版本。
// 例如 "go version go1.22.0 darwin/arm64" -> "1.22.0"。
func parseGoVersion(out string) string {
	for _, f := range strings.Fields(out) {
		if len(f) > 2 && f[0] == 'g' && f[1] == 'o' && f[2] >= '0' && f[2] <= '9' {
			return f[2:]
		}
	}
	return ""
}

// ParseGoMod 从 go.mod 内容中取出 `go` 指令声明的版本,如 "1.22.0"。
func ParseGoMod(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if rest, ok := strings.CutPrefix(line, "go "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// readGoModRequirement 读取目录下 go.mod 的版本要求。
func readGoModRequirement(dir string) (string, error) {
	if dir == "" {
		return "", os.ErrNotExist
	}
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}
	return ParseGoMod(string(data)), nil
}
