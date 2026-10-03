package environment

import (
	"context"
	"strings"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

var pythonSpec = spec{
	name:        "Python",
	command:     "python3",
	plugin:      "python",
	versionArgs: []string{"--version"},
	parse:       parsePythonVersion,
}

// DetectPython 侦测 Python 解释器的路径、版本管理器与当前生效版本。
// 以 python3 为准(§6:Machine Scope 只回答"有没有 Python、默认版本是多少",
// 项目内的虚拟环境/poetry/uv 由 Project Scope 判断)。
func DetectPython(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, pythonSpec, lookPath("python3"))
}

// parsePythonVersion 从 `python --version` 输出取版本,如 "Python 3.12.0" -> "3.12.0"。
func parsePythonVersion(out string) string {
	f := strings.Fields(out)
	if len(f) >= 2 && f[0] == "Python" {
		return f[1]
	}
	return ""
}
