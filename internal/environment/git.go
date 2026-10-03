package environment

import (
	"context"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

var gitSpec = spec{
	name:        "Git",
	command:     "git",
	plugin:      versionmanager.PluginGit,
	versionArgs: []string{"--version"},
	parse:       parseGitVersion,
}

// DetectGit 侦测 Git 的安装位置、版本管理器与当前生效版本。
func DetectGit(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, gitSpec, lookPath("git"))
}

// parseGitVersion 从 `git --version` 输出取版本,如 "git version 2.51.0" -> "2.51.0"。
func parseGitVersion(out string) string {
	f := strings.Fields(out)
	if len(f) >= 3 && f[0] == "git" {
		return f[2]
	}
	return ""
}
