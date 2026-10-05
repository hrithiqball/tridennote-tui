package diagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrNoRenderer = errors.New("mermaid-cli (mmdc) is not installed — run: npm i -g @mermaid-js/mermaid-cli")

func RenderImages(sources []string) ([]string, error) {
	mmdc, err := exec.LookPath("mmdc")
	if err != nil {
		return nil, ErrNoRenderer
	}
	dir := filepath.Join(os.TempDir(), "tridennote-mermaid")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	baseArgs := []string{"-t", "dark", "-b", "#16161e", "-s", "2", "-q"}
	if config, err := puppeteerConfig(dir); err == nil && config != "" {
		baseArgs = append(baseArgs, "-p", config)
	}
	var images []string
	for i, source := range sources {
		input := filepath.Join(dir, fmt.Sprintf("diagram-%d.mmd", i+1))
		output := filepath.Join(dir, fmt.Sprintf("diagram-%d.png", i+1))
		if err := os.WriteFile(input, []byte(source), 0o600); err != nil {
			return nil, err
		}
		cmd := exec.Command(mmdc, append([]string{"-i", input, "-o", output}, baseArgs...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("mmdc failed on diagram %d: %s", i+1, firstLine(string(out), err))
		}
		images = append(images, output)
	}
	return images, nil
}

var macBrowsers = []string{
	"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	"/Applications/Chromium.app/Contents/MacOS/Chromium",
	"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
	"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
	"/Applications/Arc.app/Contents/MacOS/Arc",
}

var pathBrowsers = []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "brave-browser", "microsoft-edge"}

func browserPath() string {
	if path := os.Getenv("PUPPETEER_EXECUTABLE_PATH"); path != "" {
		return ""
	}
	if runtime.GOOS == "darwin" {
		for _, p := range macBrowsers {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
		}
	}
	for _, name := range pathBrowsers {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func puppeteerConfig(dir string) (string, error) {
	browser := browserPath()
	if browser == "" {
		return "", nil
	}
	data, err := json.Marshal(map[string]any{"executablePath": browser, "headless": "shell"})
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "puppeteer.json")
	return path, os.WriteFile(path, data, 0o600)
}

func firstLine(out string, err error) string {
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return err.Error()
}

type Viewer struct {
	Name     string
	Command  *exec.Cmd
	Terminal bool
}

const holdScript = `for f in "$@"; do %s "$f"; echo; done; printf '\n  press enter to return to tridennote '; read _`

func ImageViewer(images []string) (Viewer, error) {
	if len(images) == 0 {
		return Viewer{}, errors.New("no diagrams to show")
	}
	if _, err := exec.LookPath("chafa"); err == nil && runtime.GOOS != "windows" {
		args := append([]string{"-c", fmt.Sprintf(holdScript, "chafa"), "sh"}, images...)
		return Viewer{Name: "chafa", Command: exec.Command("sh", args...), Terminal: true}, nil
	}
	if _, err := exec.LookPath("kitten"); err == nil && isKitty() && runtime.GOOS != "windows" {
		args := append([]string{"-c", fmt.Sprintf(holdScript, "kitten icat"), "sh"}, images...)
		return Viewer{Name: "kitty", Command: exec.Command("sh", args...), Terminal: true}, nil
	}
	switch runtime.GOOS {
	case "darwin":
		return Viewer{Name: "Preview", Command: exec.Command("open", images...)}, nil
	case "windows":
		return Viewer{Name: "image viewer", Command: exec.Command("cmd", "/c", "start", "", images[0])}, nil
	default:
		return Viewer{Name: "image viewer", Command: exec.Command("xdg-open", images[0])}, nil
	}
}

func isKitty() bool {
	return os.Getenv("KITTY_WINDOW_ID") != "" || strings.Contains(os.Getenv("TERM"), "kitty") || os.Getenv("TERM_PROGRAM") == "ghostty"
}
