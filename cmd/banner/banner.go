package banner

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Cyberpunk color palette
var (
	cyan      = lipgloss.Color("#00FFFF")
	magenta   = lipgloss.Color("#FF00FF")
	neonGreen = lipgloss.Color("#39FF14")
	neonPink  = lipgloss.Color("#FF6EC7")
	dimCyan   = lipgloss.Color("#008B8B")
	white     = lipgloss.Color("#FFFFFF")
	dim       = lipgloss.Color("#666666")
)

// PrintNeonBanner prints the JikiME ADK banner with neon gradient style
// JiKiME - lowercase 'i' shown as small vertical bars with dots
func PrintNeonBanner() {
	fmt.Println()

	// Row 1 - J i K i M E   A D K (i = small dot for lowercase effect)
	fmt.Print("\033[38;5;51m     ██╗\033[38;5;50m▪ \033[38;5;49m██╗  ██╗\033[38;5;48m▪ \033[38;5;47m███╗   ███╗\033[38;5;46m███████╗\033[0m")
	fmt.Println("\033[38;5;213m     █████╗ \033[38;5;212m██████╗ \033[38;5;211m██╗  ██╗\033[0m")

	// Row 2
	fmt.Print("\033[38;5;87m     ██║\033[38;5;86m█ \033[38;5;85m██║ ██╔╝\033[38;5;84m█ \033[38;5;83m████╗ ████║\033[38;5;82m██╔════╝\033[0m")
	fmt.Println("\033[38;5;177m    ██╔══██╗\033[38;5;176m██╔══██╗\033[38;5;175m██║ ██╔╝\033[0m")

	// Row 3
	fmt.Print("\033[38;5;123m     ██║\033[38;5;122m█ \033[38;5;121m█████╔╝ \033[38;5;120m█ \033[38;5;119m██╔████╔██║\033[38;5;118m█████╗  \033[0m")
	fmt.Println("\033[38;5;141m    ███████║\033[38;5;140m██║  ██║\033[38;5;139m█████╔╝ \033[0m")

	// Row 4
	fmt.Print("\033[38;5;159m██   ██║\033[38;5;158m█ \033[38;5;157m██╔═██╗ \033[38;5;156m█ \033[38;5;155m██║╚██╔╝██║\033[38;5;154m██╔══╝  \033[0m")
	fmt.Println("\033[38;5;105m    ██╔══██║\033[38;5;104m██║  ██║\033[38;5;103m██╔═██╗ \033[0m")

	// Row 5
	fmt.Print("\033[38;5;195m╚█████╔╝\033[38;5;194m█ \033[38;5;193m██║  ██╗\033[38;5;192m█ \033[38;5;191m██║ ╚═╝ ██║\033[38;5;190m███████╗\033[0m")
	fmt.Println("\033[38;5;69m    ██║  ██║\033[38;5;68m██████╔╝\033[38;5;67m██║  ██╗\033[0m")

	// Row 6
	fmt.Print("\033[38;5;231m ╚════╝ \033[38;5;230m▀ \033[38;5;229m╚═╝  ╚═╝\033[38;5;228m▀ \033[38;5;227m╚═╝     ╚═╝\033[38;5;226m╚══════╝\033[0m")
	fmt.Println("\033[38;5;33m    ╚═╝  ╚═╝\033[38;5;32m╚═════╝ \033[38;5;31m╚═╝  ╚═╝\033[0m")

	fmt.Println()
}

// PrintIntro prints the intro banner with version info in cyberpunk style
func PrintIntro(version string) {
	// Clear screen effect
	fmt.Print("\033[2J\033[H")

	// Print neon banner
	PrintNeonBanner()

	// Styled version
	versionStyle := lipgloss.NewStyle().
		Foreground(dimCyan).
		Italic(true)
	fmt.Println(versionStyle.Render(fmt.Sprintf("    v%s", version)))
	fmt.Println()

	// Cyberpunk separator
	separatorStyle := lipgloss.NewStyle().Foreground(cyan)
	fmt.Println(separatorStyle.Render("    ░▒▓████████████████████████████████████████████████████▓▒░"))
	fmt.Println()

	// Boot sequence style welcome
	bootPrefix := lipgloss.NewStyle().
		Foreground(neonGreen).
		Bold(true)

	systemMsg := lipgloss.NewStyle().
		Foreground(white)

	fmt.Println(bootPrefix.Render("    [SYS]") + systemMsg.Render(" Initializing JiKiME-ADK Setup Wizard..."))
	fmt.Println()

	// Welcome message
	welcomeStyle := lipgloss.NewStyle().
		Foreground(magenta).
		Bold(true)

	fmt.Println(welcomeStyle.Render("    🚀 WELCOME TO JiKiME-ADK"))
	fmt.Println()

	// Instructions
	dimStyle := lipgloss.NewStyle().Foreground(dim)
	fmt.Println(dimStyle.Render("    This wizard will configure your development environment."))
	fmt.Println(dimStyle.Render("    Press Ctrl+C at any time to abort."))
	fmt.Println()

	// Short delay for effect
	time.Sleep(300 * time.Millisecond)
}

// PrintCompact prints a compact version of the banner for smaller spaces
// JiKiME - lowercase 'i' shown as small characters
func PrintCompact() {
	fmt.Println("\033[38;5;51m ╦\033[38;5;87mi\033[38;5;123m╦╔═\033[38;5;159mi\033[38;5;195m╔╦╗\033[38;5;231m╔═╗\033[0m  \033[38;5;213m╔═╗\033[38;5;177m╔╦╗\033[38;5;141m╦╔═\033[0m")
	fmt.Println("\033[38;5;50m ║\033[38;5;86m│\033[38;5;122m╠╩╗\033[38;5;158m│\033[38;5;194m║║║\033[38;5;230m║╣ \033[0m  \033[38;5;212m╠═╣\033[38;5;176m ║║\033[38;5;140m╠╩╗\033[0m")
	fmt.Println("\033[38;5;49m╚╝\033[38;5;85m·\033[38;5;121m╩ ╩\033[38;5;157m·\033[38;5;193m╩ ╩\033[38;5;229m╚═╝\033[0m  \033[38;5;211m╩ ╩\033[38;5;175m═╩╝\033[38;5;139m╩ ╩\033[0m")
}

// PrintInitComplete prints completion message in cyberpunk style
func PrintInitComplete() {
	completeStyle := lipgloss.NewStyle().
		Foreground(neonGreen).
		Bold(true)

	fmt.Println()
	fmt.Println(completeStyle.Render("    ╔══════════════════════════════════════════════════════════╗"))
	fmt.Println(completeStyle.Render("    ║           INITIALIZATION SEQUENCE COMPLETE              ║"))
	fmt.Println(completeStyle.Render("    ╚══════════════════════════════════════════════════════════╝"))
	fmt.Println()
}

// PrintPhase prints a phase header
func PrintPhase(num int, label string) {
	phaseStyle := lipgloss.NewStyle().
		Foreground(magenta).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(white)

	statusStyle := lipgloss.NewStyle().
		Foreground(neonGreen)

	fmt.Printf("\n    %s %s %s\n",
		phaseStyle.Render(fmt.Sprintf("[PHASE %d]", num)),
		labelStyle.Render(label),
		statusStyle.Render("..."),
	)
}

// PrintPhaseOK prints phase completion status
func PrintPhaseOK() {
	okStyle := lipgloss.NewStyle().
		Foreground(neonGreen).
		Bold(true)

	fmt.Printf("              %s\n", okStyle.Render("[ OK ]"))
}
