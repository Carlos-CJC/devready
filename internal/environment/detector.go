package environment

import (
	"context"
	"os/exec"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

// spec 描述如何侦测一个开发工具:命令名、asdf 插件名、版本参数与输出解析。
type spec struct {
	name        string   // 展示名,如 "Go"
	command     string   // 命令名,如 "go"
	plugin      string   // asdf 插件名,如 "golang"
	versionArgs []string // 取版本所用的参数,如 "--version"
	parse       func(string) string
}

// detect 是各工具通用的侦测流程:定位命令 → 识别版本管理器 → 读取生效版本。
// 命令不可执行时保持 Available=false、Active="",交由上层如实降级(§2.5)。
func detect(ctx context.Context, run versionmanager.Runner, sp spec, path string) domain.ToolVersion {
	info := domain.ToolVersion{Name: sp.name, Command: sp.command, Manager: "unknown", Path: path}

	switch {
	case versionmanager.IsASDFShim(path):
		info.Manager = versionmanager.ManagerASDF
		if versions, err := versionmanager.Installed(ctx, run, sp.plugin); err == nil {
			info.Installed = versions
		}
		if v, src, err := versionmanager.ActiveVersion(ctx, run, sp.plugin); err == nil {
			info.Active, info.Source = v, src
		}
	case versionmanager.IsPyenvShim(path):
		info.Manager = versionmanager.ManagerPyenv
		if versions, err := versionmanager.PyenvInstalled(ctx, run); err == nil {
			info.Installed = versions
		}
		if v, src, err := versionmanager.PyenvActive(ctx, run); err == nil {
			info.Active, info.Source = v, src
		}
		if info.Active == "" {
			// pyenv 已装但未激活具体版本,shim 回落到系统 Python,不应标为 pyenv。
			info.Manager = "system"
		}
	case path != "":
		info.Manager = "system"
	}

	if out, err := run(ctx, sp.command, sp.versionArgs...); err == nil {
		info.Available = true
		if info.Active == "" {
			info.Active = sp.parse(out)
		}
	}
	return info
}

// lookPath 解析命令的绝对路径,失败时返回空串。
func lookPath(command string) string {
	p, _ := exec.LookPath(command)
	return p
}
