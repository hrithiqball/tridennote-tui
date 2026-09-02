package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// Configure the base URL for your local or production web app
const baseURL = "http://localhost:5173"

type sessionState int

const (
	stateInit sessionState = iota
	stateWaitAuth
	stateEditor
)

type model struct {
	state       sessionState
	editorInput textarea.Model
	err         error

	deviceCode string
	userCode   string
	apiToken   string
}

type initResponse struct {
	DeviceCode string `json:"deviceCode"`
	UserCode   string `json:"userCode"`
}

type pollResponse struct {
	Status string `json:"status"`
	Token  string `json:"token,omitempty"`
}

type tickMsg time.Time

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Write your markdown here..."
	ta.SetWidth(80)
	ta.SetHeight(20)

	return model{
		state:       stateInit,
		editorInput: ta,
	}
}

func initDeviceAuth() tea.Msg {
	resp, err := http.Post(baseURL+"/api/device-auth/init", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data initResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	return data
}

func pollDeviceAuth(deviceCode string) tea.Msg {
	reqBody, _ := json.Marshal(map[string]string{"deviceCode": deviceCode})
	resp, err := http.Post(baseURL+"/api/device-auth/poll", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data pollResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	return data
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	// Automatically start the initialization process
	return initDeviceAuth
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case error:
		m.err = msg
		return m, tea.Quit

	case initResponse:
		m.deviceCode = msg.DeviceCode
		m.userCode = msg.UserCode
		m.state = stateWaitAuth
		cmds = append(cmds, tickCmd())

	case pollResponse:
		if msg.Status == "authorized" {
			m.apiToken = msg.Token
			m.state = stateEditor
			m.editorInput.Focus()
			cmds = append(cmds, textarea.Blink)
		} else if msg.Status == "expired" {
			m.err = fmt.Errorf("Authorization code expired. Please restart.")
			return m, tea.Quit
		} else {
			// keep polling
			cmds = append(cmds, tickCmd())
		}

	case tickMsg:
		if m.state == stateWaitAuth {
			// trigger the HTTP poll, which will return a pollResponse
			cmds = append(cmds, func() tea.Msg {
				return pollDeviceAuth(m.deviceCode)
			})
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.editorInput.SetWidth(msg.Width - 4)
		m.editorInput.SetHeight(msg.Height - 4)
	}

	if m.state == stateEditor {
		var taCmd tea.Cmd
		m.editorInput, taCmd = m.editorInput.Update(msg)
		cmds = append(cmds, taCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress Ctrl+C to quit", m.err)
	}

	switch m.state {
	case stateInit:
		return "Initializing authentication..."
	case stateWaitAuth:
		return fmt.Sprintf(
			"🔐 Authentication Required\n\n"+
				"Please open this URL in your browser:\n"+
				"%s/#/activate\n\n"+
				"And enter the following code:\n\n"+
				"      %s      \n\n"+
				"Waiting for authorization... ⣾\n"+
				"(Press Ctrl+C to quit)",
			baseURL, m.userCode,
		)
	case stateEditor:
		return fmt.Sprintf(
			"Editing Document (Authenticated!):\n\n%s\n\n(Press Ctrl+C to quit)",
			m.editorInput.View(),
		)
	default:
		return "Unknown state"
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
