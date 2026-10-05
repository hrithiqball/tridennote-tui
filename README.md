# tui-cerebrum

A Terminal User Interface (TUI) application built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). It provides a simple markdown editor that authenticates via a device authorization flow with a web application.

## Features

- **Browser sign-in**: on first launch you're prompted to sign in. Press `enter` and the
  app opens `https://app.brain.pixcel.org` (always production, even for local builds) to
  confirm a short code. The resulting Neon Auth session is saved to
  `<user config dir>/cerebrum/session.json` (mode `0600`) and reused until it expires or
  you log out.
- **Vault tree**: a Neovim-style file tree on the right lists your folders and markdown
  notes only (sketches, PDFs and other attachments are hidden).
- **Note preview**: the selected note is rendered as markdown on the left.

## Prerequisites

- [Go](https://golang.org/) 1.27.1 or higher

## Building and Running

```bash
make build   # builds bin/tui-cerebrum
make run     # builds and runs
go test ./...
```

## Keys

| Key                     | Action                                      |
| ----------------------- | ------------------------------------------- |
| `↑` `↓` / `k` `j`       | Move through the tree (previews the note)   |
| `→` / `l`               | Expand folder, or open note                 |
| `←` / `h`               | Collapse folder, or jump to parent          |
| `enter`                 | Toggle folder / open note                   |
| `tab`                   | Switch focus between tree and note          |
| `pgup` `pgdn` / `ctrl+u` `ctrl+d` | Scroll the note                   |
| `w`                     | Next workspace                              |
| `r`                     | Refresh                                     |
| `L`                     | Log out (forget the saved session)          |
| `q` / `ctrl+c`          | Quit                                        |

## Deployment / Distribution

Since this is a TUI client, deployment typically means distributing the compiled binary to users.

You can build the binary for different platforms using the Makefile:

```bash
make build-linux
make build-mac
make build-windows
```

Then distribute the resulting binaries found in the `bin/` directory.
