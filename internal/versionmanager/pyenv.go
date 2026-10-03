package versionmanager

import (
	"context"
	"strings"
)

// ManagerPyenv 是 pyenv 的标识。
const ManagerPyenv = "pyenv"

const pyenvShimsMarker = "/.pyenv/shims/"

// IsPyenvShim 判断命令路径是否由 pyenv 的 shim 提供。
func IsPyenvShim(path string) bool {
	return path != "" && strings.Contains(path, pyenvShimsMarker)
}

// ParsePyenvVersion 解析 `pyenv version` 输出,返回激活版本与来源文件。
// 输出形如 "3.12.0 (set by /Users/me/.python-version)";
// 回落到系统 Python 时为 "system (set by /Users/me/.pyenv/version)",
// 此时视为未激活具体版本,版本返回空串(对应 §2.5:不把系统 Python 冒充成 pyenv 版本)。
func ParsePyenvVersion(out string) (version, source string) {
	out = strings.TrimSpace(out)
	if out == "" {
		return "", ""
	}
	version = out
	if i := strings.IndexByte(out, '('); i >= 0 {
		version = strings.TrimSpace(out[:i])
		rest := strings.TrimSpace(out[i+1:])
		if after, ok := strings.CutPrefix(rest, "set by "); ok {
			source = strings.TrimSpace(strings.TrimSuffix(after, ")"))
		}
	}
	if version == "system" {
		return "", source
	}
	return version, source
}

// ParsePyenvVersions 解析 `pyenv versions --bare` 输出,每行一个版本。
func ParsePyenvVersions(out string) []string {
	var versions []string
	for _, line := range strings.Split(out, "\n") {
		if v := strings.TrimSpace(line); v != "" {
			versions = append(versions, v)
		}
	}
	return versions
}

// PyenvInstalled 查询 pyenv 已安装的 Python 版本。
func PyenvInstalled(ctx context.Context, run Runner) ([]string, error) {
	out, err := run(ctx, "pyenv", "versions", "--bare")
	if err != nil {
		return nil, err
	}
	return ParsePyenvVersions(out), nil
}

// PyenvActive 查询 pyenv 当前激活的版本及来源文件;回落到系统 Python 时版本为空串。
func PyenvActive(ctx context.Context, run Runner) (version, source string, err error) {
	out, err := run(ctx, "pyenv", "version")
	if err != nil {
		return "", "", err
	}
	version, source = ParsePyenvVersion(out)
	return version, source, nil
}
