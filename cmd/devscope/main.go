package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/check"
	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/environment"
	"github.com/Carlos-CJC/devscope/internal/project"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

func main() {
	ctx := context.Background()
	dir, _ := os.Getwd()

	args := os.Args[1:]
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "":
		runDashboard(ctx, dir)
	case "info":
		runInfo(ctx, dir)
	case "check":
		runCheck(ctx, dir)
	case "port":
		fmt.Println("devscope port:尚未实现(见设计文档 §7)")
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令:%s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`DevScope

用法:
  devscope          查看机器 / 项目概况
  devscope info     查看项目环境要求
  devscope check    检查当前环境是否满足项目要求
  devscope port     查看监听端口(尚未实现)
`)
}

// runDashboard 对应设计文档 §3、§4、§6。当前仅实现 DEVELOPMENT 中的 Go。
func runDashboard(ctx context.Context, dir string) {
	goInfo := environment.DetectGo(ctx, versionmanager.ExecRunner)

	fmt.Println("DevScope")
	fmt.Println()

	if proj, ok := project.Detect(dir); ok {
		fmt.Println("PROJECT")
		field("name", proj.Name)
		field("type", strings.Join(proj.Types, " + "))
		fmt.Println()
	}

	fmt.Println("DEVELOPMENT")
	fmt.Printf("  %s Go  %s\n", toolMark(goInfo), versionLabel(goInfo))
}

// runInfo 对应设计文档 §9:只展示项目要求,不执行检查。
func runInfo(_ context.Context, dir string) {
	proj, ok := project.Detect(dir)
	if !ok {
		fmt.Println("当前目录不是可识别的项目。")
		return
	}

	fmt.Println("Project Information")
	fmt.Println()
	field("project", proj.Name)
	field("type", strings.Join(proj.Types, " + "))
	fmt.Println()

	fmt.Println("REQUIREMENTS")
	if len(proj.Requirements) == 0 {
		fmt.Println("  (无)")
		return
	}
	for _, r := range proj.Requirements {
		fmt.Printf("  %-10s >= %s   (%s)\n", r.Name, r.Version, r.Source)
	}
}

// runCheck 对应设计文档 §10、§12:逐项检查并输出汇总表。
func runCheck(ctx context.Context, dir string) {
	proj, ok := project.Detect(dir)
	if !ok {
		fmt.Println("当前目录不是可识别的项目,无法检查。")
		return
	}

	goInfo := environment.DetectGo(ctx, versionmanager.ExecRunner)
	results := check.Run(proj, check.Tools{Go: goInfo})

	fmt.Println("Checking project environment...")
	fmt.Println()
	for _, r := range results {
		line := r.Status.Symbol() + " " + r.Name
		if r.Current != "" {
			line += "  " + r.Current
		}
		if r.Message != "" {
			line += "  (" + r.Message + ")"
		}
		fmt.Println(line)
	}
	fmt.Println()

	printSummary(results)
}

func printSummary(results []domain.CheckResult) {
	fmt.Println("Check Summary")
	fmt.Println()
	fmt.Printf("  %-16s %-10s %s\n", "Requirement", "Status", "Current")
	for _, r := range results {
		fmt.Printf("  %-16s %s %-6s %s\n", r.Name, r.Status.Symbol(), r.Status.Label(), dash(r.Current))
	}
	fmt.Println()
	fmt.Println("  Result: " + tallyLine(check.Tally(results)))
}

func tallyLine(passed, failed, warned, skipped int) string {
	var parts []string
	if passed > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", passed))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if warned > 0 {
		parts = append(parts, fmt.Sprintf("%d warning", warned))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	if len(parts) == 0 {
		return "无检查项"
	}
	return strings.Join(parts, " · ")
}

const labelWidth = 12 // 以终端显示宽度计(中文按 2 列)

func field(label, value string) {
	pad := labelWidth - displayWidth(label)
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("  %s%s  %s\n", label, strings.Repeat(" ", pad), value)
}

// toolMark 以"是否解析出可用版本"判断工具是否就绪,避免单次探针抖动误判。
func toolMark(t domain.ToolVersion) string {
	if t.Active != "" {
		return "✓"
	}
	return "○"
}

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

func dash(s string) string {
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
