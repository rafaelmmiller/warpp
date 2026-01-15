package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Session struct {
	Name         string
	Description  string
	IsRunning    bool
	IsLayout     bool   // true if this is a layout entry (not a running session)
	ProjectRoot  string // parsed from session_root in layout file
	ProjectType  string // detected from layout or directory
	Icon         string // emoji for project type
	ClaudeStatus string // "executing", "idle", or "" (no Claude)
}

// GetAllSessions returns running sessions (first) + layouts (second) as separate entries
func GetAllSessions() ([]Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	layoutsDir := filepath.Join(home, ".tmuxifier", "layouts")
	entries, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, err
	}

	var runningSessions []Session
	var layouts []Session
	runningNames := getRunningTmuxSessions()
	layoutNames := make(map[string]bool)
	claudeStatus := GetClaudeSessionStatus() // Get claude status for all sessions at once

	// First, collect all tmuxifier layouts
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".session.sh") || e.Name() == "new-session.session.sh" {
			continue
		}

		name := strings.TrimSuffix(e.Name(), ".session.sh")
		if name == "" {
			continue
		}

		layoutNames[name] = true
		layoutPath := filepath.Join(layoutsDir, e.Name())
		layouts = append(layouts, Session{
			Name:        name,
			Description: getSessionDescription(layoutPath),
			IsRunning:   false,
			IsLayout:    true,
			ProjectRoot: getSessionRoot(layoutPath, home),
			ProjectType: detectProjectType(name),
			Icon:        "•",
		})
	}

	// Then, collect all running tmux sessions
	for _, runningName := range runningNames {
		session := Session{
			Name:         runningName,
			IsRunning:    true,
			IsLayout:     false,
			Icon:         "•",
			ClaudeStatus: claudeStatus[runningName],
		}
		if layoutNames[runningName] {
			// Has a layout - get description from it
			layoutPath := filepath.Join(layoutsDir, runningName+".session.sh")
			session.Description = getSessionDescription(layoutPath)
			session.ProjectRoot = getSessionRoot(layoutPath, home)
			session.ProjectType = detectProjectType(runningName)
		} else {
			// Orphan session
			session.Description = "(no layout)"
			session.ProjectType = "Orphan"
			session.Icon = "○"
		}
		runningSessions = append(runningSessions, session)
	}

	// Sort each group alphabetically
	sort.Slice(runningSessions, func(i, j int) bool {
		return runningSessions[i].Name < runningSessions[j].Name
	})
	sort.Slice(layouts, func(i, j int) bool {
		return layouts[i].Name < layouts[j].Name
	})

	// Combine: running sessions first, then layouts
	return append(runningSessions, layouts...), nil
}

// getRunningTmuxSessions returns list of currently running tmux sessions
func getRunningTmuxSessions() []string {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	var sessions []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions
}

// getSessionRoot extracts the default session_root path from a layout file
// Parses patterns like: session_root "${SESSION_ROOT:-~/path}" or session_root "~/path"
func getSessionRoot(layoutPath string, home string) string {
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "session_root ") {
			// Extract the argument
			arg := strings.TrimPrefix(line, "session_root ")
			arg = strings.Trim(arg, "\"'")

			// Handle ${SESSION_ROOT:-default} pattern
			if strings.HasPrefix(arg, "${SESSION_ROOT:-") && strings.HasSuffix(arg, "}") {
				arg = strings.TrimPrefix(arg, "${SESSION_ROOT:-")
				arg = strings.TrimSuffix(arg, "}")
			}

			// Skip if it's a variable reference we can't resolve
			if strings.HasPrefix(arg, "$") {
				return ""
			}

			// Expand ~ to home directory
			if strings.HasPrefix(arg, "~") {
				arg = filepath.Join(home, strings.TrimPrefix(arg, "~"))
			}

			return arg
		}
	}
	return ""
}

// getSessionDescription extracts description from session file comments
func getSessionDescription(layoutPath string) string {
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# Description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# Description:"))
		}
		if strings.HasPrefix(line, "#") && len(line) > 2 {
			desc := strings.TrimSpace(line[1:])
			if !strings.Contains(desc, "session") && !strings.Contains(desc, "tmux") {
				return desc
			}
		}
	}
	return "Tmux session layout"
}

// detectProjectType tries to guess project type from name/directory
func detectProjectType(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "blog"), strings.Contains(name, "web"):
		return "Web"
	case strings.Contains(name, "api"), strings.Contains(name, "server"):
		return "API"
	case strings.Contains(name, "extract"), strings.Contains(name, "crawl"):
		return "Data"
	default:
		return "Project"
	}
}

// getProjectIcon returns simple indicator for project type
func getProjectIcon(name string) string {
	return "•" // Simple bullet point
}

// LaunchSession launches a tmuxifier session
func LaunchSession(name string) error {
	cmd := exec.Command("tmuxifier", "load-session", name)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Attach or switch to session
	var attachCmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		attachCmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		attachCmd = exec.Command("tmux", "attach-session", "-t", name)
	}

	return attachCmd.Run()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// KillSession kills a running tmux session
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// CreateWorktree creates a new git worktree with a new branch
// Returns the path to the new worktree
func CreateWorktree(basePath, newDirName, branchName string) (string, error) {
	// Worktree goes in sibling directory
	parentDir := filepath.Dir(basePath)
	worktreePath := filepath.Join(parentDir, newDirName)

	// Create worktree with new branch
	cmd := exec.Command("git", "-C", basePath, "worktree", "add", "-b", branchName, worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create worktree: %s", string(output))
	}

	return worktreePath, nil
}

// GetRunningSessionNames returns a list of running tmux session names (exported for use in main.go)
func GetRunningSessionNames() []string {
	return getRunningTmuxSessions()
}

// CapturePane captures the content of a tmux session's active pane
func CapturePane(sessionName string, height int) string {
	// Capture the last N lines from the active pane with ANSI escape sequences (-e)
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p", "-e", "-S", fmt.Sprintf("-%d", height))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// PaneInfo holds information about a single pane
type PaneInfo struct {
	Index   int
	Width   int
	Height  int
	Content string
}

// GetSessionPanes returns all panes in the active window of a session
func GetSessionPanes(sessionName string) []PaneInfo {
	// List all panes in the session's current window
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_index}")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var panes []PaneInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		var idx int
		fmt.Sscanf(line, "%d", &idx)
		panes = append(panes, PaneInfo{Index: idx})
	}

	return panes
}

// CapturePaneByIndex captures content from a specific pane
func CapturePaneByIndex(sessionName string, paneIndex int, height int) string {
	target := fmt.Sprintf("%s:.%d", sessionName, paneIndex)
	// Use -e to preserve ANSI escape sequences (colors)
	cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-e", "-S", fmt.Sprintf("-%d", height))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// capturePanePlain captures content from a specific pane without ANSI codes (for text matching)
func capturePanePlain(sessionName string, paneIndex int, height int) string {
	target := fmt.Sprintf("%s:.%d", sessionName, paneIndex)
	cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-S", fmt.Sprintf("-%d", height))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// GetClaudeSessionStatus returns a map of session name -> claude status
// Uses process TTY matching for reliable detection
func GetClaudeSessionStatus() map[string]string {
	result := make(map[string]string)

	// Step 1: Get all claude CLI processes and their TTYs
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	// Find claude processes and extract TTYs
	claudeTTYs := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Match "claude" but not "Claude.app"
		if strings.Contains(line, "claude") && !strings.Contains(line, "Claude.app") && !strings.Contains(line, "grep") {
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				tty := fields[6] // TTY column
				claudeTTYs[tty] = true
			}
		}
	}

	// Step 2: Map TTYs to tmux sessions
	cmd = exec.Command("tmux", "list-panes", "-a", "-F", "#{session_name} #{pane_tty} #{pane_index}")
	output, err = cmd.Output()
	if err != nil {
		return result
	}

	sessionPanes := make(map[string][]int) // session -> list of pane indices with claude
	lines = strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			sessionName := fields[0]
			paneTTY := fields[1] // e.g., /dev/ttys044
			paneIdx, _ := strconv.Atoi(fields[2])

			// Convert TTY format: /dev/ttys044 -> s044
			shortTTY := strings.TrimPrefix(paneTTY, "/dev/tty")

			if claudeTTYs[shortTTY] {
				sessionPanes[sessionName] = append(sessionPanes[sessionName], paneIdx)
			}
		}
	}

	// Step 3: For each session with claude, check if executing or idle
	for sessionName, paneIndices := range sessionPanes {
		status := "idle" // default to idle

		for _, paneIdx := range paneIndices {
			content := capturePanePlain(sessionName, paneIdx, 15)
			// Check for active status line: "· <status>… (esc to interrupt"
			// This only appears when Claude is actively working, not in scrollback
			if strings.Contains(content, "(esc to interrupt") {
				status = "executing"
				break
			}
		}

		result[sessionName] = status
	}

	return result
}
