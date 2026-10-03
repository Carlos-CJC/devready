package environment

import (
	"context"
	"strings"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

var latexSpec = spec{
	name:        "LaTeX",
	command:     "xelatex",
	plugin:      "latex",
	versionArgs: []string{"--version"},
	parse:       parseLatexEngine,
}

// DetectLaTeX 侦测 LaTeX 排版引擎。以 xelatex 为探测入口(§6 展示为 "XeLaTeX")。
func DetectLaTeX(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, latexSpec, lookPath("xelatex"))
}

// parseLatexEngine 从 `xelatex --version` 首行识别引擎名,
// 如 "XeTeX 3.141592653-2.6-0.999998 (TeX Live 2026)" -> "XeLaTeX"。
func parseLatexEngine(out string) string {
	if strings.HasPrefix(out, "XeTeX") {
		return "XeLaTeX"
	}
	return ""
}
