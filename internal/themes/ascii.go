package themes

import (
	"embed"
	"strings"
)

//go:embed ascii-art/*.txt
var asciiArt embed.FS

// GetASCIIArt returns the ASCII art for the given name
func GetASCIIArt(name string) string {
	filename := "ascii-art/" + name + ".txt"

	content, err := asciiArt.ReadFile(filename)
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
	entries, err := asciiArt.ReadDir("ascii-art")
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
