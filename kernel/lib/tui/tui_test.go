package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func newSizedModel() *tuiModel {
	m := newTuiModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 23})
	return m
}

func topRow(vp viewport.Model) string {
	row, _, _ := strings.Cut(vp.View(), "\n")
	return strings.TrimRight(row, " ")
}

func TestLogPaneKeepsMostRecentEntries(t *testing.T) {
	p := &logPane{}
	for i := range maxPaneLines + 10 {
		p.append(fmt.Sprintf("line %d", i))
	}
	require.Len(t, p.lines, maxPaneLines)
	require.Equal(t, "line 10", p.lines[0])
	require.Equal(t, fmt.Sprintf("line %d", maxPaneLines+9), p.lines[len(p.lines)-1])
}

func TestLogLinesShowOnRefresh(t *testing.T) {
	m := newSizedModel()

	_, cmd := m.Update(logLineMsg{pane: PaneActions, text: "first"})
	require.NotNil(t, cmd, "the first line schedules a refresh")
	require.IsType(t, paneRefreshMsg{}, cmd())

	_, cmd = m.Update(logLineMsg{pane: PaneValidation, text: "second"})
	require.Nil(t, cmd, "a refresh is already pending")
	require.Empty(t, topRow(m.actions.vp))
	require.Empty(t, topRow(m.validation.vp))

	m.Update(paneRefreshMsg{})
	require.Equal(t, "first", topRow(m.actions.vp))
	require.Equal(t, "second", topRow(m.validation.vp))

	_, cmd = m.Update(logLineMsg{pane: PaneActions, text: "third"})
	require.NotNil(t, cmd, "a line after the refresh schedules another")
}

func TestRefreshFollowsTheBottom(t *testing.T) {
	m := newSizedModel()
	for i := range 50 {
		m.Update(logLineMsg{text: fmt.Sprintf("line %d", i)})
	}
	m.Update(paneRefreshMsg{})
	require.True(t, m.actions.vp.AtBottom())

	m.Update(logLineMsg{text: "line 50"})
	m.Update(paneRefreshMsg{})
	require.True(t, m.actions.vp.AtBottom())
	require.Contains(t, m.actions.vp.View(), "line 50")
}

func TestRefreshKeepsScrolledBackRowsInView(t *testing.T) {
	m := newSizedModel()
	m.Update(logLineMsg{text: "line 0a\nline 0b"})
	for i := 1; i < maxPaneLines; i++ {
		m.Update(logLineMsg{text: fmt.Sprintf("line %d", i)})
	}
	m.Update(paneRefreshMsg{})
	m.actions.vp.SetYOffset(101)
	require.Equal(t, "line 100", topRow(m.actions.vp))

	// Discards the two-row first entry and the one after it.
	m.Update(logLineMsg{text: "new 1"})
	m.Update(logLineMsg{text: "new 2"})
	m.Update(paneRefreshMsg{})
	require.Equal(t, 98, m.actions.vp.YOffset)
	require.Equal(t, "line 100", topRow(m.actions.vp))
}
