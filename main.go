package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tui-cerebrum/internal/api"
	"tui-cerebrum/internal/session"
	"tui-cerebrum/internal/tui"
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
		fmt.Fprintf(os.Stderr, "cerebrum: %v\n", err)
		os.Exit(1)
	}
}
