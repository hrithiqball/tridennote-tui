package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tridennote/internal/api"
	"tridennote/internal/session"
	"tridennote/internal/tui"
)

var baseURL = api.DefaultBaseURL

func main() {
	client := api.New(baseURL)

	saved, err := session.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read saved session: %v\n", err)
	}
	if saved != nil {
		client.Token = saved.Token
		client.CookieName = saved.CookieName
	}
	client.OnRotate = func(token string) {
		_ = session.Save(session.Session{Token: token, CookieName: client.CookieName})
	}

	program := tea.NewProgram(tui.New(client), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tridennote: %v\n", err)
		os.Exit(1)
	}
}
