package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tmlaunch/internal/config"
	"tmlaunch/internal/themes"
	"tmlaunch/internal/tmux"
)

type simpleModel struct {
	sessions []tmux.Session
	cursor   int
	theme    themes.Theme
	styles   themes.ThemeStyles
	asciiArt string
	width    int
	height   int
}

func (m simpleModel) Init() tea.Cmd {
	return func() tea.Msg {
		sessions, err := tmux.GetTmuxifierLayouts()
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return sessions
	}
}

func (m simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case []tmux.Session:
		m.sessions = msg
		return m, nil
	case string: // error message
		fmt.Println(msg)
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				selected := m.sessions[m.cursor]
				// Launch session synchronously like the old working version
				launchSession(selected.Name)
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
		}
	}
	return m, nil
}

func (m simpleModel) View() string {
	// ASCII art header - using configurable art
	header := m.styles.Title.Render(m.asciiArt)

	if len(m.sessions) == 0 {
		loadingBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.Border)).
			Padding(2, 4).
			Align(lipgloss.Center).
			Render("Loading sessions...")

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, header, "", loadingBox))
	}

	// Create sessions list with consistent formatting
	var items []string

	// Calculate max widths for consistent alignment
	maxNameWidth := 0
	maxDescWidth := 0
	for _, session := range m.sessions {
		if len(session.Name) > maxNameWidth {
			maxNameWidth = len(session.Name)
		}
		if len(session.Description) > maxDescWidth {
			maxDescWidth = len(session.Description)
		}
	}

	// Cap description width to prevent overly long lines
	if maxDescWidth > 25 {
		maxDescWidth = 25
	}

	for i, session := range m.sessions {
		cursor := " "
		style := m.styles.Normal

		if i == m.cursor {
			cursor = "→"
			style = m.styles.Selected.Padding(0, 1)
		}

		status := ""
		statusStyle := m.styles.Muted
		if session.IsRunning {
			status = "RUNNING"
			statusStyle = m.styles.Success
		} else {
			status = "STOPPED"
		}

		// Truncate description if too long
		desc := session.Description
		if len(desc) > maxDescWidth {
			desc = desc[:maxDescWidth-3] + "..."
		}

		// Fixed-width formatting for perfect alignment
		line := fmt.Sprintf(" %s %s %-*s  %-*s  %s ",
			cursor,
			session.Icon,
			maxNameWidth, session.Name,
			maxDescWidth, m.styles.Muted.Render(desc),
			statusStyle.Render(fmt.Sprintf("%-7s", status)))

		items = append(items, style.Render(line))
	}

	// Create bordered content box with responsive width
	boxWidth := 70
	if m.width > 0 && m.width < 80 {
		boxWidth = m.width - 10 // Leave some margin
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Border)).
		Padding(1, 2).
		Width(boxWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, items...))

	// Footer with keybindings
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Muted)).
		Render("↑/↓ Navigate  •  Enter Launch  •  q Quit")

	// Combine everything with proper spacing
	content := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"",
		contentBox,
		"",
		footer,
	)

	// Center everything on screen using actual terminal dimensions
	if m.width == 0 || m.height == 0 {
		// Fallback if no size info yet
		return content
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func main() {
	// Handle config commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			handleConfigCommand()
			return
		case "init-config":
			if err := config.InitConfig(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "test-ascii":
			handleTestASCII()
			return
		case "--version", "-v":
			fmt.Println("tmlaunch v1.0.0")
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Get theme from config
	theme := themes.GetTheme(cfg.Theme)

	// Get ASCII art from config
	asciiArt := themes.GetASCIIArt(cfg.ASCIIArt)

	m := simpleModel{
		theme:    theme,
		styles:   theme.Styles(),
		asciiArt: asciiArt,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleConfigCommand() {
	if err := config.ShowConfig(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleTestASCII() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Testing ASCII art: %s\n", cfg.ASCIIArt)
	art := themes.GetASCIIArt(cfg.ASCIIArt)
	fmt.Println("ASCII Art Output:")
	fmt.Println(art)
}

func printHelp() {
	fmt.Println("tmlaunch - Terminal User Interface for tmuxifier sessions")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tmlaunch               Launch the TUI interface")
	fmt.Println("  tmlaunch config        Show current configuration")
	fmt.Println("  tmlaunch init-config   Create default config file")
	fmt.Println("  tmlaunch test-ascii    Test current ASCII art setting")
	fmt.Println("  tmlaunch --help        Show this help")
	fmt.Println("  tmlaunch --version     Show version")
	fmt.Println()
	fmt.Println("Config file location: ~/.config/tmlaunch/config.json")
	fmt.Printf("Available themes: %s\n", strings.Join(themes.ListAvailableThemes(), ", "))
	fmt.Println("Available ASCII art: tmlaunch, simple, minimal, blocks")
}

func launchSession(layoutName string) {
	// Load the session
	cmd := exec.Command("tmuxifier", "load-session", layoutName)
	cmd.Run()

	// Now exec tmux attach to completely replace this process
	var args []string
	if os.Getenv("TMUX") != "" {
		args = []string{"tmux", "switch-client", "-t", layoutName}
	} else {
		args = []string{"tmux", "attach-session", "-t", layoutName}
	}

	// Find tmux path dynamically
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Printf("tmux not found: %v\n", err)
		return
	}

	syscall.Exec(tmuxPath, args, os.Environ())
}
