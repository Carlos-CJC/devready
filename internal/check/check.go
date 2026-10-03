// Package check 依据项目需求检查当前环境,产出结构化结果。
//
// 对应设计文档 §10、§11、§13:逐项检查,状态取 PASS/FAIL/WARN/SKIP 四态。
package check

import (
	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/project"
	"github.com/Carlos-CJC/devscope/internal/version"
)

// Tools 是检查所需的工具状态快照。
type Tools struct {
	Go domain.ToolVersion
}

// Run 对项目的全部需求依次检查,返回结果列表。
func Run(proj project.Project, tools Tools) []domain.CheckResult {
	var results []domain.CheckResult
	for _, r := range proj.Requirements {
		switch r.Kind {
		case "go":
			results = append(results, GoVersion(tools.Go.Active, r.Version))
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
