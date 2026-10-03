package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	case "r":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	default:
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestQuitKeys(t *testing.T) {
	m := newModel(context.Background(), ".", Snapshot{})
	for _, k := range []string{"q"} {
		_, cmd := m.Update(keyMsg(k))
		if !isQuit(cmd) {
			t.Fatalf("key %q did not quit", k)
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !isQuit(cmd) {
		t.Fatal("esc did not quit")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !isQuit(cmd) {
		t.Fatal("ctrl+c did not quit")
	}
}

func TestRefreshKeyReCollects(t *testing.T) {
	m := newModel(context.Background(), ".", Snapshot{})
	_, cmd := m.Update(keyMsg("r"))
	if cmd == nil {
		t.Fatal("r did not trigger a re-collect")
	}
}

func TestSnapshotMsgArmsNextTick(t *testing.T) {
	m := newModel(context.Background(), ".", Snapshot{})
	snap := Snapshot{}
	_, cmd := m.Update(snapshotMsg{snap: snap})
	if cmd == nil {
		t.Fatal("snapshotMsg did not re-arm the tick")
	}
}

func TestTickTriggersCollect(t *testing.T) {
	m := newModel(context.Background(), ".", Snapshot{})
	_, cmd := m.Update(tickMsg{})
	if cmd == nil {
		t.Fatal("tickMsg did not trigger a collect")
	}
}
