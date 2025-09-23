package themes

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Primary    string // Main accent color
	Secondary  string // Secondary accent
	Background string // Background
	Text       string // Main text
	Muted      string // Dim text
	Success    string // Running sessions
	Warning    string // Stopped sessions
	Error      string // Error states
	Border     string // Box borders
	Cursor     string // Selection cursor
	ASCIIArt   string // Theme-specific ASCII art
}

// Styles creates lipgloss styles from theme colors
func (t Theme) Styles() ThemeStyles {
	return ThemeStyles{
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Primary)).
			Bold(true),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Text)).
			Bold(true),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Border)),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Background)).
			Background(lipgloss.Color(t.Primary)).
			Bold(true),

		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Text)),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Muted)),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Success)),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Warning)),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Error)),
	}
}

type ThemeStyles struct {
	Title    lipgloss.Style
	Header   lipgloss.Style
	Border   lipgloss.Style
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Muted    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Error    lipgloss.Style
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	switch name {
	case "carbonfox":
		return CarbonFox()
	case "kanagawa":
		return Kanagawa()
	default:
		return Default()
	}
}
