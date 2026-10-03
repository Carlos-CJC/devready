// Package versionmanager 侦测工具由哪个版本管理器管理,并解析其版本状态。
//
// 目前仅支持 asdf(其 Go 重写版 0.19 使用 `asdf set` 等新语法)。
// 解析函数均为纯函数,便于单测;外部命令通过 Runner 注入。
package versionmanager

import (
	"context"
	"os/exec"
	"strings"
)

// ManagerASDF 是 asdf 的标识。
const ManagerASDF = "asdf"

// PluginGo 是 Go 在 asdf 中的插件名。
const PluginGo = "golang"

// 激活版本以 "*" 前缀标记;未设置版本时版本列显示为 "______"。
const (
	activeMarker = "*"
	unsetMarker  = "______"
	shimsMarker  = "/.asdf/shims/"
)

// Runner 执行外部命令并返回其标准输出。
type Runner func(ctx context.Context, name string, args ...string) (string, error)

// ExecRunner 是 Runner 的默认实现。
func ExecRunner(ctx context.Context, name string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	return string(out), err
}

// IsASDFShim 判断命令路径是否由 asdf 的 shim 提供。
func IsASDFShim(path string) bool {
	return path != "" && strings.Contains(path, shimsMarker)
}

// ParseList 解析 `asdf list <plugin>` 的输出,返回已安装版本。
// 输出形如:
//
//	 *1.22.0
//	  1.20.0
func ParseList(out string) []string {
	var versions []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, activeMarker))
		if line == "" {
			continue
		}
		versions = append(versions, line)
	}
	return versions
}

// ParseCurrent 解析 `asdf current <plugin>` 输出的数据行,返回激活版本与来源文件。
// 未设置时版本列显示为 "______",此时返回 ("", "")。
// 输出形如:
//
//	Name            Version         Source              Installed
//	golang          1.22.0          /path/.tool-versions true
func ParseCurrent(out string) (version, source string) {
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] == "Name" {
			continue
		}
		if fields[1] == unsetMarker {
			return "", ""
		}
		if len(fields) >= 3 && fields[2] != unsetMarker {
			return fields[1], fields[2]
		}
		return fields[1], ""
	}
	return "", ""
}

// ParseToolVersions 从 .tool-versions 内容中取出指定插件的版本。
// 每行格式为 "<plugin> <version>",支持以 # 开头的注释。
func ParseToolVersions(content, plugin string) string {
	for _, line := range strings.Split(content, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == plugin {
			return fields[1]
		}
	}
	return ""
}

// Installed 查询 asdf 中某插件已安装的版本。
func Installed(ctx context.Context, run Runner, plugin string) ([]string, error) {
	out, err := run(ctx, "asdf", "list", plugin)
	if err != nil {
		return nil, err
	}
	return ParseList(out), nil
}

// ActiveVersion 查询 asdf 中某插件的当前激活版本及其来源文件。
func ActiveVersion(ctx context.Context, run Runner, plugin string) (version, source string, err error) {
	out, err := run(ctx, "asdf", "current", plugin)
	if err != nil {
		return "", "", err
	}
	version, source = ParseCurrent(out)
	return version, source, nil
}
