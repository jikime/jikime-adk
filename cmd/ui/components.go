package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// PrintCyberHeader prints a cyberpunk-style section header
func PrintCyberHeader(icon, title string) {
	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		icon+" ",
		SectionHeader.Render(title),
	)
	fmt.Println()
	fmt.Println(header)
}

// PrintCyberBox prints content in a styled box
func PrintCyberBox(content string) {
	fmt.Println(BoxStyle.Render(content))
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Println(IconCheck + " " + SuccessText.Render(message))
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Println(IconCross + " " + ErrorText.Render(message))
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Println(IconWarning + " " + WarningText.Render(message))
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Println(IconDot + " " + InfoText.Render(message))
}

// PrintDim prints dimmed text
func PrintDim(message string) {
	fmt.Println(DimText.Render(message))
}

// PrintPrompt prints a prompt indicator
func PrintPrompt(message string) {
	fmt.Println(IconArrow + " " + PromptStyle.Render(message))
}

// PrintSystemMessage prints a "system" style message (hacker aesthetic)
func PrintSystemMessage(message string) {
	prefix := lipgloss.NewStyle().
		Foreground(NeonGreen).
		Bold(true).
		Render("[SYS]")
	fmt.Printf("%s %s\n", prefix, NormalText.Render(message))
}

// PrintBootSequence prints a boot sequence style message with animation
func PrintBootSequence(message string, delay time.Duration) {
	prefix := lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true).
		Render("▶")

	fmt.Printf("%s %s", prefix, DimText.Render(message))

	// Simple dot animation
	for i := 0; i < 3; i++ {
		time.Sleep(delay / 3)
		fmt.Print(lipgloss.NewStyle().Foreground(NeonGreen).Render("."))
	}

	fmt.Println(SuccessText.Render(" OK"))
}

// PrintProgressBar prints a styled progress bar
func PrintProgressBar(current, total int, width int, label string) string {
	if width <= 0 {
		width = 40
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat(ProgressFull, filled)
	if filled < width {
		bar += ProgressHead
		bar += strings.Repeat(ProgressEmpty, width-filled-1)
	}

	barStyle := lipgloss.NewStyle().Foreground(Cyan)
	percentStyle := lipgloss.NewStyle().Foreground(NeonGreen).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(White)

	return fmt.Sprintf("%s %s %s",
		labelStyle.Render(label),
		barStyle.Render(bar),
		percentStyle.Render(fmt.Sprintf("%3d%%", int(percent*100))),
	)
}

// AnimatedProgress shows an animated progress bar
func AnimatedProgress(label string, duration time.Duration) {
	width := 40
	steps := 50
	stepDuration := duration / time.Duration(steps)

	for i := 0; i <= steps; i++ {
		fmt.Print("\r" + PrintProgressBar(i, steps, width, "  "+label))
		time.Sleep(stepDuration)
	}
	fmt.Println()
}

// PrintPhaseStart prints the start of a phase with animation
func PrintPhaseStart(phase int, label string) {
	phaseStyle := lipgloss.NewStyle().
		Foreground(Magenta).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(White)

	prefix := phaseStyle.Render(fmt.Sprintf("[PHASE %d]", phase))
	fmt.Printf("\n%s %s\n", prefix, labelStyle.Render(label))
}

// PrintPhaseComplete prints phase completion
func PrintPhaseComplete(phase int) {
	completeStyle := lipgloss.NewStyle().
		Foreground(NeonGreen).
		Bold(true)
	fmt.Printf("         %s\n", completeStyle.Render("COMPLETE"))
}

// PrintSummaryRow prints a summary row with label and value
func PrintSummaryRow(icon, label, value string) {
	labelStyle := lipgloss.NewStyle().
		Foreground(DimCyan).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(White)

	fmt.Printf("  %s %s %s\n",
		icon,
		labelStyle.Render(label+":"),
		valueStyle.Render(value),
	)
}

// PrintGlitchText prints text with a glitch effect
func PrintGlitchText(text string) {
	colors := []lipgloss.Color{Cyan, Magenta, NeonGreen}
	result := ""

	for i, char := range text {
		color := colors[i%len(colors)]
		result += lipgloss.NewStyle().Foreground(color).Render(string(char))
	}

	fmt.Println(result)
}

// PrintWelcomeBox prints a welcome box with cyberpunk style
func PrintWelcomeBox(title, subtitle, version string) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Cyan)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(NeonPink)

	versionStyle := lipgloss.NewStyle().
		Foreground(Dim).
		Italic(true)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		subtitleStyle.Render(subtitle),
		"",
		versionStyle.Render("v"+version),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Cyan).
		Padding(1, 4).
		Render(content)

	fmt.Println(box)
}

// PrintNextSteps prints styled next steps
func PrintNextSteps(steps []string) {
	headerStyle := lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	numberStyle := lipgloss.NewStyle().
		Foreground(Magenta).
		Bold(true)

	stepStyle := lipgloss.NewStyle().
		Foreground(White)

	fmt.Println()
	fmt.Println(headerStyle.Render("📋 NEXT STEPS"))
	fmt.Println()

	for i, step := range steps {
		fmt.Printf("  %s %s\n",
			numberStyle.Render(fmt.Sprintf("[%d]", i+1)),
			stepStyle.Render(step),
		)
	}
}

// CyberSpinner returns spinner frames
func CyberSpinner() []string {
	return []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
	}
}

// PrintSpinner prints a spinner with message
func PrintSpinner(message string, done chan bool) {
	frames := CyberSpinner()
	spinnerStyle := lipgloss.NewStyle().Foreground(Cyan)
	msgStyle := lipgloss.NewStyle().Foreground(White)

	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r" + strings.Repeat(" ", len(message)+10) + "\r")
			return
		default:
			fmt.Printf("\r%s %s",
				spinnerStyle.Render(frames[i%len(frames)]),
				msgStyle.Render(message),
			)
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}
}
