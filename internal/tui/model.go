package tui

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"tui-cerebrum/internal/api"
	"tui-cerebrum/internal/session"
	"tui-cerebrum/internal/tree"
)

type state int

const (
	stateBoot state = iota
	stateLogin
	stateLoading
	stateBrowse
)

type focus int

const (
	focusTree focus = iota
	focusViewer
)

type Model struct {
	client *api.Client
	state  state
	width  int
	height int
	err    error

	spinner spinner.Model

	device        *api.DeviceCode
	browserOpened bool

	user       *api.User
	workspaces []api.Workspace
	workspace  int
	roots      []*tree.Node
	visible    []*tree.Node
	cursor     int
	offset     int
	focus      focus

	viewer      viewport.Model
	notes       map[string]api.Note
	rendered    map[string]string
	openID      string
	loadingNote string
	previewSeq  int
}

func New(client *api.Client) Model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = accentStyle
	return Model{
		client:   client,
		spinner:  sp,
		viewer:   viewport.New(0, 0),
		notes:    map[string]api.Note{},
		rendered: map[string]string{},
	}
}

type sessionCheckedMsg struct {
	user *api.User
	err  error
}

type deviceStartedMsg struct {
	code api.DeviceCode
	err  error
}

type pollTickMsg struct{ deviceCode string }

type pollResultMsg struct {
	deviceCode string
	result     api.DevicePoll
	err        error
}

type workspacesMsg struct {
	workspaces []api.Workspace
	err        error
}

type treeMsg struct {
	workspaceID string
	rows        []api.TreeNode
	err         error
}

type previewTickMsg struct{ seq int }

type noteMsg struct {
	id   string
	note api.Note
	err  error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkSession())
}

func (m Model) checkSession() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		if client.Token == "" {
			return sessionCheckedMsg{err: api.ErrUnauthorized}
		}
		user, err := client.CurrentUser()
		return sessionCheckedMsg{user: user, err: err}
	}
}

func (m Model) startDevice() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		code, err := client.StartDeviceAuth()
		return deviceStartedMsg{code: code, err: err}
	}
}

func pollAfter(deviceCode string, interval int) tea.Cmd {
	if interval <= 0 {
		interval = 3
	}
	return tea.Tick(time.Duration(interval)*time.Second, func(time.Time) tea.Msg {
		return pollTickMsg{deviceCode: deviceCode}
	})
}

func (m Model) poll(deviceCode string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		result, err := client.PollDeviceAuth(deviceCode)
		return pollResultMsg{deviceCode: deviceCode, result: result, err: err}
	}
}

func (m Model) loadWorkspaces() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		workspaces, err := client.Workspaces()
		return workspacesMsg{workspaces: workspaces, err: err}
	}
}

func (m Model) loadTree() tea.Cmd {
	if len(m.workspaces) == 0 {
		return nil
	}
	client := m.client
	id := m.workspaces[m.workspace].ID
	return func() tea.Msg {
		rows, err := client.Tree(id)
		return treeMsg{workspaceID: id, rows: rows, err: err}
	}
}

func (m Model) loadNote(id string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		note, err := client.Note(id)
		return noteMsg{id: id, note: note, err: err}
	}
}

func (m *Model) signedOut() tea.Cmd {
	_ = session.Clear()
	m.client.Token = ""
	m.user = nil
	m.device = nil
	m.browserOpened = false
	m.state = stateLogin
	m.roots, m.visible = nil, nil
	m.notes = map[string]api.Note{}
	m.rendered = map[string]string{}
	m.openID = ""
	return m.startDevice()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)

	case sessionCheckedMsg:
		if msg.err != nil {
			if errors.Is(msg.err, api.ErrUnauthorized) {
				return m, m.signedOut()
			}
			m.err = msg.err
			return m, nil
		}
		m.user = msg.user
		m.state = stateLoading
		return m, m.loadWorkspaces()

	case deviceStartedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.device = &msg.code
		return m, pollAfter(msg.code.DeviceCode, msg.code.Interval)

	case pollTickMsg:
		if m.state != stateLogin || m.device == nil || m.device.DeviceCode != msg.deviceCode {
			return m, nil
		}
		return m, m.poll(msg.deviceCode)

	case pollResultMsg:
		if m.state != stateLogin || m.device == nil || m.device.DeviceCode != msg.deviceCode {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err
			return m, pollAfter(msg.deviceCode, m.device.Interval*2)
		}
		m.err = nil
		switch msg.result.Status {
		case "authorized":
			m.client.Token = msg.result.SessionToken
			m.client.CookieName = msg.result.CookieName
			if err := session.Save(session.Session{Token: m.client.Token, CookieName: m.client.CookieName}); err != nil {
				m.err = err
			}
			m.state = stateBoot
			return m, m.checkSession()
		case "expired":
			m.device = nil
			m.browserOpened = false
			return m, m.startDevice()
		default:
			return m, pollAfter(msg.deviceCode, m.device.Interval)
		}

	case workspacesMsg:
		if msg.err != nil {
			return m.handleAPIError(msg.err)
		}
		m.workspaces = msg.workspaces
		m.workspace = 0
		for i, w := range m.workspaces {
			if w.IsDefault {
				m.workspace = i
			}
		}
		return m, m.loadTree()

	case treeMsg:
		if msg.err != nil {
			return m.handleAPIError(msg.err)
		}
		if len(m.workspaces) == 0 || m.workspaces[m.workspace].ID != msg.workspaceID {
			return m, nil
		}
		expanded := tree.ExpandedIDs(m.roots)
		selected := ""
		if node := m.selected(); node != nil {
			selected = node.ID
		}
		m.roots = tree.Build(msg.rows)
		tree.Restore(m.roots, expanded)
		m.visible = tree.Visible(m.roots)
		m.cursor = 0
		for i, n := range m.visible {
			if n.ID == selected {
				m.cursor = i
			}
		}
		m.state = stateBrowse
		m.err = nil
		m.layout()
		return m, m.schedulePreview()

	case previewTickMsg:
		if msg.seq != m.previewSeq {
			return m, nil
		}
		node := m.selected()
		if node == nil || node.IsFolder {
			return m, nil
		}
		return m, m.open(node.ID)

	case noteMsg:
		if m.loadingNote == msg.id {
			m.loadingNote = ""
		}
		if msg.err != nil {
			return m.handleAPIError(msg.err)
		}
		m.notes[msg.id] = msg.note
		if m.openID == msg.id {
			m.refreshViewer(true)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleAPIError(err error) (tea.Model, tea.Cmd) {
	if errors.Is(err, api.ErrUnauthorized) {
		return m, m.signedOut()
	}
	m.err = err
	return m, nil
}

func (m *Model) open(id string) tea.Cmd {
	if m.openID != id {
		m.openID = id
		m.refreshViewer(true)
	}
	if _, cached := m.notes[id]; cached || m.loadingNote == id {
		return nil
	}
	m.loadingNote = id
	return m.loadNote(id)
}

func (m *Model) schedulePreview() tea.Cmd {
	m.previewSeq++
	seq := m.previewSeq
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg { return previewTickMsg{seq: seq} })
}

func (m Model) selected() *tree.Node {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil
	}
	return m.visible[m.cursor]
}

func (m *Model) moveCursor(to int) tea.Cmd {
	if len(m.visible) == 0 {
		return nil
	}
	to = max(0, min(to, len(m.visible)-1))
	if to == m.cursor {
		return nil
	}
	m.cursor = to
	m.ensureCursorVisible()
	return m.schedulePreview()
}

func (m *Model) relayoutTree(keep *tree.Node) {
	m.visible = tree.Visible(m.roots)
	for i, n := range m.visible {
		if n == keep {
			m.cursor = i
		}
	}
	m.ensureCursorVisible()
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return *m, tea.Quit
	}

	switch m.state {
	case stateLogin:
		switch key {
		case "q", "esc":
			return *m, tea.Quit
		case "enter", "o":
			if m.device != nil {
				m.browserOpened = true
				if err := openBrowser(m.device.VerificationURL); err != nil {
					m.err = err
				}
			}
		case "r":
			if m.err != nil {
				m.err = nil
				return *m, m.startDevice()
			}
		}
		return *m, nil
	case stateBrowse:
		return m.handleBrowseKey(key, msg)
	default:
		if key == "q" {
			return *m, tea.Quit
		}
		if key == "r" && m.err != nil {
			m.err = nil
			m.state = stateBoot
			return *m, m.checkSession()
		}
	}
	return *m, nil
}

func (m *Model) handleBrowseKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return *m, tea.Quit
	case "L":
		return *m, m.signedOut()
	case "r":
		m.notes = map[string]api.Note{}
		m.rendered = map[string]string{}
		return *m, m.loadTree()
	case "w":
		if len(m.workspaces) > 1 {
			m.workspace = (m.workspace + 1) % len(m.workspaces)
			m.roots, m.visible = nil, nil
			m.cursor, m.offset = 0, 0
			m.openID = ""
			m.refreshViewer(true)
			return *m, m.loadTree()
		}
		return *m, nil
	case "tab":
		if m.focus == focusTree {
			m.focus = focusViewer
		} else {
			m.focus = focusTree
		}
		return *m, nil
	case "pgdown", "ctrl+d":
		m.viewer.HalfPageDown()
		return *m, nil
	case "pgup", "ctrl+u":
		m.viewer.HalfPageUp()
		return *m, nil
	}

	if m.focus == focusViewer {
		switch key {
		case "esc", "left", "h":
			m.focus = focusTree
			return *m, nil
		case "g", "home":
			m.viewer.GotoTop()
			return *m, nil
		case "G", "end":
			m.viewer.GotoBottom()
			return *m, nil
		}
		var cmd tea.Cmd
		m.viewer, cmd = m.viewer.Update(msg)
		return *m, cmd
	}

	node := m.selected()
	switch key {
	case "up", "k":
		return *m, m.moveCursor(m.cursor - 1)
	case "down", "j":
		return *m, m.moveCursor(m.cursor + 1)
	case "g", "home":
		return *m, m.moveCursor(0)
	case "G", "end":
		return *m, m.moveCursor(len(m.visible) - 1)
	case "right", "l":
		if node == nil {
			return *m, nil
		}
		if node.IsFolder {
			if !node.Expanded {
				node.Expanded = true
				m.relayoutTree(node)
				return *m, nil
			}
			if len(node.Children) > 0 {
				return *m, m.moveCursor(m.cursor + 1)
			}
			return *m, nil
		}
		m.focus = focusViewer
		return *m, m.open(node.ID)
	case "left", "h":
		if node == nil {
			return *m, nil
		}
		if node.IsFolder && node.Expanded {
			node.Expanded = false
			m.relayoutTree(node)
			return *m, nil
		}
		if node.Parent != nil {
			node.Parent.Expanded = false
			m.relayoutTree(node.Parent)
			return *m, m.schedulePreview()
		}
		return *m, nil
	case "enter", " ":
		if node == nil {
			return *m, nil
		}
		if node.IsFolder {
			node.Expanded = !node.Expanded
			m.relayoutTree(node)
			return *m, nil
		}
		m.focus = focusViewer
		return *m, m.open(node.ID)
	}
	return *m, nil
}
