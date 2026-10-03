package environment

import (
	"context"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

var platformioSpec = spec{
	name:        "PlatformIO",
	command:     "pio",
	plugin:      "platformio",
	versionArgs: []string{"--version"},
	parse:       parsePlatformIOVersion,
}

// DetectPlatformIO 侦测 PlatformIO Core 的路径与版本。
func DetectPlatformIO(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, platformioSpec, lookPath("pio"))
}

// parsePlatformIOVersion 从 `pio --version` 取版本,
// 如 "PlatformIO Core, version 6.1.15" -> "6.1.15"。
func parsePlatformIOVersion(out string) string {
	f := strings.Fields(out)
	if len(f) >= 4 && f[0] == "PlatformIO" {
		return f[3]
	}
	return ""
}
