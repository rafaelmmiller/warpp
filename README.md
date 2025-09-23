# TMlaunch

A beautiful terminal user interface for managing and launching tmuxifier sessions.

![TMlaunch Preview](assets/tmlaunch-preview.png)

## Features

- 🎨 **Beautiful TUI** - Clean, modern interface built with Bubble Tea
- 📋 **Session Overview** - View all your tmuxifier layouts at a glance
- 🟢 **Live Status** - See which sessions are currently running
- 🚀 **Quick Launch** - Launch sessions with a single keypress
- 📝 **Descriptions** - Shows session descriptions from layout comments
- 🎯 **Smart Sorting** - Running sessions appear first, then alphabetical

## Prerequisites

- [tmux](https://github.com/tmux/tmux) - Terminal multiplexer
- [tmuxifier](https://github.com/jimeh/tmuxifier) - Tmux session management tool
- Go 1.21+ (for building from source)

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd tmlaunch

# Build the application
go build -o tmlaunch .

# Move to your PATH (optional)
mv tmlaunch /usr/local/bin/
```

## How it Works

TMlaunch integrates seamlessly with your existing tmuxifier setup:

1. **Scans** your `~/.tmuxifier/layouts/` directory for `.session.sh` files
2. **Checks** which sessions are currently running via `tmux list-sessions`
3. **Displays** all sessions with their status in a beautiful TUI
4. **Launches** selected sessions using `tmuxifier load-session`
5. **Attaches** to the session (or switches if already in tmux)

## Usage

### Basic Usage

```bash
tmlaunch                # Launch the TUI interface
tmlaunch config         # Show current configuration
tmlaunch init-config    # Create default config file
tmlaunch --help         # Show help
tmlaunch --version      # Show version
```

### TUI Controls

- `↑/↓` or `k/j` - Navigate between sessions
- `Enter` - Launch selected session
- `q` or `Ctrl+C` - Quit

## Session Descriptions

Add descriptions to your tmuxifier layouts by including comments:

```bash
#!/usr/bin/env bash
# Description: My awesome web development environment

# Your tmuxifier layout code here...
```

TMlaunch will automatically extract and display these descriptions in the interface.

## Configuration

TMlaunch can be configured via a JSON config file at `~/.config/tmlaunch/config.json`.

### Example config.json

```json
{
  "theme": "carbonfox",
  "ascii_art": "tmlaunch"
}
```

### Available Options

- `theme`: Theme to use for the interface
  - `"default"` (default) - Clean, minimal theme
  - `"carbonfox"` - Dark theme with blue accents
  - `"kanagawa"` - Warm, nature-inspired theme

- `ascii_art`: ASCII art style for the header
  - `"tmlaunch"` (default) - Full TMLAUNCH banner
  - `"simple"` - Compact box-drawing style
  - `"minimal"` - Just text
  - `"blocks"` - Block character style
  - `"ThePersistenceOfMemory"` - Salvador Dali-inspired art

### Managing Configuration

```bash
# Create a config file with defaults
tmlaunch init-config

# View current configuration
tmlaunch config

# Edit the config file manually
$EDITOR ~/.config/tmlaunch/config.json
```

The config file is created automatically with default values if it doesn't exist.

## Requirements

Make sure you have tmuxifier properly configured with session layouts in `~/.tmuxifier/layouts/`. Each layout should be a `.session.sh` file containing your tmux session configuration.

## Example Tmuxifier Layout

```bash
#!/usr/bin/env bash
# Description: Web development environment

session_root "~/projects/myapp"

if initialize_session "myapp"; then
  new_window "editor"
  run_cmd "vim ."
  
  new_window "server"
  run_cmd "npm run dev"
  
  new_window "terminal"
  
  select_window 1
fi

finalize_and_go_to_session
```

## Contributing

TMlaunch is an open-source project and contributions are welcome! You can help by:

- 🎨 **Adding new themes** - Create new color schemes in `internal/themes/`
- 🖼️ **Adding ASCII art** - Add new ASCII art files to `internal/themes/ascii-art/`
- 🐛 **Fixing bugs** - Report issues and submit bug fixes
- 📖 **Improving documentation** - Help make the docs clearer and more comprehensive
- ✨ **Adding features** - Propose and implement new functionality

To contribute:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see the LICENSE file for details.
