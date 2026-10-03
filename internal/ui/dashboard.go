package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/environment"
	"github.com/Carlos-CJC/devready/internal/project"
	"github.com/Carlos-CJC/devready/internal/services"
	"github.com/Carlos-CJC/devready/internal/system"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

// Snapshot 是一次 Dashboard 采集的完整快照,是渲染层的唯一输入(见设计文档 §16)。
type Snapshot struct {
	Project  *project.Project
	System   domain.SystemStatus
	Tools    []domain.ToolVersion
	Services []domain.ServiceInfo
	PortsErr error
}

// Collect 采集一次 Dashboard 所需的全部数据。
func Collect(ctx context.Context, dir string) Snapshot {
	run := versionmanager.ExecRunner
	snap := Snapshot{System: system.Collect(ctx, run)}

	if p, ok := project.Detect(dir); ok {
		snap.Project = &p
	}
	snap.Tools = []domain.ToolVersion{
		environment.DetectGit(ctx, run),
		environment.DetectGo(ctx, run),
		environment.DetectPython(ctx, run),
		environment.DetectDocker(ctx, run),
		environment.DetectDockerCompose(ctx, run),
		environment.DetectLaTeX(ctx, run),
		environment.DetectPlatformIO(ctx, run),
	}

	ports, err := services.ListeningPorts(ctx, run)
	snap.PortsErr = err
	snap.Services = services.Detect(ports)

	return snap
}

// RunDashboard 是 `devscope` 的入口:终端里跑实时刷新 TUI,非终端(管道/重定向)降级为纯文本。
func RunDashboard(ctx context.Context, dir string) error {
	snap := Collect(ctx, dir)

	if !isTerminal(os.Stdout) {
		fmt.Print(Render(snap))
		return nil
	}

	p := tea.NewProgram(newModel(ctx, dir, snap), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func isTerminal(f *os.File) bool { return term.IsTerminal(int(f.Fd())) }

// Render 把快照渲染为完整面板文本。Lip Gloss 在非终端环境自动去色,故此处无需分支。
func Render(snap Snapshot) string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("DevScope"))
	if snap.System.Hostname != "" {
		b.WriteString("  " + snap.System.Hostname)
	}
	b.WriteString("\n\n")

	if p := snap.Project; p != nil {
		b.WriteString(section("PROJECT") + "\n")
		b.WriteString(fieldLine("name", p.Name))
		b.WriteString(fieldLine("type", strings.Join(p.Types, " + ")))
		b.WriteString("\n")
	}

	renderSystem(&b, snap.System)
	renderDevelopment(&b, snap.Tools)
	renderServices(&b, snap.Services, snap.PortsErr)

	return b.String()
}

func section(s string) string { return styleSection.Render(s) }

func fieldLine(label, value string) string {
	pad := labelWidth - displayWidth(label)
	if pad < 0 {
		pad = 0
	}
	return fmt.Sprintf("  %s%s  %s\n", styleLabel.Render(label), strings.Repeat(" ", pad), value)
}

func renderSystem(b *strings.Builder, st domain.SystemStatus) {
	b.WriteString(section("SYSTEM") + "\n\n")

	if c, ok := st.CPU.Value, st.CPU.Status == domain.MetricOK; ok {
		line := "  " + styleLabel.Render(pad("CPU")) + c.Model
		if c.Cores > 0 {
			line += fmt.Sprintf("  %d 核", c.Cores)
		}
		fmt.Fprintf(b, "%s\n", line)
		fmt.Fprintf(b, "  %s%s  %s\n", strings.Repeat(" ", sysLabelWidth), coloredBar(c.Utilization, 24), pctLabel(c.Utilization))
		fmt.Fprintf(b, "  %s%.2f  %.2f  %.2f\n", styleLabel.Render(pad("LOAD")), c.Load1, c.Load5, c.Load15)
	} else {
		fmt.Fprintf(b, "  %s%s\n", styleLabel.Render(pad("CPU")), styleDim.Render(metricNote(st.CPU.Status)))
	}
	b.WriteString("\n")

	if m, ok := st.Memory.Value, st.Memory.Status == domain.MetricOK; ok {
		fmt.Fprintf(b, "  %s%s / %s\n", styleLabel.Render(pad("MEMORY")), humanBytes(m.Used), humanBytes(m.Total))
		fmt.Fprintf(b, "  %s%s  %s\n", strings.Repeat(" ", sysLabelWidth), coloredBar(m.UsedPercent(), 24), pctLabel(m.UsedPercent()))
	} else {
		fmt.Fprintf(b, "  %s%s\n", styleLabel.Render(pad("MEMORY")), styleDim.Render(metricNote(st.Memory.Status)))
	}
	b.WriteString("\n")

	b.WriteString("  " + styleLabel.Render("STORAGE") + "\n")
	if disks, ok := st.Disks.Value, st.Disks.Status == domain.MetricOK; ok && len(disks) > 0 {
		for _, d := range disks {
			fmt.Fprintf(b, "  %-20s %s / %s   %s  %s\n",
				diskLabel(d.MountPoint), humanBytes(d.Used), humanBytes(d.Total),
				coloredBar(d.UsedPercent(), 13), pctLabel(d.UsedPercent()))
		}
	} else {
		b.WriteString("  " + styleDim.Render(metricNote(st.Disks.Status)) + "\n")
	}
	b.WriteString("\n")

	b.WriteString("  " + styleLabel.Render("NETWORK") + "\n")
	switch {
	case st.Network.Status != domain.MetricOK:
		b.WriteString("  " + styleDim.Render(metricNote(st.Network.Status)) + "\n")
	case len(st.Network.Value) == 0:
		b.WriteString("  " + styleDim.Render("—  (无活动接口)") + "\n")
	default:
		nets := st.Network.Value
		const maxNet = 4
		for i, n := range nets {
			if i == maxNet {
				b.WriteString("  " + styleDim.Render(fmt.Sprintf("…  另有 %d 个接口", len(nets)-maxNet)) + "\n")
				break
			}
			fmt.Fprintf(b, "  %-20s ↓ %s   ↑ %s\n", n.Name, humanRate(n.RxBps), humanRate(n.TxBps))
		}
	}
	b.WriteString("\n")
}

func renderDevelopment(b *strings.Builder, tools []domain.ToolVersion) {
	b.WriteString(section("DEVELOPMENT") + "\n")
	for _, t := range tools {
		if t.Path == "" && t.Active == "" {
			continue // 未安装:Machine Scope 只回答"有没有装"(§6)
		}
		mark, style := "○", styleDim
		if t.Active != "" {
			mark, style = "✓", styleGood
		}
		fmt.Fprintf(b, "  %s %-15s %s\n", style.Render(mark), t.Name, versionLabel(t))
	}
	b.WriteString("\n")
}

func renderServices(b *strings.Builder, svcs []domain.ServiceInfo, portsErr error) {
	b.WriteString(section("SERVICES") + "\n")
	for _, s := range svcs {
		detail := s.Detail
		if portsErr != nil && s.Name == "sshd" {
			detail = "无法读取端口" // 读不到端口时如实标注,不默认"未运行"(§2.5)
		}
		mark, style := "○", styleDim
		if s.Running {
			mark, style = "●", styleGood
		}
		fmt.Fprintf(b, "  %s %-10s %s\n", style.Render(mark), s.Name, detail)
	}
}
