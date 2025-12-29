# Claude Watch

A lightweight, cross-platform desktop app that monitors Claude Code context usage across all your active sessions.

![Claude Watch Screenshot](screenshot.png)

## Features

- 🖥️ **Floating window** - Always on top, stays visible while you work
- 📊 **Real-time monitoring** - Updates every 5 seconds
- 🔔 **Smart alerts** - Notification + sound at 75% and 90% context usage
- 🌍 **Cross-platform** - Works on macOS, Windows, and Linux
- 🚀 **Lightweight** - ~10MB binary, minimal resource usage
- 📁 **Multi-project** - Monitors all active Claude Code sessions automatically

## Installation

### Prerequisites

1. **Go 1.21+** - [Download Go](https://go.dev/dl/)
2. **Wails CLI** - Install with:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
3. **Platform dependencies**:
   - **macOS**: Xcode command line tools (`xcode-select --install`)
   - **Windows**: WebView2 (usually pre-installed on Windows 10/11)
   - **Linux**: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`

### Build

```bash
# Clone or download this project
cd claude-watch

# Development mode (hot reload)
make dev

# Build for your current platform
make build

# Build for all platforms
make build-all
```

### Install (macOS)

```bash
make install-mac
# App will be available in /Applications
```

## Usage

1. **Launch the app** - Double-click `claude-watch.app` (macOS) or `claude-watch.exe` (Windows)
2. **Position the window** - Drag to your preferred corner (top-right recommended)
3. **Work normally** - The app monitors all Claude Code sessions in the background
4. **Get alerted** - At 75% and 90% context, you'll receive:
   - System notification
   - Audio alert
   - Voice warning (macOS at 90%)

### What it monitors

The app scans `~/.claude/projects/` for active sessions (modified in the last 60 minutes) and calculates context usage by:
- Summing all output tokens (they accumulate in context)
- Adding current input + cache tokens
- Displaying as percentage of 200k token limit

## Configuration

Currently, configuration is done by editing the source:

| Setting | Location | Default |
|---------|----------|---------|
| Update interval | `frontend/main.js` | 5000ms |
| Active session cutoff | `monitor.go` | 60 minutes |
| Max projects shown | `monitor.go` | 10 |
| Context limit | `monitor.go` | 200,000 tokens |

## Project Structure

```
claude-watch/
├── main.go           # Wails entry point
├── app.go            # Core app + notifications
├── monitor.go        # JSONL parsing logic
├── frontend/
│   ├── index.html    # UI structure
│   ├── style.css     # Styling
│   └── main.js       # Update logic
├── wails.json        # Wails config
├── go.mod            # Go dependencies
└── Makefile          # Build commands
```

## Development

```bash
# Run with hot reload
make dev

# The app will open and refresh when you edit files
```

## Troubleshooting

### "No active Claude sessions"
- Sessions must be modified within the last 60 minutes
- Check that Claude Code is storing sessions in `~/.claude/projects/`

### Notifications not working
- **macOS**: Grant notification permissions in System Preferences
- **Windows**: Check notification settings in Windows Settings
- **Linux**: Ensure `notify-send` is installed

### Build errors
- Run `wails doctor` to diagnose missing dependencies
- Ensure Go and Wails are in your PATH

## License

MIT License - Feel free to modify and distribute.

## Contributing

Pull requests welcome! Ideas for future features:
- [ ] System tray mode
- [ ] Configurable thresholds
- [ ] Session naming/labels
- [ ] Export context summary
- [ ] MCP server integration
