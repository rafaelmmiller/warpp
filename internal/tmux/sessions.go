package tmux

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Session struct {
	Name        string
	Description string
	IsRunning   bool
	ProjectType string // detected from layout or directory
	Icon        string // emoji for project type
}

// GetTmuxifierLayouts returns all available tmuxifier layouts
func GetTmuxifierLayouts() ([]Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	layoutsDir := filepath.Join(home, ".tmuxifier", "layouts")
	entries, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	runningSessions := getRunningTmuxSessions()

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".session.sh") {
			continue
		}

		name := strings.TrimSuffix(e.Name(), ".session.sh")
		if name == "" {
			continue
		}

		session := Session{
			Name:        name,
			Description: getSessionDescription(filepath.Join(layoutsDir, e.Name())),
			IsRunning:   contains(runningSessions, name),
			ProjectType: detectProjectType(name),
			Icon:        getProjectIcon(name),
		}

		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		// Running sessions first, then alphabetical
		if sessions[i].IsRunning != sessions[j].IsRunning {
			return sessions[i].IsRunning
		}
		return sessions[i].Name < sessions[j].Name
	})

	return sessions, nil
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

// getSessionDescription extracts description from session file comments
func getSessionDescription(filepath string) string {
	content, err := os.ReadFile(filepath)
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
