package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"tui-cerebrum/internal/api"
	"tui-cerebrum/internal/session"
)

const (
	cookieName = "__Secure-neon-auth.session_token"
	token      = "session-abc"
)

type fakeServer struct {
	mu         sync.Mutex
	authorized bool
}

func (f *fakeServer) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, data any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}
	authed := func(r *http.Request) bool {
		c, err := r.Cookie(cookieName)
		return err == nil && c.Value == token
	}
	mux.HandleFunc("POST /api/device-auth/init", func(w http.ResponseWriter, r *http.Request) {
		reply(w, api.DeviceCode{DeviceCode: "dev-1", UserCode: "WXYZ-2345", VerificationURL: "https://app.brain.pixcel.org/#/activate?code=WXYZ-2345", Interval: 3})
	})
	mux.HandleFunc("POST /api/device-auth/poll", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.authorized {
			reply(w, api.DevicePoll{Status: "authorized", SessionToken: token, CookieName: cookieName})
			return
		}
		reply(w, api.DevicePoll{Status: "pending"})
	})
	mux.HandleFunc("GET /api/auth/get-session", func(w http.ResponseWriter, r *http.Request) {
		if !authed(r) {
			_, _ = w.Write([]byte("null"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"user": api.User{ID: "u1", Email: "harith@example.com"}})
	})
	mux.HandleFunc("GET /api/workspaces", func(w http.ResponseWriter, r *http.Request) {
		if !authed(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		reply(w, []api.Workspace{{ID: "w1", Name: "Personal", IsDefault: true}})
	})
	mux.HandleFunc("GET /api/workspaces/w1/tree", func(w http.ResponseWriter, r *http.Request) {
		note, sketch, pdf := "note", "excalidraw", "file"
		folder := "f1"
		reply(w, []api.TreeNode{
			{ID: "n1", Kind: "file", FileType: &note, Name: "Daily log"},
			{ID: "f1", Kind: "folder", Name: "Projects"},
			{ID: "n2", Kind: "file", FileType: &note, Name: "Roadmap", ParentID: &folder},
			{ID: "x1", Kind: "file", FileType: &sketch, Name: "Whiteboard.excalidraw", ParentID: &folder},
			{ID: "p1", Kind: "file", FileType: &pdf, Name: "invoice.pdf"},
		})
	})
	mux.HandleFunc("GET /api/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		reply(w, api.Note{Title: "Roadmap", Blocks: []api.Block{
			{Type: "h2", Content: "Q4"},
			{Type: "bulleted", Content: "Ship the terminal client"},
			{Type: "bulleted", Content: "Neovim-style tree"},
		}})
	})
	return mux
}

func step(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

func key(k string) tea.KeyMsg {
	switch k {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func snapshot(t *testing.T, name string, m Model) string {
	t.Helper()
	view := ansi.Strip(m.View())
	if dir := os.Getenv("TUI_SNAPSHOT_DIR"); dir != "" {
		_ = os.WriteFile(filepath.Join(dir, name+".txt"), []byte(view), 0o644)
	}
	return view
}

func TestSignInAndBrowse(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	fake := &fakeServer{}
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	m := New(api.New(srv.URL))
	m, _ = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 24})

	m, cmd := step(t, m, m.checkSession()())
	if m.state != stateLogin {
		t.Fatalf("expected login state, got %v", m.state)
	}
	m, _ = step(t, m, cmd())
	if m.device == nil || m.device.UserCode != "WXYZ-2345" {
		t.Fatalf("device code not shown: %+v", m.device)
	}
	if view := snapshot(t, "1-login", m); !strings.Contains(view, "WXYZ-2345") || !strings.Contains(view, "not signed in") {
		t.Fatalf("login view missing prompt:\n%s", view)
	}

	m, _ = step(t, m, m.poll("dev-1")())
	if m.state != stateLogin {
		t.Fatalf("pending poll should stay on login")
	}

	fake.authorized = true
	m, cmd = step(t, m, m.poll("dev-1")())
	saved, err := session.Load()
	if err != nil || saved == nil || saved.Token != token || saved.CookieName != cookieName {
		t.Fatalf("session not persisted: %+v %v", saved, err)
	}
	m, cmd = step(t, m, cmd())
	m, cmd = step(t, m, cmd())
	m, _ = step(t, m, cmd())
	if m.state != stateBrowse {
		t.Fatalf("expected browse state, got %v err=%v", m.state, m.err)
	}

	view := snapshot(t, "2-tree", m)
	for _, want := range []string{"Projects/", "Daily log.md", "harith@example.com"} {
		if !strings.Contains(view, want) {
			t.Fatalf("tree missing %q:\n%s", want, view)
		}
	}
	for _, hidden := range []string{"invoice", "Whiteboard"} {
		if strings.Contains(view, hidden) {
			t.Fatalf("non-markdown file %q should be hidden:\n%s", hidden, view)
		}
	}

	m, _ = step(t, m, key("right"))
	if !m.visible[0].Expanded {
		t.Fatalf("right arrow should expand folder")
	}
	m, _ = step(t, m, key("down"))
	if m.selected().ID != "n2" {
		t.Fatalf("down should land on Roadmap, got %s", m.selected().ID)
	}
	m, cmd = step(t, m, key("enter"))
	m, _ = step(t, m, cmd())
	view = snapshot(t, "3-note", m)
	if !strings.Contains(view, "Ship the terminal client") {
		t.Fatalf("note content not rendered:\n%s", view)
	}

	m, _ = step(t, m, key("tab"))
	m, _ = step(t, m, key("left"))
	if m.selected().ID != "f1" || m.visible[0].Expanded {
		t.Fatalf("left on a child should collapse parent and select it")
	}

	m, _ = step(t, m, key("L"))
	if m.state != stateLogin {
		t.Fatalf("logout should return to login")
	}
	if saved, _ := session.Load(); saved != nil {
		t.Fatalf("logout should clear saved session")
	}
}

func TestExpiredSessionPromptsLogin(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := httptest.NewServer((&fakeServer{}).handler(t))
	defer srv.Close()

	client := api.New(srv.URL)
	client.Token = "revoked"
	client.CookieName = cookieName
	m := New(client)
	m, _ = step(t, m, m.checkSession()())
	if m.state != stateLogin {
		t.Fatalf("revoked session should prompt login, got %v", m.state)
	}
}
