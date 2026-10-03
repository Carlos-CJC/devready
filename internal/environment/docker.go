package environment

import (
	"context"
	"strings"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

var dockerSpec = spec{
	name:        "Docker",
	command:     "docker",
	plugin:      "docker",
	versionArgs: []string{"--version"},
	parse:       parseDockerVersion,
}

// DetectDocker 侦测 Docker CLI 的安装位置与版本。
func DetectDocker(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, dockerSpec, lookPath("docker"))
}

// parseDockerVersion 从 `docker --version` 取版本。
// 例如 "Docker version 29.8.1, build 4a63305" -> "29.8.1"。
func parseDockerVersion(out string) string {
	f := strings.Fields(out)
	if len(f) >= 3 && f[0] == "Docker" {
		return strings.TrimSuffix(f[2], ",")
	}
	return ""
}

var composeSpec = spec{
	name:        "Docker Compose",
	command:     "docker",
	plugin:      "docker-compose",
	versionArgs: []string{"compose", "version"},
	parse:       parseComposeVersion,
}

// DetectDockerCompose 侦测 `docker compose` 子命令是否可用及其版本。
func DetectDockerCompose(ctx context.Context, run versionmanager.Runner) domain.ToolVersion {
	return detect(ctx, run, composeSpec, lookPath("docker"))
}

// parseComposeVersion 从 `docker compose version` 取版本,
// 例如 "Docker Compose version v2.24.0" -> "2.24.0"。
func parseComposeVersion(out string) string {
	f := strings.Fields(out)
	if len(f) >= 4 && f[0] == "Docker" && f[1] == "Compose" {
		return strings.TrimPrefix(f[3], "v")
	}
	return ""
}
