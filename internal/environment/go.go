// Package environment 侦测开发工具的运行环境与版本管理状态。
// 项目对工具的要求由 project 包负责,本包只回答"这台机器上工具是什么状态"。
package environment

import (
	"context"
	"os/exec"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

// DetectGo 侦测 Go 的安装位置、版本管理器与当前生效版本。
func DetectGo(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	path, _ := exec.LookPath("go")
	return detectGoWith(ctx, run, path)
}

// detectGoWith 是 DetectGo 的核心实现,path 由调用方解析以便测试。
func detectGoWith(ctx context.Context, run versionmanager.Runner, path string) domain.ToolVersion {
	info := domain.ToolVersion{Name: "Go", Command: "go", Manager: "unknown", Path: path}

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
