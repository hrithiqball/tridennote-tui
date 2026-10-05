package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type editorOption struct {
	ID       string
	Name     string
	Bins     []string
	MacPaths []string
	Wait     string
	Terminal bool
	Color    lipgloss.Color
	Nerd     string
}

const builtinEditor = "builtin"

var editorCatalog = []editorOption{
	{ID: builtinEditor, Name: "Built-in", Color: violet, Nerd: ""},
	{ID: "vscode", Name: "VS Code", Bins: []string{"code"}, MacPaths: []string{"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"}, Wait: "--wait", Color: lipgloss.Color("#3b9cff"), Nerd: ""},
	{ID: "vscode-insiders", Name: "VS Code Insiders", Bins: []string{"code-insiders"}, MacPaths: []string{"/Applications/Visual Studio Code - Insiders.app/Contents/Resources/app/bin/code-insiders"}, Wait: "--wait", Color: lipgloss.Color("#24bfa5"), Nerd: ""},
	{ID: "cursor", Name: "Cursor", Bins: []string{"cursor"}, MacPaths: []string{"/Applications/Cursor.app/Contents/Resources/app/bin/cursor"}, Wait: "--wait", Color: lipgloss.Color("#e5e5e5"), Nerd: ""},
	{ID: "zed", Name: "Zed", Bins: []string{"zed", "zeditor"}, MacPaths: []string{"/Applications/Zed.app/Contents/MacOS/cli"}, Wait: "--wait", Color: lipgloss.Color("#6ea8ff"), Nerd: ""},
	{ID: "nvim", Name: "Neovim", Bins: []string{"nvim"}, Terminal: true, Color: lipgloss.Color("#57a143"), Nerd: ""},
	{ID: "vim", Name: "Vim", Bins: []string{"vim"}, Terminal: true, Color: lipgloss.Color("#3fb950"), Nerd: ""},
	{ID: "nano", Name: "Nano", Bins: []string{"nano"}, Terminal: true, Color: lipgloss.Color("#b48ead"), Nerd: ""},
	{ID: "system", Name: "$EDITOR", Color: muted, Nerd: ""},
}

var waitFlags = map[string]string{
	"code":          "--wait",
	"code-insiders": "--wait",
	"codium":        "--wait",
	"cursor":        "--wait",
	"subl":          "--wait",
	"zed":           "--wait",
	"zeditor":       "--wait",
	"cli":           "--wait",
	"mate":          "-w",
}

var externalFallbackOrder = []string{"vscode", "cursor", "zed", "vscode-insiders", "nvim", "vim", "nano"}

func findEditor(id string) (editorOption, bool) {
	for _, e := range editorCatalog {
		if e.ID == id {
			return e, true
		}
	}
	return editorOption{}, false
}

func systemEditorCommand() []string {
	for _, env := range []string{"TRIDENNOTE_EDITOR", "VISUAL", "EDITOR"} {
		if fields := splitCommand(os.Getenv(env)); len(fields) > 0 {
			return withWaitFlag(fields)
		}
	}
	return nil
}

func splitCommand(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if info, err := os.Stat(value); err == nil && !info.IsDir() {
		return []string{value}
	}
	var fields []string
	var current strings.Builder
	var quote rune
	inField := false
	for _, r := range value {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote == 0 && (r == '"' || r == '\''):
			quote = r
			inField = true
		case quote == 0 && (r == ' ' || r == '\t'):
			if inField {
				fields = append(fields, current.String())
				current.Reset()
				inField = false
			}
		default:
			current.WriteRune(r)
			inField = true
		}
	}
	if inField {
		fields = append(fields, current.String())
	}
	return fields
}

func withWaitFlag(fields []string) []string {
	name := strings.ToLower(fields[0])
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}
	for _, ext := range []string{".exe", ".cmd", ".bat"} {
		name = strings.TrimSuffix(name, ext)
	}
	flag, ok := waitFlags[name]
	if !ok {
		return fields
	}
	for _, f := range fields[1:] {
		if f == flag || f == "-w" || f == "--wait" {
			return fields
		}
	}
	return append(fields, flag)
}

func (e editorOption) command() []string {
	switch e.ID {
	case builtinEditor:
		return nil
	case "system":
		return systemEditorCommand()
	}
	for _, bin := range e.Bins {
		if path, err := exec.LookPath(bin); err == nil {
			return e.withWait(path)
		}
	}
	if runtime.GOOS == "darwin" {
		for _, p := range e.MacPaths {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return e.withWait(p)
			}
		}
	}
	return nil
}

func (e editorOption) withWait(path string) []string {
	if e.Wait == "" {
		return []string{path}
	}
	return []string{path, e.Wait}
}

func (e editorOption) available() bool {
	return e.ID == builtinEditor || e.command() != nil
}

func externalEditorCommand(preferred string) ([]string, string) {
	if e, ok := findEditor(preferred); ok && e.ID != builtinEditor {
		if cmd := e.command(); cmd != nil {
			return cmd, e.Name
		}
	}
	if cmd := systemEditorCommand(); cmd != nil {
		return cmd, filepath.Base(cmd[0])
	}
	for _, id := range externalFallbackOrder {
		e, _ := findEditor(id)
		if cmd := e.command(); cmd != nil {
			return cmd, e.Name
		}
	}
	return nil, ""
}
