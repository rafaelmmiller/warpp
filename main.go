package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"warpp/internal/config"
	"warpp/internal/themes"
	"warpp/internal/tmux"
)

// Claude's spinner frames for executing status
var claudeSpinnerFrames = []string{"·", "✻", "✽", "✶", "✳", "✢"}

// ansiRegex matches ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// truncateWithANSI truncates a string to maxWidth visible characters, preserving ANSI codes
func truncateWithANSI(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	var result strings.Builder
	visibleLen := 0
	i := 0

	for i < len(s) && visibleLen < maxWidth {
		// Check if we're at an ANSI escape sequence
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Find the end of the escape sequence
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				// Include the 'm'
				result.WriteString(s[i : j+1])
				i = j + 1
				continue
			}
		}

		// Regular character
		result.WriteByte(s[i])
		visibleLen++
		i++
	}

	// If we truncated, add ellipsis and reset code
	if visibleLen >= maxWidth && i < len(s) {
		// Check if there's more visible content
		remaining := s[i:]
		stripped := ansiRegex.ReplaceAllString(remaining, "")
		if len(stripped) > 0 {
			result.WriteString("…\x1b[0m")
		}
	}

	return result.String()
}

type simpleModel struct {
	sessions       []tmux.Session
	cursor         int
	theme          themes.Theme
	styles         themes.ThemeStyles
	asciiFrames    []string // ASCII art animation frames
	asciiFrame     int      // Current ASCII art frame index
	width          int
	height         int
	confirmingKill bool
	spinnerFrame   int // Current frame of Claude spinner animation
	// Worktree flow states
	worktreeInputStep   int    // 0=none, 1=session name, 2=branch name
	worktreeSessionName string // text input for session name
	worktreeBranchName  string // text input for branch name
	worktreeLayoutName  string // original layout name for worktree
	worktreeProjectRoot string // base path for worktree creation
	errorMessage        string // error message to display
}

// tickMsg is sent periodically to update the spinner animation
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m simpleModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			sessions, err := tmux.GetAllSessions()
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return sessions
		},
		tickCmd(),
	)
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
	case tickMsg:
		m.spinnerFrame = (m.spinnerFrame + 1) % len(claudeSpinnerFrames)
		if len(m.asciiFrames) > 1 {
			m.asciiFrame = (m.asciiFrame + 1) % len(m.asciiFrames)
		}
		return m, tickCmd()
	case string: // error message
		fmt.Println(msg)
		return m, tea.Quit
	case tea.KeyMsg:
		// Clear error message on any key
		if m.errorMessage != "" {
			m.errorMessage = ""
			return m, nil
		}

		// Handle worktree input flow
		if m.worktreeInputStep > 0 {
			switch msg.String() {
			case "esc":
				if m.worktreeInputStep == 2 {
					m.worktreeInputStep = 1
				} else {
					m.worktreeInputStep = 0
					m.worktreeSessionName = ""
					m.worktreeBranchName = ""
				}
				return m, nil
			case "enter":
				if m.worktreeInputStep == 1 {
					if m.worktreeSessionName == "" {
						return m, nil
					}
					m.worktreeInputStep = 2
				} else if m.worktreeInputStep == 2 {
					if m.worktreeBranchName == "" {
						return m, nil
					}
					// Create worktree and launch
					worktreePath, err := tmux.CreateWorktree(m.worktreeProjectRoot, m.worktreeSessionName, m.worktreeBranchName)
					if err != nil {
						m.errorMessage = err.Error()
						m.worktreeInputStep = 0
						m.worktreeSessionName = ""
						m.worktreeBranchName = ""
						return m, nil
					}
					// Launch with worktree
					launchWorktreeSession(m.worktreeLayoutName, m.worktreeSessionName, worktreePath)
					return m, tea.Quit
				}
				return m, nil
			case "backspace":
				if m.worktreeInputStep == 1 && len(m.worktreeSessionName) > 0 {
					m.worktreeSessionName = m.worktreeSessionName[:len(m.worktreeSessionName)-1]
				} else if m.worktreeInputStep == 2 && len(m.worktreeBranchName) > 0 {
					m.worktreeBranchName = m.worktreeBranchName[:len(m.worktreeBranchName)-1]
				}
				return m, nil
			default:
				// Add printable characters
				char := msg.String()
				if len(char) == 1 && char[0] >= 32 && char[0] < 127 {
					if m.worktreeInputStep == 1 {
						m.worktreeSessionName += char
					} else if m.worktreeInputStep == 2 {
						m.worktreeBranchName += char
					}
				}
				return m, nil
			}
		}

		// Handle confirmation dialog inputs
		if m.confirmingKill {
			switch msg.String() {
			case "y", "Y", "enter":
				if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
					selected := m.sessions[m.cursor]
					if selected.IsRunning {
						tmux.KillSession(selected.Name)
						m.confirmingKill = false
						// Refresh sessions list
						return m, func() tea.Msg {
							sessions, err := tmux.GetAllSessions()
							if err != nil {
								return fmt.Sprintf("Error: %v", err)
							}
							return sessions
						}
					}
				}
				m.confirmingKill = false
				return m, nil
			case "n", "N", "esc", "q":
				m.confirmingKill = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				selected := m.sessions[m.cursor]
				if selected.IsRunning {
					// Attach to running session
					launchSession(selected.Name)
					return m, tea.Quit
				} else if selected.IsLayout {
					// Check if session with same name is already running
					runningNames := tmux.GetRunningSessionNames()
					isAlreadyRunning := false
					for _, name := range runningNames {
						if name == selected.Name {
							isAlreadyRunning = true
							break
						}
					}

					if isAlreadyRunning {
						// Start worktree flow
						if selected.ProjectRoot == "" {
							m.errorMessage = "Layout must have session_root defined to create worktree sessions."
							return m, nil
						}
						if !tmux.IsGitRepo(selected.ProjectRoot) {
							m.errorMessage = "Project must be a git repository. Run 'git init' first."
							return m, nil
						}
						m.worktreeInputStep = 1
						m.worktreeLayoutName = selected.Name
						m.worktreeProjectRoot = selected.ProjectRoot
						m.worktreeSessionName = selected.Name + "-"
						return m, nil
					}
					// Normal layout launch
					launchSession(selected.Name)
					return m, tea.Quit
				}
			}
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		case "k":
			// Uppercase K triggers kill, lowercase k moves up
			if m.cursor > 0 {
				m.cursor--
			}
		case "K":
			// Kill session - only for running sessions
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				selected := m.sessions[m.cursor]
				if selected.IsRunning {
					m.confirmingKill = true
				}
			}
		case "n":
			launchNewSession()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m simpleModel) View() string {
	// ASCII art header - using current animation frame
	currentArt := ""
	if len(m.asciiFrames) > 0 {
		currentArt = m.asciiFrames[m.asciiFrame]
	}
	header := m.styles.Title.Render(currentArt)

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

	// Calculate max name width for alignment
	maxNameWidth := 0
	for _, session := range m.sessions {
		if len(session.Name) > maxNameWidth {
			maxNameWidth = len(session.Name)
		}
	}

	// Separate running sessions and layouts
	var runningSessions, layouts []tmux.Session
	for _, session := range m.sessions {
		if session.IsRunning {
			runningSessions = append(runningSessions, session)
		} else if session.IsLayout {
			layouts = append(layouts, session)
		}
	}

	// Track cursor position across both groups
	cursorIdx := 0

	// Orange style for Claude icons
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00"))

	// Render running sessions first
	if len(runningSessions) > 0 {
		items = append(items, m.styles.Muted.Render("SESSIONS:"))
		for _, session := range runningSessions {
			cursor := " "
			style := m.styles.Normal

			if cursorIdx == m.cursor {
				cursor = "→"
				style = m.styles.Selected.Padding(0, 1)
			}
			cursorIdx++

			// Determine icon based on Claude status
			var icon string
			switch session.ClaudeStatus {
			case "executing":
				icon = orangeStyle.Render(claudeSpinnerFrames[m.spinnerFrame])
			case "idle":
				icon = orangeStyle.Render("●")
			default:
				icon = session.Icon // Default icon (•)
			}

			line := fmt.Sprintf(" %s %s %s",
				cursor,
				icon,
				session.Name)

			items = append(items, style.Render(line))
		}
	}

	// Render layouts
	if len(layouts) > 0 {
		if len(runningSessions) > 0 {
			items = append(items, "") // Add spacing between sections
		}
		items = append(items, m.styles.Muted.Render("LAYOUTS:"))
		for _, session := range layouts {
			cursor := " "
			style := m.styles.Normal

			if cursorIdx == m.cursor {
				cursor = "→"
				style = m.styles.Selected.Padding(0, 1)
			}
			cursorIdx++

			line := fmt.Sprintf(" %s %s %s",
				cursor,
				session.Icon,
				session.Name)

			items = append(items, style.Render(line))
		}
	}

	// Create bordered content box with responsive width
	listWidth := 35
	previewWidth := 50
	if m.width > 0 && m.width < 100 {
		listWidth = 30
		previewWidth = m.width - listWidth - 10
	}

	// Calculate fixed height for both panels
	panelHeight := 15
	if m.height > 0 {
		panelHeight = m.height - 20
		if panelHeight < 8 {
			panelHeight = 8
		}
		if panelHeight > 30 {
			panelHeight = 30
		}
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Border)).
		Padding(1, 2).
		Width(listWidth).
		Height(panelHeight + 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, items...))

	// Build preview panel for running sessions
	var previewBox string
	if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
		selected := m.sessions[m.cursor]
		if selected.IsRunning {
			// Get all panes in the session
			panes := tmux.GetSessionPanes(selected.Name)
			numPanes := len(panes)
			if numPanes == 0 {
				numPanes = 1
			}

			totalPreviewHeight := panelHeight

			// Calculate height per pane (account for separators between panes)
			separatorLines := numPanes - 1
			availableForPanes := totalPreviewHeight - separatorLines
			heightPerPane := availableForPanes / numPanes
			if heightPerPane < 3 {
				heightPerPane = 3
			}

			maxLineWidth := previewWidth - 3
			var paneContents []string

			for i, pane := range panes {
				// Capture content for this pane
				content := tmux.CapturePaneByIndex(selected.Name, pane.Index, heightPerPane)

				// Truncate lines (ANSI-aware)
				lines := strings.Split(content, "\n")
				var truncatedLines []string
				for j, line := range lines {
					if j >= heightPerPane {
						break
					}
					truncatedLines = append(truncatedLines, truncateWithANSI(line, maxLineWidth))
				}

				paneContents = append(paneContents, strings.Join(truncatedLines, "\n"))

				// Add separator between panes (not after last one)
				if i < len(panes)-1 {
					borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Border))
					separator := borderStyle.Render(strings.Repeat("─", maxLineWidth))
					paneContents = append(paneContents, separator)
				}
			}

			// If no panes found, fall back to single pane capture
			if len(panes) == 0 {
				content := tmux.CapturePane(selected.Name, totalPreviewHeight)
				lines := strings.Split(content, "\n")
				var truncatedLines []string
				for i, line := range lines {
					if i >= totalPreviewHeight {
						break
					}
					truncatedLines = append(truncatedLines, truncateWithANSI(line, maxLineWidth))
				}
				paneContents = append(paneContents, strings.Join(truncatedLines, "\n"))
			}

			previewBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(m.theme.Border)).
				Padding(0, 1).
				Width(previewWidth).
				Height(totalPreviewHeight + 2).
				Render(lipgloss.JoinVertical(lipgloss.Left, paneContents...))
		} else {
			// Show layout info for non-running sessions
			infoText := lipgloss.JoinVertical(lipgloss.Left,
				m.styles.Title.Render(selected.Name),
				"",
				m.styles.Muted.Render("Layout not running"),
				"",
				m.styles.Normal.Render("Press Enter to launch"),
			)
			previewBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(m.theme.Border)).
				Padding(1, 2).
				Width(previewWidth).
				Height(panelHeight + 2).
				Render(infoText)
		}
	}

	// Combine list and preview side by side
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, contentBox, "  ", previewBox)

	// Footer with keybindings
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Muted)).
		Render("↑/↓ Navigate  •  Enter Launch  •  K Kill  •  n New  •  q Quit")

	// Build main content
	var content string

	if m.errorMessage != "" {
		// Show error message
		errorBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.Error)).
			Padding(1, 3).
			Align(lipgloss.Center).
			Render(lipgloss.JoinVertical(lipgloss.Center,
				m.errorMessage,
				"",
				lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Muted)).Render("Press any key to continue"),
			))

		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			"",
			mainContent,
			"",
			errorBox,
		)
	} else if m.worktreeInputStep > 0 {
		// Show worktree input dialog
		var promptText, inputValue, hint string
		if m.worktreeInputStep == 1 {
			promptText = "Enter session name:"
			inputValue = m.worktreeSessionName
			hint = "Enter to continue  •  Esc to cancel"
		} else {
			promptText = "Enter branch name:"
			inputValue = m.worktreeBranchName
			hint = "Enter to create  •  Esc to go back"
		}

		inputLine := lipgloss.NewStyle().
			Background(lipgloss.Color(m.theme.Background)).
			Foreground(lipgloss.Color(m.theme.Text)).
			Padding(0, 1).
			Render(inputValue + "█")

		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.Primary)).
			Padding(1, 3).
			Align(lipgloss.Center).
			Render(lipgloss.JoinVertical(lipgloss.Center,
				promptText,
				"",
				inputLine,
				"",
				lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Muted)).Render(hint),
			))

		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			"",
			mainContent,
			"",
			inputBox,
		)
	} else if m.confirmingKill && len(m.sessions) > 0 && m.cursor < len(m.sessions) {
		selected := m.sessions[m.cursor]

		// Create confirmation dialog
		confirmText := fmt.Sprintf("Kill session '%s'?", selected.Name)
		confirmHint := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.Muted)).
			Render("y/Enter to confirm  •  n/Esc to cancel")

		confirmBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.Warning)).
			Padding(1, 3).
			Align(lipgloss.Center).
			Render(lipgloss.JoinVertical(lipgloss.Center, confirmText, "", confirmHint))

		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			"",
			mainContent,
			"",
			confirmBox,
		)
	} else {
		// Combine everything with proper spacing
		content = lipgloss.JoinVertical(lipgloss.Center,
			header,
			"",
			mainContent,
			"",
			footer,
		)
	}

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
			fmt.Println("warpp v1.0.0")
			return
		case "--help", "-h":
			printHelp()
			return
		case "--new", "-n":
			launchNewSession()
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

	// Get ASCII art frames from config
	asciiFrames := themes.GetASCIIArtFrames(cfg.ASCIIArt)

	m := simpleModel{
		theme:       theme,
		styles:      theme.Styles(),
		asciiFrames: asciiFrames,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Clear screen on exit for clean terminal
	fmt.Print("\033[H\033[2J")
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

	frames := themes.GetASCIIArtFrames(cfg.ASCIIArt)
	fmt.Printf("Testing ASCII art: %s (%d frames)\n", cfg.ASCIIArt, len(frames))
	fmt.Println("Frame 1:")
	fmt.Println(frames[0])
}

func printHelp() {
	fmt.Println("warpp - Warp into your tmux sessions")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  warpp                  Launch the TUI interface")
	fmt.Println("  warpp --new, -n        Create new session in current directory")
	fmt.Println("  warpp config           Show current configuration")
	fmt.Println("  warpp init-config      Create default config file")
	fmt.Println("  warpp test-ascii       Test current ASCII art setting")
	fmt.Println("  warpp --help           Show this help")
	fmt.Println("  warpp --version        Show version")
	fmt.Println()
	fmt.Println("Config file location: ~/.config/warpp/config.json")
	fmt.Printf("Available themes: %s\n", strings.Join(themes.ListAvailableThemes(), ", "))
	fmt.Println("Available ASCII art: fire, blocks, minimal")
}

func launchSession(sessionName string) {
	// Check if this is a tmuxifier layout or orphan session
	home, _ := os.UserHomeDir()
	layoutPath := filepath.Join(home, ".tmuxifier", "layouts", sessionName+".session.sh")

	if _, err := os.Stat(layoutPath); err == nil {
		// Has a layout file - use tmuxifier
		cmd := exec.Command("tmuxifier", "load-session", sessionName)
		cmd.Run()
	}
	// If no layout, session already exists - just attach

	// Attach/switch to session
	var args []string
	if os.Getenv("TMUX") != "" {
		args = []string{"tmux", "switch-client", "-t", sessionName}
	} else {
		args = []string{"tmux", "attach-session", "-t", sessionName}
	}

	tmuxPath, _ := exec.LookPath("tmux")
	syscall.Exec(tmuxPath, args, os.Environ())
}

func launchWorktreeSession(layoutName, sessionName, worktreePath string) {
	// Set environment variables to override session_root and session name
	os.Setenv("SESSION_ROOT", worktreePath)
	os.Setenv("SESSION_NAME", sessionName)

	// Load the layout using bash -c to ensure env is passed
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("SESSION_ROOT=%q SESSION_NAME=%q tmuxifier load-session %s", worktreePath, sessionName, layoutName))
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Attach to the session
	var args []string
	if os.Getenv("TMUX") != "" {
		args = []string{"tmux", "switch-client", "-t", sessionName}
	} else {
		args = []string{"tmux", "attach-session", "-t", sessionName}
	}

	tmuxPath, _ := exec.LookPath("tmux")
	syscall.Exec(tmuxPath, args, os.Environ())
}

func launchNewSession() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	sessionName := filepath.Base(cwd)

	// Sanitize session name: remove leading dots and colons (tmux doesn't handle them well)
	sessionName = strings.TrimPrefix(sessionName, ".")
	sessionName = strings.ReplaceAll(sessionName, ":", "-")

	// Set environment variables in current process so they're inherited
	os.Setenv("NEW_SESSION_NAME", sessionName)
	os.Setenv("NEW_SESSION_ROOT", cwd)

	// Load the new-session layout using bash -c to ensure env is passed
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("cd %q && tmuxifier load-session new-session", cwd))
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Attach to the session
	var args []string
	if os.Getenv("TMUX") != "" {
		args = []string{"tmux", "switch-client", "-t", sessionName}
	} else {
		args = []string{"tmux", "attach-session", "-t", sessionName}
	}

	tmuxPath, _ := exec.LookPath("tmux")
	syscall.Exec(tmuxPath, args, os.Environ())
}
