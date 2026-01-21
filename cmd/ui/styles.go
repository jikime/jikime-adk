package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Cyberpunk color palette
var (
	// Primary colors
	Cyan       = lipgloss.Color("#00FFFF")
	Magenta    = lipgloss.Color("#FF00FF")
	NeonGreen  = lipgloss.Color("#39FF14")
	NeonPink   = lipgloss.Color("#FF6EC7")
	NeonBlue   = lipgloss.Color("#00D4FF")
	NeonPurple = lipgloss.Color("#BF00FF")
	NeonOrange = lipgloss.Color("#FF9500")

	// UI colors
	Highlight = lipgloss.Color("#FFFF00")
	Dim       = lipgloss.Color("#666666")
	DimCyan   = lipgloss.Color("#008B8B")
	White     = lipgloss.Color("#FFFFFF")
	Black     = lipgloss.Color("#000000")
	DarkGray  = lipgloss.Color("#1a1a2e")
	MidGray   = lipgloss.Color("#333333")

	// Status colors
	Success = lipgloss.Color("#00FF00")
	Warning = lipgloss.Color("#FFD700")
	Error   = lipgloss.Color("#FF0000")
	Info    = lipgloss.Color("#00BFFF")
)

// Box styles
var (
	// Main container box with neon border
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Cyan).
			Padding(1, 2)

	// Section box
	SectionBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(DimCyan).
			Padding(0, 1).
			MarginTop(1)

	// Highlighted box
	HighlightBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(NeonGreen).
			Padding(1, 2)
)

// Text styles
var (
	// Title style - big and bold
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Cyan).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(NeonPink).
			Italic(true)

	// Section header
	SectionHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(NeonGreen).
			Background(DarkGray).
			Padding(0, 1)

	// Normal text
	NormalText = lipgloss.NewStyle().
			Foreground(White)

	// Dim text for descriptions
	DimText = lipgloss.NewStyle().
			Foreground(Dim)

	// Highlight text
	HighlightText = lipgloss.NewStyle().
			Bold(true).
			Foreground(Highlight)

	// Success text
	SuccessText = lipgloss.NewStyle().
			Foreground(Success)

	// Error text
	ErrorText = lipgloss.NewStyle().
			Foreground(Error)

	// Warning text
	WarningText = lipgloss.NewStyle().
			Foreground(Warning)

	// Info text
	InfoText = lipgloss.NewStyle().
			Foreground(Info)

	// Cyber prompt style
	PromptStyle = lipgloss.NewStyle().
			Foreground(Magenta).
			Bold(true)

	// Code/command style
	CodeStyle = lipgloss.NewStyle().
			Foreground(NeonGreen).
			Background(MidGray).
			Padding(0, 1)

	// Version style
	VersionStyle = lipgloss.NewStyle().
			Foreground(DimCyan).
			Italic(true)
)

// Icon styles with colors
var (
	IconCheck   = SuccessText.Render("✓")
	IconCross   = ErrorText.Render("✗")
	IconArrow   = lipgloss.NewStyle().Foreground(Cyan).Render("❯")
	IconDot     = lipgloss.NewStyle().Foreground(NeonPink).Render("●")
	IconStar    = lipgloss.NewStyle().Foreground(Highlight).Render("★")
	IconRocket  = lipgloss.NewStyle().Foreground(NeonOrange).Render("🚀")
	IconGlobe   = lipgloss.NewStyle().Foreground(NeonBlue).Render("🌐")
	IconUser    = lipgloss.NewStyle().Foreground(NeonPurple).Render("👤")
	IconGit     = lipgloss.NewStyle().Foreground(NeonOrange).Render("🔀")
	IconTag     = lipgloss.NewStyle().Foreground(NeonPink).Render("🏷️")
	IconFolder  = lipgloss.NewStyle().Foreground(Highlight).Render("📁")
	IconFile    = lipgloss.NewStyle().Foreground(White).Render("📄")
	IconClock   = lipgloss.NewStyle().Foreground(NeonBlue).Render("⏱️")
	IconChart   = lipgloss.NewStyle().Foreground(NeonGreen).Render("📊")
	IconWarning = WarningText.Render("⚠️")
)

// Progress bar characters
const (
	ProgressFull  = "█"
	ProgressEmpty = "░"
	ProgressHead  = "▓"
)

// Decorative elements
var (
	Separator = lipgloss.NewStyle().
			Foreground(DimCyan).
			Render("─────────────────────────────────────────────────────────────")

	DoubleSeparator = lipgloss.NewStyle().
			Foreground(Cyan).
			Render("═════════════════════════════════════════════════════════════")

	GlitchSeparator = lipgloss.NewStyle().
			Foreground(NeonGreen).
			Render("░▒▓█████████████████████████████████████████████████████▓▒░")
)

// ASCII decorations
const (
	CornerTL = "╔"
	CornerTR = "╗"
	CornerBL = "╚"
	CornerBR = "╝"
	LineH    = "═"
	LineV    = "║"
)
