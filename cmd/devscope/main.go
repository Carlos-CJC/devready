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
	"github.com/Carlos-CJC/devscope/internal/services"
	"github.com/Carlos-CJC/devscope/internal/system"
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
  devscope          查看机器 / 项目概况
  devscope info     查看项目环境要求
  devscope check    检查当前环境是否满足项目要求
  devscope port     查看监听端口
`)
}

// runDashboard 对应设计文档 §3、§4、§5、§6。
func runDashboard(ctx context.Context, dir string) {
	st := system.Collect(ctx, versionmanager.ExecRunner)

	fmt.Print("DevScope")
	if st.Hostname != "" {
		fmt.Print("  " + st.Hostname)
	}
	fmt.Println()
	fmt.Println()

	if proj, ok := project.Detect(dir); ok {
		fmt.Println("PROJECT")
		field("name", proj.Name)
		field("type", strings.Join(proj.Types, " + "))
		fmt.Println()
	}

	renderSystem(st)

	goInfo := environment.DetectGo(ctx, versionmanager.ExecRunner)
	fmt.Println("DEVELOPMENT")
	fmt.Printf("  %s Go  %s\n", toolMark(goInfo), versionLabel(goInfo))
	fmt.Println()

	renderServices(ctx)
}

// renderServices 渲染 Machine Scope 的 SERVICES 区块(§7),数据来自监听端口与本机事实。
func renderServices(ctx context.Context) {
	fmt.Println("SERVICES")
	ports, err := services.ListeningPorts(ctx, versionmanager.ExecRunner)
	svcs := services.Detect(ports)
	if err != nil {
		// 读不到端口时,sshd 状态无从判断,如实标注而非默认"未运行"。
		for i := range svcs {
			if svcs[i].Name == "sshd" {
				svcs[i].Detail = "无法读取端口"
			}
		}
	}
	for _, s := range svcs {
		mark := "○"
		if s.Running {
			mark = "●"
		}
		fmt.Printf("  %s %-10s %s\n", mark, s.Name, s.Detail)
	}
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

// renderSystem 渲染 Machine Scope 的 SYSTEM 区块(§5)。采集不到的指标显示为 ○,不伪造。
func renderSystem(st domain.SystemStatus) {
	fmt.Println("SYSTEM")
	fmt.Println()

	if c, ok := st.CPU.Value, st.CPU.Status == domain.MetricOK; ok {
		line := "  " + pad("CPU") + c.Model
		if c.Cores > 0 {
			line += fmt.Sprintf("  %d 核", c.Cores)
		}
		fmt.Println(line)
		fmt.Printf("  %s%s  %s\n", strings.Repeat(" ", sysLabelWidth), bar(c.Utilization, 24), pctLabel(c.Utilization))
		fmt.Printf("  %s%.2f  %.2f  %.2f\n", pad("LOAD"), c.Load1, c.Load5, c.Load15)
	} else {
		fmt.Printf("  %s%s\n", pad("CPU"), metricNote(st.CPU.Status))
	}
	fmt.Println()

	if m, ok := st.Memory.Value, st.Memory.Status == domain.MetricOK; ok {
		fmt.Printf("  %s%s / %s\n", pad("MEMORY"), humanBytes(m.Used), humanBytes(m.Total))
		fmt.Printf("  %s%s  %s\n", strings.Repeat(" ", sysLabelWidth), bar(m.UsedPercent(), 24), pctLabel(m.UsedPercent()))
	} else {
		fmt.Printf("  %s%s\n", pad("MEMORY"), metricNote(st.Memory.Status))
	}
	fmt.Println()

	fmt.Println("  STORAGE")
	if disks, ok := st.Disks.Value, st.Disks.Status == domain.MetricOK; ok && len(disks) > 0 {
		for _, d := range disks {
			fmt.Printf("  %-20s %s / %s   %s  %s\n",
				diskLabel(d.MountPoint), humanBytes(d.Used), humanBytes(d.Total),
				bar(d.UsedPercent(), 13), pctLabel(d.UsedPercent()))
		}
	} else {
		fmt.Printf("  %s\n", metricNote(st.Disks.Status))
	}
	fmt.Println()

	fmt.Println("  NETWORK")
	switch {
	case st.Network.Status != domain.MetricOK:
		fmt.Printf("  %s\n", metricNote(st.Network.Status))
	case len(st.Network.Value) == 0:
		fmt.Println("  —  (无活动接口)")
	default:
		nets := st.Network.Value
		const maxNet = 4
		for i, n := range nets {
			if i == maxNet {
				fmt.Printf("  …  另有 %d 个接口\n", len(nets)-maxNet)
				break
			}
			fmt.Printf("  %-20s ↓ %s   ↑ %s\n", n.Name, humanRate(n.RxBps), humanRate(n.TxBps))
		}
	}
	fmt.Println()
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

const sysLabelWidth = 8 // SYSTEM 区块子标签(CPU / LOAD / MEMORY)的显示宽度

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

// bar 渲染一个宽 width 的进度条,如 "████░░░░"。
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
