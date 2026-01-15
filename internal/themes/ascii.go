package themes

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var defaultASCIIArt = `████████╗███╗   ███╗██╗      █████╗ ██╗   ██╗███╗   ██╗ ██████╗██╗  ██╗
╚══██╔══╝████╗ ████║██║     ██╔══██╗██║   ██║████╗  ██║██╔════╝██║  ██║
   ██║   ██╔████╔██║██║     ███████║██║   ██║██╔██╗ ██║██║     ███████║
   ██║   ██║╚██╔╝██║██║     ██╔══██║██║   ██║██║╚██╗██║██║     ██╔══██║
   ██║   ██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝██║ ╚████║╚██████╗██║  ██║
   ╚═╝   ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝`

// GetASCIIArtFrames returns all animation frames for the given ASCII art name
// Frames are loaded from ~/.config/warpp/ascii-art/<name>/*.txt sorted alphabetically
func GetASCIIArtFrames(name string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{defaultASCIIArt}
	}

	artDir := filepath.Join(home, ".config", "warpp", "ascii-art", name)
	entries, err := os.ReadDir(artDir)
	if err != nil {
		if name != "fire" {
			return GetASCIIArtFrames("fire")
		}
		return []string{defaultASCIIArt}
	}

	// Collect and sort .txt files
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		if name != "fire" {
			return GetASCIIArtFrames("fire")
		}
		return []string{defaultASCIIArt}
	}

	// Load each frame
	var frames []string
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(artDir, file))
		if err == nil {
			frames = append(frames, strings.TrimRight(string(content), "\n\r"))
		}
	}

	if len(frames) == 0 {
		return []string{defaultASCIIArt}
	}

	return frames
}

// GetASCIIArt returns the first frame of ASCII art (for backwards compatibility)
func GetASCIIArt(name string) string {
	frames := GetASCIIArtFrames(name)
	if len(frames) > 0 {
		return frames[0]
	}
	return defaultASCIIArt
}

// ListAvailableASCIIArt returns all available ASCII art names (directories in ascii-art folder)
func ListAvailableASCIIArt() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{"fire"}
	}

	dirname := filepath.Join(home, ".config", "warpp", "ascii-art")
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return []string{"fire"}
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	if len(names) == 0 {
		return []string{"fire"}
	}

	return names
}
