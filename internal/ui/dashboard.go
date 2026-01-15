package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"warpp/internal/themes"
	"warpp/internal/tmux"
)

type Model struct {
	sessions []tmux.Session
	cursor   int
	theme    themes.Theme
	styles   themes.ThemeStyles
	width    int
	height   int
	ready    bool
}

func NewModel(themeName string) Model {
	theme := themes.GetTheme(themeName)
	return Model{
		theme:  theme,
		styles: theme.Styles(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			sessions, _ := tmux.GetTmuxifierLayouts()
			return sessionsLoadedMsg{sessions}
		},
	)
}

type sessionsLoadedMsg struct {
	sessions []tmux.Session
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		return m, nil

	case tea.KeyMsg:
		if !m.ready {
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				selected := m.sessions[m.cursor]
				// Launch session and quit
				go func() {
					tmux.LaunchSession(selected.Name)
				}()
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}

		case "home", "g":
			m.cursor = 0

		case "end", "G":
			m.cursor = len(m.sessions) - 1
		}
	}

	return m, nil
}

type errorMsg struct {
	err error
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Header with ASCII art
	header := m.styles.Title.Render(m.theme.ASCIIArt)

	// Sessions list
	var sessionItems []string
	for i, session := range m.sessions {
		var style lipgloss.Style
		cursor := "  "

		if i == m.cursor {
			style = m.styles.Selected
			cursor = "➤ "
		} else {
			style = m.styles.Normal
		}

		// Status indicator
		status := ""
		if session.IsRunning {
			status = m.styles.Success.Render("[RUNNING]")
		} else {
			status = m.styles.Muted.Render("[STOPPED]")
		}

		// Format: → icon name    description    [STATUS]
		line := fmt.Sprintf("%s%s %-12s %s %s",
			cursor,
			session.Icon,
			session.Name,
			m.styles.Muted.Render(session.Description),
			status,
		)

		sessionItems = append(sessionItems, style.Render(line))
	}

	// Main content box
	content := strings.Join(sessionItems, "\n")
	if len(sessionItems) == 0 {
		content = m.styles.Muted.Render("No tmuxifier sessions found")
	}

	// Footer with help
	footer := m.styles.Muted.Render("↑/↓ Navigate • Enter Launch • q Quit")

	// Layout everything
	mainBox := m.styles.Border.
		Width(m.width - 4).
		Height(m.height - 6).
		Render(content)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		mainBox,
		"",
		footer,
	)
}

func (m Model) GetTheme() themes.Theme {
	return m.theme
}
