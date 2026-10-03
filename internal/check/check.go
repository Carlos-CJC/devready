// Package check 依据项目需求检查当前环境,产出结构化结果。
//
// 对应设计文档 §10、§11、§13:逐项检查,状态取 PASS/FAIL/WARN/SKIP 四态。
package check

import (
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/project"
	"github.com/Carlos-CJC/devscope/internal/version"
)

// Tools 是检查所需的工具状态快照。
type Tools struct {
	Go         domain.ToolVersion
	Python     domain.ToolVersion
	Docker     domain.ToolVersion
	Compose    domain.ToolVersion
	LaTeX      domain.ToolVersion
	PlatformIO domain.ToolVersion
}

// Run 对项目的全部需求依次检查,返回结果列表。
func Run(proj project.Project, tools Tools) []domain.CheckResult {
	var results []domain.CheckResult
	for _, r := range proj.Requirements {
		switch r.Kind {
		case "go":
			results = append(results, GoVersion(tools.Go.Active, r.Version))
		case "python":
			results = append(results, PythonVersion(tools.Python, r.Version, proj.PythonEnv))
		case "docker":
			results = append(results, Presence("Docker", tools.Docker))
		case "docker-compose":
			results = append(results, Presence("Docker Compose", tools.Compose))
		case "platformio":
			results = append(results, Presence("PlatformIO", tools.PlatformIO))
		case "latex":
			results = append(results, Presence("LaTeX", tools.LaTeX))
		}
	}
	return results
}

// GoVersion 检查生效的 Go 版本是否满足最低要求。
// required 为最低版本(如 "1.22.0");active 为当前生效版本;空值分别表示
// "项目未声明要求"与"Go 不可用"。
func GoVersion(active, required string) domain.CheckResult {
	res := domain.CheckResult{Name: "Go", Current: active}
	if required != "" {
		res.Required = ">=" + required
	}
	switch {
	case required == "":
		res.Status = domain.StatusSkip
		res.Message = "项目未声明 Go 版本要求"
	case active == "":
		res.Status = domain.StatusFail
		res.Message = "Go 不可用或未激活"
	case version.AtLeast(active, required):
		res.Status = domain.StatusPass
	default:
		res.Status = domain.StatusFail
		res.Message = "低于要求的 " + required
	}
	return res
}

// PythonVersion 检查生效的 Python 版本是否满足项目要求。
// required 为空表示项目只声明"需要 Python"而未给出版本(§12 → WARN);
// env 为项目识别到的环境标记(如 ".venv"),附加在 Current 上展示。
func PythonVersion(tool domain.ToolVersion, required, env string) domain.CheckResult {
	res := domain.CheckResult{Name: "Python", Current: pythonCurrent(tool.Active, env)}
	if required != "" {
		res.Required = ">=" + required
	}
	switch {
	case tool.Active == "":
		res.Status = domain.StatusFail
		res.Message = "Python 不可用或未激活"
	case required == "":
		res.Status = domain.StatusWarn
		res.Message = "项目未声明版本,无法校验"
	case version.AtLeast(tool.Active, required):
		res.Status = domain.StatusPass
	default:
		res.Status = domain.StatusFail
		res.Message = "低于要求的 " + required
		if len(tool.Installed) > 0 {
			res.Message += ";已装 " + strings.Join(tool.Installed, "/") + " 但未激活"
		}
	}
	return res
}

// Presence 检查某工具是否可用:可用即 PASS 并展示版本,否则 FAIL。
func Presence(name string, tool domain.ToolVersion) domain.CheckResult {
	res := domain.CheckResult{Name: name, Current: tool.Active}
	if tool.Available {
		res.Status = domain.StatusPass
	} else {
		res.Status = domain.StatusFail
		res.Message = "未安装或不可用"
	}
	return res
}

// pythonCurrent 把 Python 版本与环境标记合并为展示串,如 "3.12.0 (.venv)"。
func pythonCurrent(active, env string) string {
	if active == "" {
		return ""
	}
	if env != "" {
		return active + " (" + env + ")"
	}
	return active
}

// Tally 汇总各状态的数量。
func Tally(results []domain.CheckResult) (passed, failed, warned, skipped int) {
	for _, r := range results {
		switch r.Status {
		case domain.StatusPass:
			passed++
		case domain.StatusFail:
			failed++
		case domain.StatusWarn:
			warned++
		default:
			skipped++
		}
	}
	return
}
