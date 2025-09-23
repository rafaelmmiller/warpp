package themes

import (
	"os"
	"path/filepath"
	"strings"
)

// GetASCIIArt returns the ASCII art for the given name
func GetASCIIArt(name string) string {
	// Try to read from the local directory first (for development/runtime changes)
	filename := filepath.Join("internal", "themes", "ascii-art", name+".txt")

	content, err := os.ReadFile(filename)
	if err != nil {
		// Fallback to hardcoded default if file not found
		if name != "tmlaunch" {
			return GetASCIIArt("tmlaunch")
		}
		// If even tmlaunch fails, return hardcoded fallback
		return `████████╗███╗   ███╗██╗      █████╗ ██╗   ██╗███╗   ██╗ ██████╗██╗  ██╗
╚══██╔══╝████╗ ████║██║     ██╔══██╗██║   ██║████╗  ██║██╔════╝██║  ██║
   ██║   ██╔████╔██║██║     ███████║██║   ██║██╔██╗ ██║██║     ███████║
   ██║   ██║╚██╔╝██║██║     ██╔══██║██║   ██║██║╚██╗██║██║     ██╔══██║
   ██║   ██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝██║ ╚████║╚██████╗██║  ██║
   ╚═╝   ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝`
	}

	return strings.TrimRight(string(content), "\n\r")
}

// ListAvailableASCIIArt returns all available ASCII art names
func ListAvailableASCIIArt() []string {
	dirname := filepath.Join("internal", "themes", "ascii-art")
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return []string{"tmlaunch"}
	}

	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".txt") {
			name := strings.TrimSuffix(entry.Name(), ".txt")
			names = append(names, name)
		}
	}

	return names
}
