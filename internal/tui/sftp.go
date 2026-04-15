package tui

import (
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/juparave/photoptim/internal/optimizer"
	"github.com/juparave/photoptim/internal/remotefs"
	sftpfs "github.com/juparave/photoptim/internal/sftp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sftpItem represents a file or directory in the SFTP list.
type sftpItem struct {
	name     string
	isDir    bool
	size     int64
	selected bool
}

func (i sftpItem) FilterValue() string { return i.name }
func (i sftpItem) Description() string { return "" }
func (i sftpItem) Title() string {
	icon := fileIcon
	if i.isDir {
		icon = folderIcon
		return fmt.Sprintf("%s %s", icon, i.name)
	}

	sizeStr := formatFileSize(i.size)
	return fmt.Sprintf("%s %-30s %s", icon, i.name, sizeStr)
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fGB", float64(size)/(1024*1024*1024))
	}
}

// Mobile device size presets
var resizePresets = []struct {
	name   string
	width  int
	height int
}{
	{"Disabled", 0, 0},
	{"iPhone 15 Pro Max", 1290, 2796},
	{"iPhone 15/14", 1179, 2556},
	{"Samsung Galaxy S23 Ultra", 1440, 3088},
	{"Google Pixel 7 Pro", 1440, 3120},
	{"iPad Pro 12.9\"", 2048, 2732},
	{"iPad Mini", 1488, 2266},
	{"Full HD", 1920, 1080},
	{"2K QHD", 2560, 1440},
	{"4K UHD", 3840, 2160},
}

// sftpItemDelegate handles rendering SFTP list items.
type sftpItemDelegate struct {
	model *SFTPModel
}

func (d sftpItemDelegate) Height() int                               { return 1 }
func (d sftpItemDelegate) Spacing() int                              { return 0 }
func (d sftpItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d sftpItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sftpItem)
	if !ok {
		return
	}

	// Selection state for files
	selectionIndicator := "   "
	if !i.isDir {
		selectionIndicator = fmt.Sprintf(" %s ", uncheckedIcon)
		if i.selected {
			selectionIndicator = fmt.Sprintf(" %s ", checkedIcon)
		}
	}

	// Style selection
	isFocused := index == m.Index()
	baseStyle := itemStyle
	if i.isDir {
		baseStyle = directoryStyle
	}

	title := i.Title()
	var str string
	if isFocused {
		str = selectedItemStyle.Render(fmt.Sprintf("%s%s %s", arrowIcon, selectionIndicator, title))
	} else {
		str = baseStyle.Render(fmt.Sprintf("  %s %s", selectionIndicator, title))
	}

	fmt.Fprint(w, str)
}

// SFTPState represents the current view of the SFTP TUI.
type SFTPState int

const (
	ConnectionState SFTPState = iota
	BrowserState
)

type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
)

const (
	sizeThresholdMB = 10 // Highlight files larger than 10MB
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
	sortMode      SortMode

	sftpClient    *sftpfs.Client
	fileList      list.Model
	selectedFiles map[string]struct{}

	// Progress tracking for optimization
	progress            progress.Model
	optimizing          bool
	optimizationResults []string
	currentFile         string
	filesProcessed      int
	totalFiles          int
	optimizedCount      int
	failedCount         int

	// Resize parameters
	maxWidth     int
	maxHeight    int
	resizePreset int // 0 = disabled, 1+ = preset index
}

// --- Bubble Tea Messages ---
type (
	sftpConnectSuccessMsg struct {
		client *sftpfs.Client
		path   string
	}
	sftpConnectErrorMsg     struct{ err error }
	filesListedMsg          struct{ files []list.Item }
	fileListErrorMsg        struct{ err error }
	optimizationCompleteMsg struct {
		optimized int
		failed    int
		results   []string
	}
	optimizationErrorMsg    struct{ err error }
	optimizationProgressMsg struct {
		current  int
		total    int
		filename string
		result   string
	}
	startOptimizationMsg struct{ files []string }
	optimizeFileMsg      struct {
		filePath string
		index    int
		total    int
	}
	fileOptimizedMsg struct {
		result  string
		success bool
	}
)

// --- Bubble Tea Commands ---

func (m SFTPModel) connectCmd() tea.Cmd {
	return func() tea.Msg {
		host := m.inputs[0].Value()
		port, _ := strconv.Atoi(m.inputs[1].Value())
		user := m.inputs[2].Value()
		password := m.inputs[3].Value()
		keyPath := m.inputs[4].Value()
		remotePath := m.inputs[5].Value()

		cfg := remotefs.ConnectionConfig{
			Host:       host,
			Port:       port,
			User:       user,
			Password:   password,
			KeyPath:    keyPath,
			RemotePath: remotePath,
		}
		client := &sftpfs.Client{}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := client.Connect(ctx, cfg); err != nil {
			return sftpConnectErrorMsg{err: err}
		}

		// After successful connection, the client.Root() will have the resolved path (home or specified)
		return sftpConnectSuccessMsg{client: client, path: "."}
	}
}

func (m SFTPModel) listFilesCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.sftpClient.List(context.Background(), m.currentPath)
		if err != nil {
			return fileListErrorMsg{err}
		}

		var filteredEntries []remotefs.RemoteEntry
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name, ".") {
				filteredEntries = append(filteredEntries, entry)
			}
		}

		items := make([]sftpItem, 0, len(filteredEntries)+1)

		// Add ".." entry if not at virtual root
		if m.currentPath != "." {
			items = append(items, sftpItem{name: "..", isDir: true})
		}

		for _, entry := range filteredEntries {
			filePath := path.Join(m.currentPath, entry.Name)
			_, isSelected := m.selectedFiles[filePath]
			items = append(items, sftpItem{name: entry.Name, isDir: entry.IsDir, size: entry.Size, selected: isSelected})
		}

		sort.Slice(items, func(i, j int) bool {
			// ".." always first
			if items[i].name == ".." {
				return true
			}
			if items[j].name == ".." {
				return false
			}

			// Directories first
			if items[i].isDir != items[j].isDir {
				return items[i].isDir
			}

			switch m.sortMode {
			case SortBySize:
				if items[i].isDir {
					return items[i].name < items[j].name
				}
				return items[i].size > items[j].size
			default:
				return items[i].name < items[j].name
			}
		})

		listItems := make([]list.Item, len(items))
		for i, item := range items {
			listItems[i] = item
		}
		return filesListedMsg{listItems}
	}
}

func (m SFTPModel) optimizeFileCmd(filePath string, index int, total int) tea.Cmd {
	return func() tea.Msg {
		filename := filepath.Base(filePath)

		if m.sftpClient == nil {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: no SFTP connection", filename),
				success: false,
			}
		}

		ctx := context.Background()
		opt := optimizer.New()
		opt.Quality = 80

		reader, _, err := m.sftpClient.Open(ctx, filePath)
		if err != nil {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: failed to open (%v)", filename, err),
				success: false,
			}
		}

		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: failed to read (%v)", filename, err),
				success: false,
			}
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == "" || (ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp") {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: unsupported format (%s)", filename, ext),
				success: false,
			}
		}

		format := strings.TrimPrefix(ext, ".")
		optimizedData, res, err := opt.OptimizeBytes(data, format, optimizer.Params{
			JPEGQuality: opt.Quality,
			MaxWidth:    m.maxWidth,
			MaxHeight:   m.maxHeight,
		})
		if err != nil && !res.Skipped {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: optimization failed (%v)", filename, err),
				success: false,
			}
		}

		if res.Skipped && res.Reason == "no-compression-gain" {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("ℹ️  %s: original is already optimal", filename),
				success: true,
			}
		}

		if res.Skipped {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: skipped (%s)", filename, res.Reason),
				success: false,
			}
		}

		originalSize := res.OriginalSize
		optimizedSize := res.OptimizedSize

		writer, err := m.sftpClient.Create(ctx, filePath, true)
		if err != nil {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: failed to create output file (%v)", filename, err),
				success: false,
			}
		}

		_, err = writer.Write(optimizedData)
		writer.Close()
		if err != nil {
			return fileOptimizedMsg{
				result:  fmt.Sprintf("❌ %s: failed to write (%v)", filename, err),
				success: false,
			}
		}

		savings := originalSize - optimizedSize
		savingsPercent := float64(savings) / float64(originalSize) * 100
		return fileOptimizedMsg{
			result: fmt.Sprintf("✅ %s: %s -> %s (%.1f%% saved)", filename,
				formatFileSize(originalSize), formatFileSize(optimizedSize), savingsPercent),
			success: true,
		}
	}
}

// --- Model Initialization and Methods ---

func NewSFTPModel() SFTPModel {
	m := SFTPModel{
		state:         ConnectionState,
		focusIndex:    0,
		currentPath:   ".",
		selectedFiles: make(map[string]struct{}),
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(secondaryColor)
	m.spinner = s

	var inputs []textinput.Model = make([]textinput.Model, 6)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "sftp.example.com"
	inputs[0].Focus()
	inputs[0].Prompt = "Host: "
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "22"
	inputs[1].Prompt = "Port: "
	inputs[1].SetValue("22")
	inputs[1].Width = 5

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "username"
	inputs[2].Prompt = "User: "
	inputs[2].Width = 20

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "password (or passphrase for key)"
	inputs[3].Prompt = "Pass: "
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'
	inputs[3].Width = 30

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Key Path (e.g. ~/.ssh/id_rsa)"
	inputs[4].Prompt = "Key:  "
	inputs[4].Width = 30

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Remote Path (optional home or /path)"
	inputs[5].Prompt = "Path: "
	inputs[5].SetValue("")
	inputs[5].Width = 30

	for i := range inputs {
		inputs[i].PromptStyle = primaryColorStyle
	}
	m.inputs = inputs

	l := list.New([]list.Item{}, sftpItemDelegate{model: &m}, 0, 0)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	m.fileList = l

	p := progress.New(progress.WithDefaultGradient())
	m.progress = p

	return m
}

func (m SFTPModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SFTPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h, v := appStyle.GetFrameSize()
		m.fileList.SetSize(m.width-h-4, m.height-v-6)
		m.progress.Width = m.width - h - 10
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case sftpConnectSuccessMsg:
		m.sftpClient = msg.client
		m.currentPath = msg.path
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
		m.updateListTitle()
		return m, nil

	case fileListErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case startOptimizationMsg:
		m.optimizing = true
		m.optimizationResults = []string{}
		m.filesProcessed = 0
		m.totalFiles = len(msg.files)
		m.optimizedCount = 0
		m.failedCount = 0
		m.currentFile = ""
		m.progress.SetPercent(0)
		m.status = "Starting optimization..."

		if len(msg.files) > 0 {
			return m, m.optimizeFileCmd(msg.files[0], 0, len(msg.files))
		}
		return m, nil

	case fileOptimizedMsg:
		m.filesProcessed++
		if msg.success {
			m.optimizedCount++
		} else {
			m.failedCount++
		}

		m.optimizationResults = append(m.optimizationResults, msg.result)
		progressPercent := float64(m.filesProcessed) / float64(m.totalFiles)
		cmd = m.progress.SetPercent(progressPercent)

		if m.filesProcessed >= m.totalFiles {
			m.loading = false
			m.optimizing = false
			if m.failedCount > 0 {
				m.status = fmt.Sprintf("Optimization complete: %d optimized, %d failed", m.optimizedCount, m.failedCount)
			} else {
				m.status = fmt.Sprintf("Optimization complete: %d files optimized successfully", m.optimizedCount)
			}
			m.selectedFiles = make(map[string]struct{})
			return m, tea.Batch(cmd, m.listFilesCmd())
		} else {
			files := m.getSelectedFiles()
			if m.filesProcessed < len(files) {
				return m, tea.Batch(cmd, m.optimizeFileCmd(files[m.filesProcessed], m.filesProcessed, len(files)))
			}
		}
		return m, cmd

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
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
		return m.updateBrowser(msg)
	}

	return m, nil
}

func (m SFTPModel) View() string {
	var content string

	if m.err != nil {
		content = errorColorStyle.Render(fmt.Sprintf("Error: %v\n\n(press any key to return)", m.err))
	} else if m.loading {
		if m.optimizing {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("\n   %s %s\n\n", m.spinner.View(), headerStyle.Render(m.status)))
			b.WriteString(m.progress.View())
			if len(m.optimizationResults) > 0 {
				b.WriteString("\n\nRecent results:\n")
				start := 0
				if len(m.optimizationResults) > 3 {
					start = len(m.optimizationResults) - 3
				}
				for i := start; i < len(m.optimizationResults); i++ {
					b.WriteString(fmt.Sprintf("  %s\n", m.optimizationResults[i]))
				}
			}
			content = b.String()
		} else {
			content = fmt.Sprintf("\n   %s %s\n\n", m.spinner.View(), statusStyle.Render(m.status))
		}
	} else {
		switch m.state {
		case BrowserState:
			content = m.fileList.View()
			preset := resizePresets[m.resizePreset]
			resizeStatus := fmt.Sprintf("Resize: %s", preset.name)
			if m.maxWidth > 0 || m.maxHeight > 0 {
				resizeStatus = fmt.Sprintf("Resize: %s (%dx%d)", preset.name, m.maxWidth, m.maxHeight)
			}
			content += footerStyle.Render(fmt.Sprintf("\n%s (press 'r' to cycle)", resizeStatus))
		default:
			content = m.connectionView()
		}
	}

	return appStyle.Render(content)
}

var errorColorStyle = lipgloss.NewStyle().Foreground(errorColor)

func (m SFTPModel) connectionView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Connect to SFTP Server") + "\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
	}

	button := "  [ Connect ]  "
	if m.focusIndex == len(m.inputs) {
		button = titleStyle.Render("  [ Connect ]  ")
	}

	fmt.Fprintf(&b, "\n%s\n\n", button)
	b.WriteString(inactiveColorStyle.Render("(q to quit, tab to navigate)"))

	return b.String()
}

var inactiveColorStyle = lipgloss.NewStyle().Foreground(inactiveColor)

func (m *SFTPModel) updateConnection(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}
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

	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	}
	return m, cmd
}

func (m *SFTPModel) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.sftpClient != nil {
				m.sftpClient.Close()
			}
			return m, tea.Quit
		case " ":
			i, ok := m.fileList.SelectedItem().(sftpItem)
			if ok && !i.isDir {
				m.toggleFileSelection(i.name)
			}
			return m, nil
		case "enter":
			i, ok := m.fileList.SelectedItem().(sftpItem)
			if !ok {
				return m, nil
			}
			if i.name == ".." {
				m.currentPath = path.Dir(m.currentPath)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			}
			if i.isDir {
				m.currentPath = path.Join(m.currentPath, i.name)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			} else if len(m.selectedFiles) > 0 {
				files := m.getSelectedFiles()
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
					return startOptimizationMsg{files: files}
				})
			}
		case "backspace", "left":
			if m.currentPath != "." {
				m.currentPath = path.Dir(m.currentPath)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			}
		case "s":
			if m.sortMode == SortByName {
				m.sortMode = SortBySize
			} else {
				m.sortMode = SortByName
			}
			m.loading = true
			m.status = "Resorting files..."
			return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
		case "r":
			m.resizePreset = (m.resizePreset + 1) % len(resizePresets)
			preset := resizePresets[m.resizePreset]
			m.maxWidth = preset.width
			m.maxHeight = preset.height
			m.status = fmt.Sprintf("Resize: %s (%dx%d) - press 'r' to cycle", preset.name, preset.width, preset.height)
			return m, nil
		}
	}

	fileList, newCmd := m.fileList.Update(msg)
	m.fileList = fileList
	cmd = newCmd
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
			m.inputs[i].PromptStyle = primaryColorStyle
			continue
		}
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = lipgloss.NewStyle()
	}
	return m, tea.Batch(cmds...)
}

func (m *SFTPModel) toggleFileSelection(filename string) {
	filePath := path.Join(m.currentPath, filename)
	if _, ok := m.selectedFiles[filePath]; ok {
		delete(m.selectedFiles, filePath)
	} else {
		m.selectedFiles[filePath] = struct{}{}
	}

	items := m.fileList.Items()
	newItems := make([]list.Item, len(items))
	for i, item := range items {
		if sftpItem, ok := item.(sftpItem); ok {
			itemFilePath := path.Join(m.currentPath, sftpItem.name)
			_, isSelected := m.selectedFiles[itemFilePath]
			newSftpItem := sftpItem
			newSftpItem.selected = isSelected
			newItems[i] = newSftpItem
		} else {
			newItems[i] = item
		}
	}
	m.fileList.SetItems(newItems)
	m.updateListTitle()
}

func (m *SFTPModel) updateListTitle() {
	sortModeStr := "name"
	if m.sortMode == SortBySize {
		sortModeStr = "size"
	}
	selectedCount := len(m.selectedFiles)
	selectionStr := ""
	if selectedCount > 0 {
		selectionStr = fmt.Sprintf(" | %d selected", selectedCount)
	}
	m.fileList.Title = fmt.Sprintf("Remote Files: %s (sorted by %s)%s - Press 's' to toggle sort, space to select, enter to optimize", m.currentPath, sortModeStr, selectionStr)
}

func (m *SFTPModel) getSelectedFiles() []string {
	files := make([]string, 0, len(m.selectedFiles))
	for file := range m.selectedFiles {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}
