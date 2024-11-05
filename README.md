# mac-vim-switch

> [!NOTE]
> This project is no longer maintained.
> Because high latency when switching input methods,
> I recommend using [MacVimSwitch](https://github.com/jackiexiao/macvimswitch) instead. not mac-vim-switch.

[中文版](README.zh-CN.md)
A macOS input method switcher designed for Vim users, built on top of macism.

## Features

- Automatically switch to ABC input method when pressing ESC
- Toggle between ABC and WeType Pinyin input methods with Shift key
- Runs as a background service
- Easy integration with macOS system

## Prerequisites

- macOS
- [macism](https://github.com/laishulu/macism)

## Installation

```bash
# Install macism first
brew tap laishulu/homebrew
brew install macism

# Install mac-vim-switch
brew tap jackiexiao/mac-vim-switch
brew install mac-vim-switch

# Start the service
brew services start mac-vim-switch
```

## System Requirements

1. Grant Accessibility Permission
   - Go to System Preferences > Security & Privacy > Privacy > Accessibility
   - Click the lock icon to make changes
   - Add mac-vim-switch to the list of allowed apps
   - Check the checkbox next to mac-vim-switch

2. System Preferences > Keyboard > Shortcuts > Input Sources
   - Disable "Select the previous input source"
   - Disable "Select next source in Input menu"

## Usage

The service will automatically start after installation and run in the background. You can:

- Press ESC to switch to ABC input method
- Press Shift to toggle between ABC and WeType Pinyin

### Service Management

```bash
# Start the service
brew services start mac-vim-switch

# Stop the service
brew services stop mac-vim-switch

# Restart the service
brew services restart mac-vim-switch

# Check service status
brew services list
```

### Checking Available Input Methods

If you want to use different input methods, you can check available input method IDs:

```bash
macism
```

### Logs

Logs are stored in `~/.mac-vim-switch.log`

### Configuration

You can configure input methods using the following commands:

```bash
# List available input methods
mac-vim-switch list

# Set primary input method (default: com.apple.keylayout.ABC)
mac-vim-switch config primary "com.apple.keylayout.ABC"

# Set secondary input method (default: com.tencent.inputmethod.wetype.pinyin, use `macism` to get current input method)
mac-vim-switch config secondary "your.input.method.id"
```

Configuration is stored in `~/.config/mac-vim-switch/config.json`

## For Developers

### Building from Source

1. Clone the repository
```bash
git clone https://github.com/jackiexiao/mac-vim-switch.git
cd mac-vim-switch
```

2. Install dependencies:
   - Go 1.16 or later
   - macism (`brew install macism`)
   - Xcode Command Line Tools (for CGo compilation)

3. Build the project:
```bash
go build ./cmd/mac-vim-switch
```

### Development and Debugging

1. Run in debug mode with logs:
```bash
# Build and run locally
go build ./cmd/mac-vim-switch
./mac-vim-switch

# Check logs in real-time
tail -f ~/.mac-vim-switch.log
```

2. Test different commands:
```bash
# Check version
./mac-vim-switch --version

# List available input methods
./mac-vim-switch list

# Check current configuration
./mac-vim-switch config

# Run health check
./mac-vim-switch health
```

3. Debug CGo and keyboard events:
```bash
# Build with debug symbols
go build -gcflags="all=-N -l" ./cmd/mac-vim-switch

# Run with verbose CGo logging
GODEBUG=cgocheck=2 ./mac-vim-switch

# Check if keyboard events are being captured
log stream --predicate 'process == "mac-vim-switch"'
```

4. Common development tasks:
   - Modify input method behavior: Edit `switchToInputMethod()` in main.go
   - Add new commands: Add cases to the main command switch
   - Modify keyboard handling: Edit the CGo callback in main.go
   - Change default settings: Modify the const values at the top of main.go

5. Testing installation:
```bash
# Build and install locally
go build ./cmd/mac-vim-switch
sudo cp mac-vim-switch /usr/local/bin/

# Test as a service
brew services stop mac-vim-switch  # Stop existing service if running
./mac-vim-switch                  # Run directly to see logs
```

### Troubleshooting Development Issues

1. CGo compilation errors:
   - Ensure Xcode Command Line Tools are installed: `xcode-select --install`
   - Check CGo flags in main.go
   - Try cleaning the build: `go clean -cache`

2. Keyboard event issues:
   - Check Accessibility permissions in System Preferences
   - Run with sudo to test permissions: `sudo ./mac-vim-switch`
   - Enable debug logging: `GODEBUG=cgocheck=2 ./mac-vim-switch`

3. Input method switching issues:
   - Test macism directly: `macism "com.apple.keylayout.ABC"`
   - Check available methods: `macism`
   - Verify permissions: `mac-vim-switch health`

### Project Structure

- `cmd/mac-vim-switch/main.go`: Main program
  - CGo bindings for keyboard events
  - Input method switching logic
  - Configuration management
- `Formula/mac-vim-switch.rb`: Homebrew formula
- `mac-vim-switch.plist`: LaunchAgent configuration
- `.config/mac-vim-switch/config.json`: Runtime configuration

### Making Changes

1. Update version number in:
   - `main.go` (`version` constant)
   - `Formula/mac-vim-switch.rb`

2. Test changes:
   - Build and run locally
   - Check logs for errors
   - Verify keyboard events
   - Test configuration changes

3. Before committing:
   - Run `go fmt ./...`
   - Run `go vet ./...`
   - Test on a clean system

## Troubleshooting

1. If the service doesn't work:
   - Check if accessibility permission is granted
   - Check the logs: `cat ~/.mac-vim-switch.log`
   - Try restarting the service: `brew services restart mac-vim-switch`

2. If input method switching doesn't work:
   - Run `macism` to check available input methods
   - Make sure the input methods are correctly installed

3. Run health check:
```bash
mac-vim-switch health
```

This will check:
- If macism is properly installed
- If configuration files are accessible
- If log files are writable

## License

MIT License