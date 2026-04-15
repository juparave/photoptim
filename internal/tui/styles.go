package tui

import "github.com/charmbracelet/lipgloss"

// Theme colors using AdaptiveColor for light/dark mode support
var (
	primaryColor   = lipgloss.AdaptiveColor{Light: "#005F87", Dark: "#87CEEB"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#AF005F", Dark: "#FF5F87"}
	accentColor    = lipgloss.AdaptiveColor{Light: "#008700", Dark: "#A7E22E"}
	errorColor     = lipgloss.AdaptiveColor{Light: "#D70000", Dark: "#FF5F5F"}
	warningColor   = lipgloss.AdaptiveColor{Light: "#AF8700", Dark: "#FFAF00"}
	inactiveColor  = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#4A4A4A"}
)

// Common Styles
var (
	appStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(inactiveColor).
			MarginTop(1)

	titleStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Item Styles
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(primaryColor)
	largeFileStyle    = lipgloss.NewStyle().PaddingLeft(4).Foreground(warningColor)
	directoryStyle    = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	// Input Styles
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(inactiveColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Selection Indicators
	checkedIcon   = "☑"
	uncheckedIcon = "☐"
	folderIcon    = "📁"
	fileIcon      = "📄"
	arrowIcon     = "❯"
)
