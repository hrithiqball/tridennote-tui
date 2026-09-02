# tui-cerebrum

A Terminal User Interface (TUI) application built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). It provides a simple markdown editor that authenticates via a device authorization flow with a web application.

## Features

- **Device Authentication**: Securely connects to your account using an OAuth-like device flow.
- **Markdown Editor**: Simple, terminal-based markdown editor.
- **Responsive**: Adapts to terminal window resizing.

## Prerequisites

- [Go](https://golang.org/) 1.27.1 or higher

## Building and Running

You can use the provided `Makefile` to build and run the application.

```bash
# Build the binary
make build

# Run the application
make run
```

## How to use

1. Run the application (`make run`).
2. The TUI will display a URL and a code.
3. Open the URL in your browser and enter the code to authenticate.
4. Once authenticated, the TUI will transition to the markdown editor.
5. Press `Ctrl+C` or `Esc` to quit.

## Deployment / Distribution

Since this is a TUI client, deployment typically means distributing the compiled binary to users.

You can build the binary for different platforms using the Makefile:

```bash
make build-linux
make build-mac
make build-windows
```

Then distribute the resulting binaries found in the `bin/` directory.
