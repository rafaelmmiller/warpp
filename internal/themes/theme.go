package themes

import (
	"encoding/json"
	"github.com/charmbracelet/lipgloss"
	"os"
	"path/filepath"
	"strings"
)

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

// ThemeJSON represents a theme loaded from JSON
type ThemeJSON struct {
	Name       string `json:"name"`
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
	Text       string `json:"text"`
	Muted      string `json:"muted"`
	Success    string `json:"success"`
	Warning    string `json:"warning"`
	Error      string `json:"error"`
	Border     string `json:"border"`
	Cursor     string `json:"cursor"`
}

// toTheme converts ThemeJSON to Theme
func (tj ThemeJSON) toTheme() Theme {
	return Theme{
		Name:       tj.Name,
		Primary:    tj.Primary,
		Secondary:  tj.Secondary,
		Background: tj.Background,
		Text:       tj.Text,
		Muted:      tj.Muted,
		Success:    tj.Success,
		Warning:    tj.Warning,
		Error:      tj.Error,
		Border:     tj.Border,
		Cursor:     tj.Cursor,
		ASCIIArt:   "", // ASCII art is handled separately
	}
}

// loadThemeFromFile attempts to load a theme from a JSON file
func loadThemeFromFile(name string) (Theme, bool) {
	filename := filepath.Join("themes", name+".json")

	content, err := os.ReadFile(filename)
	if err != nil {
		return Theme{}, false
	}

	var themeJSON ThemeJSON
	if err := json.Unmarshal(content, &themeJSON); err != nil {
		return Theme{}, false
	}

	return themeJSON.toTheme(), true
}

// ListAvailableThemes returns all available theme names
func ListAvailableThemes() []string {
	themes := []string{"default", "carbonfox", "kanagawa"} // hardcoded themes

	// Add themes from JSON files
	dirname := "themes"
	if entries, err := os.ReadDir(dirname); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				name := strings.TrimSuffix(entry.Name(), ".json")
				// Avoid duplicates
				found := false
				for _, existing := range themes {
					if existing == name {
						found = true
						break
					}
				}
				if !found {
					themes = append(themes, name)
				}
			}
		}
	}

	return themes
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	// Try to load from JSON file first
	if theme, ok := loadThemeFromFile(name); ok {
		return theme
	}

	// Fallback to hardcoded themes
	switch name {
	case "carbonfox":
		return CarbonFox()
	case "kanagawa":
		return Kanagawa()
	default:
		return Default()
	}
}
