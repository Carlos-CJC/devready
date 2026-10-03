package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Carlos-CJC/devready/internal/check"
	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/environment"
	"github.com/Carlos-CJC/devready/internal/project"
	"github.com/Carlos-CJC/devready/internal/services"
	"github.com/Carlos-CJC/devready/internal/ui"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
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
		if err := ui.RunDashboard(ctx, dir); err != nil {
			fmt.Fprintln(os.Stderr, "运行面板失败:"+err.Error())
			os.Exit(1)
		}
	case "info":
		runInfo(dir)
	case "check":
		runCheck(ctx, dir)
	case "port":
		runPort(ctx)
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
  devscope          实时面板(终端内每 2 秒刷新,按 q 退出)
  devscope info     查看项目环境要求
  devscope check    检查当前环境是否满足项目要求
  devscope port     查看监听端口
`)
}

// runInfo 对应设计文档 §9:只展示项目要求,不执行检查。
func runInfo(dir string) {
	proj, ok := project.Detect(dir)
	if !ok {
		fmt.Println("当前目录不是可识别的项目。")
		return
	}

	fmt.Println("Project Information")
	fmt.Println()
	ui.Field("project", proj.Name)
	ui.Field("type", strings.Join(proj.Types, " + "))
	fmt.Println()

	fmt.Println("REQUIREMENTS")
	if len(proj.Requirements) == 0 {
		fmt.Println("  (无)")
		return
	}
	for _, r := range proj.Requirements {
		req := "required"
		if r.Version != "" {
			req = ">= " + r.Version
		}
		fmt.Printf("  %-16s %s   (%s)\n", r.Name, req, r.Source)
	}
}

// runCheck 对应设计文档 §10、§12:逐项检查并输出汇总表。
func runCheck(ctx context.Context, dir string) {
	proj, ok := project.Detect(dir)
	if !ok {
		fmt.Println("当前目录不是可识别的项目,无法检查。")
		return
	}

	results := check.Run(proj, collectTools(ctx, proj))

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

// collectTools 只侦测项目实际要求的工具,避免无关的外部命令开销。
func collectTools(ctx context.Context, proj project.Project) check.Tools {
	run := versionmanager.ExecRunner
	var t check.Tools
	for _, r := range proj.Requirements {
		switch r.Kind {
		case "go":
			t.Go = environment.DetectGo(ctx, run)
		case "python":
			t.Python = environment.DetectPython(ctx, run)
		case "docker":
			t.Docker = environment.DetectDocker(ctx, run)
		case "docker-compose":
			t.Compose = environment.DetectDockerCompose(ctx, run)
		case "latex":
			t.LaTeX = environment.DetectLaTeX(ctx, run)
		case "platformio":
			t.PlatformIO = environment.DetectPlatformIO(ctx, run)
		}
	}
	return t
}

func printSummary(results []domain.CheckResult) {
	fmt.Println("Check Summary")
	fmt.Println()
	fmt.Printf("  %-16s %-10s %s\n", "Requirement", "Status", "Current")
	for _, r := range results {
		fmt.Printf("  %-16s %s %-6s %s\n", r.Name, r.Status.Symbol(), r.Status.Label(), ui.Dash(r.Current))
	}
	fmt.Println()
	fmt.Println("  Result: " + tallyLine(check.Tally(results)))
}

// runPort 对应设计文档 §7:把监听端口翻译成"这台机器现在有哪些服务在工作"。
func runPort(ctx context.Context) {
	ports, err := services.ListeningPorts(ctx, versionmanager.ExecRunner)
	if err != nil {
		fmt.Println("无法列出监听端口:" + err.Error())
		return
	}

	fmt.Println("Listening Ports")
	fmt.Println()
	if len(ports) == 0 {
		fmt.Println("  (无监听端口)")
		return
	}
	for _, p := range ports {
		owner := p.Process
		if owner == "" {
			owner = "?"
		}
		line := fmt.Sprintf("  %-8s %s", fmt.Sprintf(":%d", p.Port), owner)
		if p.Docker {
			line += "  (Docker 映射)"
		}
		fmt.Println(line)
	}
	fmt.Println()
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
