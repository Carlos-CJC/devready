package domain

// CheckStatus 是检查项的统一状态(见设计文档 §11、§13)。
type CheckStatus int

const (
	StatusPass CheckStatus = iota // 满足要求
	StatusFail                    // 明确不满足
	StatusWarn                    // 存在潜在问题,但不阻止运行
	StatusSkip                    // 无法或无需检查
)

// Symbol 返回状态的符号标记。
func (s CheckStatus) Symbol() string {
	switch s {
	case StatusPass:
		return "✓"
	case StatusFail:
		return "✗"
	case StatusWarn:
		return "⚠"
	default:
		return "○"
	}
}

// Label 返回状态的文字标记。
func (s CheckStatus) Label() string {
	switch s {
	case StatusPass:
		return "PASS"
	case StatusFail:
		return "FAIL"
	case StatusWarn:
		return "WARN"
	default:
		return "SKIP"
	}
}

// CheckResult 是单项检查的结构化结果。
type CheckResult struct {
	Name     string // 检查项名称,如 "Go"
	Status   CheckStatus
	Required string // 要求,如 ">=1.22.0"
	Current  string // 实测,如 "1.22.0"
	Message  string // 补充说明
}
