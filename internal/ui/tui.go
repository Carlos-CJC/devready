package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// refreshInterval 是面板自动重采的间隔(见设计文档 §4)。
const refreshInterval = 2 * time.Second

type model struct {
	ctx     context.Context
	dir     string
	snap    Snapshot
	updated time.Time
	ready   bool
}

type snapshotMsg struct {
	snap Snapshot
	at   time.Time
}

type tickMsg time.Time

func newModel(ctx context.Context, dir string, initial Snapshot) model {
	return model{ctx: ctx, dir: dir, snap: initial, ready: true, updated: time.Now()}
}

func (m model) Init() tea.Cmd { return tickCmd() }

// collectCmd 在后台协程采集,避免阻塞 UI(§16 并发与超时)。
func (m model) collectCmd() tea.Cmd {
	return func() tea.Msg {
		return snapshotMsg{snap: Collect(m.ctx, m.dir), at: time.Now()}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, m.collectCmd() // 手动立即刷新
		}
	case snapshotMsg:
		m.snap = msg.snap
		m.updated = msg.at
		m.ready = true
		return m, tickCmd() // 采集完成后才重新计时,避免采集慢于间隔时任务堆积
	case tickMsg:
		return m, m.collectCmd()
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "\n  " + styleDim.Render("DevReady  正在采集…") + "\n"
	}
	return Render(m.snap) + "\n" + m.footer() + "\n"
}

func (m model) footer() string {
	return styleDim.Render(fmt.Sprintf("  刷新于 %s · 每 %s 自动刷新 · 按 q 退出",
		m.updated.Format("15:04:05"), refreshInterval))
}
