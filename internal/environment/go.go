// Package environment 侦测开发工具的运行环境与版本管理状态。
// 项目对工具的要求由 project 包负责,本包只回答"这台机器上工具是什么状态"。
package environment

import (
	"context"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

var goSpec = spec{
	name:        "Go",
	command:     "go",
	plugin:      versionmanager.PluginGo,
	versionArgs: []string{"version"},
	parse:       parseGoVersion,
}

// DetectGo 侦测 Go 的安装位置、版本管理器与当前生效版本。
func DetectGo(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, goSpec, lookPath("go"))
}

// detectGoWith 是 DetectGo 的核心实现,path 由调用方解析以便测试。
func detectGoWith(ctx context.Context, run versionmanager.Runner, path string) domain.ToolVersion {
	return detect(ctx, run, goSpec, path)
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
