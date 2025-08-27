package tui

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"photoptim/internal/remotefs"
	sftpfs "photoptim/internal/sftp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Padding(1, 2)
)

// NOTE: item and itemDelegate definitions are in item.go

// SFTPState represents the current view of the SFTP TUI.
type SFTPState int

const (
	ConnectionState SFTPState = iota
	BrowserState
)

// SFTPModel is the main model for the SFTP TUI.
type SFTPModel struct {
	state         SFTPState
	inputs        []textinput.Model
	focusIndex    int
	spinner       spinner.Model
	loading       bool
	status        string
	currentPath   string
	err           error
	width, height int

	sftpClient *sftpfs.Client
	fileList   list.Model
}

// --- Bubble Tea Messages ---
type sftpConnectSuccessMsg struct{ client *sftpfs.Client }
type sftpConnectErrorMsg struct{ err error }
type filesListedMsg struct{ files []list.Item }
type fileListErrorMsg struct{ err error }

// --- Bubble Tea Commands ---

func (m SFTPModel) connectCmd() tea.Cmd {
	return func() tea.Msg {
		host := m.inputs[0].Value()
		port, _ := strconv.Atoi(m.inputs[1].Value())
		user := m.inputs[2].Value()
		password := m.inputs[3].Value()
		remotePath := m.inputs[4].Value()

		cfg := remotefs.ConnectionConfig{Host: host, Port: port, User: user, Password: password, RemotePath: remotePath}
		client := &sftpfs.Client{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.Connect(ctx, cfg); err != nil {
			return sftpConnectErrorMsg{err: err}
		}

		return sftpConnectSuccessMsg{client: client}
	}
}

func (m SFTPModel) listFilesCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.sftpClient.List(context.Background(), m.currentPath)
		if err != nil {
			return fileListErrorMsg{err}
		}

		items := make([]list.Item, len(entries))
		for i, entry := range entries {
			items[i] = item{name: entry.Name, isDir: entry.IsDir}
		}
		return filesListedMsg{items}
	}
}

// --- Model Initialization and Methods ---

func NewSFTPModel() SFTPModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	var inputs []textinput.Model = make([]textinput.Model, 5)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "sftp.example.com"
	inputs[0].Focus()
	inputs[0].CharLimit = 156
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "22"
	inputs[1].CharLimit = 5
	inputs[1].Width = 5
	inputs[1].SetValue("22")

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "username"
	inputs[2].CharLimit = 56
	inputs[2].Width = 20

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "password"
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'
	inputs[3].CharLimit = 128
	inputs[3].Width = 20

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "/"
	inputs[4].CharLimit = 256
	inputs[4].Width = 30
	inputs[4].SetValue("/")

	l := list.New([]list.Item{}, itemDelegate{}, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return SFTPModel{
		state:       ConnectionState,
		inputs:      inputs,
		focusIndex:  0,
		spinner:     s,
		fileList:    l,
		currentPath: ".",
	}
}

func (m SFTPModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SFTPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		docStyle.Width(m.width)
		m.fileList.SetSize(m.width-4, m.height-4)
		return m, nil

	case sftpConnectSuccessMsg:
		m.sftpClient = msg.client
		m.state = BrowserState
		m.loading = true
		m.status = "Listing files..."
		return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())

	case sftpConnectErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case filesListedMsg:
		m.loading = false
		m.fileList.SetItems(msg.files)
		m.fileList.Title = fmt.Sprintf("Remote Files: %s", m.currentPath)
		return m, nil

	case fileListErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			if m.sftpClient != nil {
				m.sftpClient.Close()
			}
			return m, tea.Quit
		}
	}

	switch m.state {
	case ConnectionState:
		return m.updateConnection(msg)
	case BrowserState:
		m.fileList, cmd = m.fileList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m SFTPModel) View() string {
	if m.err != nil {
		return docStyle.Render(fmt.Sprintf("Error: %v\n\n(press any key to return)", m.err))
	}

	if m.loading {
		return fmt.Sprintf("\n   %s %s\n\n", m.spinner.View(), m.status)
	}

	switch m.state {
	case BrowserState:
		return docStyle.Render(m.fileList.View())
	default:
		return m.connectionView()
	}
}

func (m SFTPModel) updateConnection(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			return m.updateFocus(msg)
		case "enter":
			if m.focusIndex == len(m.inputs) {
				m.loading = true
				m.status = "Connecting..."
				return m, tea.Batch(m.spinner.Tick, m.connectCmd())
			}
			return m.updateFocus(tea.KeyMsg{Type: tea.KeyTab})
		}
	}

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}
	return m, cmd
}

func (m SFTPModel) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			i, ok := m.fileList.SelectedItem().(item)
			if ok && i.isDir {
				m.currentPath = path.Join(m.currentPath, i.name)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			}
		case "backspace":
			if m.currentPath != "." {
				m.currentPath = path.Dir(m.currentPath)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			}
		}
	}

	m.fileList, cmd = m.fileList.Update(msg)
	return m, cmd
}

func (m *SFTPModel) updateFocus(msg tea.Msg) (tea.Model, tea.Cmd) {
	s := msg.(tea.KeyMsg).String()

	if s == "up" || s == "shift+tab" {
		m.focusIndex--
	} else {
		m.focusIndex++
	}

	if m.focusIndex > len(m.inputs) {
		m.focusIndex = 0
	} else if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs)
	}

	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			continue
		}
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = lipgloss.NewStyle()
	}

	return m, tea.Batch(cmds...)
}

func (m SFTPModel) connectionView() string {
	var b strings.Builder

	b.WriteString("Connect to SFTP Server\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
	}

	button := "[ Connect ]"
	if m.focusIndex != len(m.inputs) {
		button = "  Connect  "
	}

	fmt.Fprintf(&b, "\n%s\n\n", button)
	b.WriteString("(q to quit, tab to navigate)\n")

	return docStyle.Render(b.String())
}
