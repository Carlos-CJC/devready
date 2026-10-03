// Package ui 负责把采集到的数据渲染成文本 / TUI。这是唯一依赖 Bubble Tea、Lip Gloss 的层。
//
// 样式通过 Lip Gloss 表达,它在非终端环境会自动降级为无色输出,因此纯文本降级无需另写一套渲染。
package ui

import (
	"fmt"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/charmbracelet/lipgloss"
)

const (
	labelWidth    = 12 // PROJECT 等块的标签列宽
	sysLabelWidth = 8  // SYSTEM 块子标签(CPU / LOAD / MEMORY)的列宽
)

var (
	styleTitle   = lipgloss.NewStyle().Bold(true)
	styleSection = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleLabel   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleGood    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleBad     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

// Field 渲染一行 "标签  值",标签按显示宽度对齐(CJK 占 2 列)。
func Field(label, value string) {
	pad := labelWidth - displayWidth(label)
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("  %s%s  %s\n", label, strings.Repeat(" ", pad), value)
}

// Dash 把空串替换为占位符。
func Dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// displayWidth 估算字符串的终端显示宽度,全角/CJK 字符按 2 列计。
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x2E80 {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func pad(s string) string {
	if n := sysLabelWidth - displayWidth(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

// metricNote 把采集状态翻译为降级提示(对应 §2.5:取不到就说取不到)。
func metricNote(s domain.MetricStatus) string {
	if s == domain.MetricUnavailable {
		return "—  (本平台不支持)"
	}
	return "—  (采集失败)"
}

// diskLabel 给挂载点一个简短标签:根卷显示为 System,其余显示挂载路径。
func diskLabel(mount string) string {
	if mount == "/" {
		return "System"
	}
	return mount
}

// bar 渲染进度条;附带的高/低占用配色由 styles 决定。
func bar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct/100*float64(width) + 0.5)
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// coloredBar 在 bar 基础上按占用率着色(绿 / 黄 / 红)。
func coloredBar(pct float64, width int) string {
	s := bar(pct, width)
	switch {
	case pct >= 90:
		return styleBad.Render(s)
	case pct >= 70:
		return styleWarn.Render(s)
	default:
		return styleGood.Render(s)
	}
}

// pctLabel 渲染百分比;极小的非零值显示为 "<1%"。
func pctLabel(pct float64) string {
	if pct > 0 && pct < 1 {
		return "<1%"
	}
	return fmt.Sprintf("%.0f%%", pct)
}

// humanBytes 把字节数格式化为便于阅读的单位。
func humanBytes(b uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
		tb = gb * 1024
	)
	f := float64(b)
	switch {
	case b >= tb:
		return fmt.Sprintf("%.1f TB", f/tb)
	case b >= gb:
		if f/gb < 100 {
			return fmt.Sprintf("%.1f GB", f/gb)
		}
		return fmt.Sprintf("%.0f GB", f/gb)
	case b >= mb:
		return fmt.Sprintf("%.0f MB", f/mb)
	case b >= kb:
		return fmt.Sprintf("%.0f KB", f/kb)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// humanRate 把字节/秒格式化为便于阅读的速率。
func humanRate(bps uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
	)
	f := float64(bps)
	switch {
	case bps >= mb:
		return fmt.Sprintf("%.1f MB/s", f/mb)
	case bps >= kb:
		return fmt.Sprintf("%.1f KB/s", f/kb)
	default:
		return fmt.Sprintf("%d B/s", bps)
	}
}

// versionLabel 渲染工具版本,附带版本管理器来源。
func versionLabel(t domain.ToolVersion) string {
	if t.Active == "" {
		return "—  (不可用)"
	}
	label := t.Active
	if t.Manager != "" && t.Manager != "unknown" {
		label += "  (" + t.Manager + ")"
	}
	return label
}

