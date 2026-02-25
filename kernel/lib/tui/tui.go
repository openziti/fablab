/*
	(c) Copyright NetFoundry Inc. Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

// Messages

type logLineMsg struct {
	pane string
	text string
}

type iterationMsg struct {
	num int
}

type execDoneMsg struct {
	err error
}

type tickMsg time.Time

// Styles

var (
	statusBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62"))

	paneTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("240"))

	focusedPaneTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63"))

	doneSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10"))

	doneErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))
)

const (
	focusActions    = 0
	focusValidation = 1
)

// Model

type tuiModel struct {
	iteration    int
	totalStart   time.Time
	iterStart    time.Time
	actionsVP    viewport.Model
	validationVP viewport.Model
	actionsLines []string
	validLines   []string
	width        int
	height       int
	focus        int
	done         bool
	doneAt       time.Time
	err          error
}

func newTuiModel() *tuiModel {
	return &tuiModel{
		totalStart: time.Now(),
		iterStart:  time.Now(),
		focus:      focusActions,
	}
}

func (m *tuiModel) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.focus = (m.focus + 1) % 2
			return m, nil
		}

		// Forward key events to the focused viewport.
		if m.focus == focusActions {
			var cmd tea.Cmd
			m.actionsVP, cmd = m.actionsVP.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			var cmd tea.Cmd
			m.validationVP, cmd = m.validationVP.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.MouseMsg:
		// Forward mouse events to both viewports; each viewport checks
		// whether the event falls within its bounds using YPosition.
		var cmd tea.Cmd
		m.actionsVP, cmd = m.actionsVP.Update(msg)
		cmds = append(cmds, cmd)
		m.validationVP, cmd = m.validationVP.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()

	case logLineMsg:
		if msg.pane == PaneValidation {
			m.validLines = append(m.validLines, msg.text)
			wasAtBottom := m.validationVP.AtBottom()
			m.validationVP.SetContent(strings.Join(m.validLines, "\n"))
			if wasAtBottom {
				m.validationVP.GotoBottom()
			}
		} else {
			m.actionsLines = append(m.actionsLines, msg.text)
			wasAtBottom := m.actionsVP.AtBottom()
			m.actionsVP.SetContent(strings.Join(m.actionsLines, "\n"))
			if wasAtBottom {
				m.actionsVP.GotoBottom()
			}
		}

	case iterationMsg:
		m.iteration = msg.num
		m.iterStart = time.Now()

	case execDoneMsg:
		m.done = true
		m.doneAt = time.Now()
		m.err = msg.err

	case tickMsg:
		cmds = append(cmds, tickCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m *tuiModel) recalcLayout() {
	// Layout: 1 line status bar + 1 line actions title + actions viewport + 1 line validation title + validation viewport
	// Total overhead = 3 lines (status bar + 2 pane titles)
	overhead := 3
	remaining := max(m.height-overhead, 2)
	actionsHeight := remaining / 2
	validationHeight := remaining - actionsHeight

	m.actionsVP.Width = m.width
	m.actionsVP.Height = actionsHeight
	m.actionsVP.YPosition = 2 // after status bar + actions title
	m.validationVP.Width = m.width
	m.validationVP.Height = validationHeight
	m.validationVP.YPosition = 2 + actionsHeight + 1 // after actions viewport + validation title

	// Re-set content to recalculate line wrapping.
	m.actionsVP.SetContent(strings.Join(m.actionsLines, "\n"))
	m.actionsVP.GotoBottom()
	m.validationVP.SetContent(strings.Join(m.validLines, "\n"))
	m.validationVP.GotoBottom()
}

func (m *tuiModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var b strings.Builder

	// Status bar — freeze times once done.
	now := time.Now()
	if m.done {
		now = m.doneAt
	}
	iterStr := fmt.Sprintf(" Iteration: %03d", m.iteration)
	totalStr := fmt.Sprintf("Total: %s", formatDuration(now.Sub(m.totalStart)))
	iterTimeStr := fmt.Sprintf("Iter: %s", formatDuration(now.Sub(m.iterStart)))

	var statusRight string
	if m.done {
		if m.err != nil {
			statusRight = doneErrorStyle.Render(fmt.Sprintf(" FAILED: %v ", m.err))
		} else {
			statusRight = doneSuccessStyle.Render(" COMPLETED — press q to exit ")
		}
	}

	statusContent := fmt.Sprintf("%s  |  %s  |  %s  %s", iterStr, totalStr, iterTimeStr, statusRight)
	statusBar := statusBarStyle.Width(m.width).Render(statusContent)
	b.WriteString(statusBar)
	b.WriteString("\n")

	// Actions pane title
	actionsTitle := " Actions "
	if m.focus == focusActions {
		b.WriteString(focusedPaneTitleStyle.Width(m.width).Render(actionsTitle))
	} else {
		b.WriteString(paneTitleStyle.Width(m.width).Render(actionsTitle))
	}
	b.WriteString("\n")

	// Actions viewport
	b.WriteString(m.actionsVP.View())
	b.WriteString("\n")

	// Validation pane title
	validTitle := " Validation "
	if m.focus == focusValidation {
		b.WriteString(focusedPaneTitleStyle.Width(m.width).Render(validTitle))
	} else {
		b.WriteString(paneTitleStyle.Width(m.width).Render(validTitle))
	}
	b.WriteString("\n")

	// Validation viewport
	b.WriteString(m.validationVP.View())

	return b.String()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	if m >= 60 {
		h := m / 60
		m = m % 60
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	return fmt.Sprintf("%dm%02ds", m, s)
}

// RunTUI starts the TUI and returns a Program handle. The caller should run the exec loop
// in a goroutine, sending iterationMsg and execDoneMsg via program.Send().
// The logrus hook is installed automatically.
// When the TUI exits, active is set to false and logrus output is restored.
func RunTUI() (*tea.Program, error) {
	active.Store(true)

	model := newTuiModel()
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	hook := newLogHook(p)
	logrus.AddHook(hook)

	// Redirect logrus output to discard — the hook handles display.
	origOut := logrus.StandardLogger().Out
	logrus.SetOutput(io.Discard)

	go func() {
		if _, err := p.Run(); err != nil {
			// If the TUI fails, restore output so logs are visible.
			active.Store(false)
			logrus.SetOutput(origOut)
			logrus.WithError(err).Error("TUI error")
		}
		active.Store(false)
		logrus.SetOutput(origOut)
	}()

	// Give the TUI a moment to start up and get its initial window size.
	time.Sleep(100 * time.Millisecond)

	return p, nil
}

// SendIteration sends an iteration update to the TUI program.
func SendIteration(p *tea.Program, num int) {
	p.Send(iterationMsg{num: num})
}

// SendDone signals the TUI that the exec loop has finished.
func SendDone(p *tea.Program, err error) {
	p.Send(execDoneMsg{err: err})
}
