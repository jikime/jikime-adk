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

// StripAnsi removes ANSI escape codes from string for accurate length calculation
func StripAnsi(s string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// visibleWidth returns the terminal display width of a string,
// correctly handling emojis (2 cells) and CJK characters using lipgloss.
func visibleWidth(s string) int {
	return lipgloss.Width(s)
}

// SectionLineWidth is the fixed total width for all bordered section headers.
const SectionLineWidth = 70

// PrintBorderedSection prints a section header with diamond and horizontal line.
// All sections render to the same total width (SectionLineWidth).
func PrintBorderedSection(title string) {
	diamondStyle := lipgloss.NewStyle().Foreground(DimCyan)
	titleStyle := lipgloss.NewStyle().Foreground(NeonPink).Bold(true)
	lineStyle := lipgloss.NewStyle().Foreground(DimCyan)

	prefix := "  ◇ "
	titleText := strings.ToUpper(title) + " "

	// Use lipgloss.Width for accurate terminal cell width (handles emoji/CJK)
	usedWidth := lipgloss.Width(prefix) + lipgloss.Width(titleText)
	remaining := SectionLineWidth - usedWidth
	if remaining < 2 {
		remaining = 2
	}

	fmt.Println()
	fmt.Println(
		diamondStyle.Render(prefix) +
			titleStyle.Render(titleText) +
			lineStyle.Render(strings.Repeat("─", remaining)),
	)
}

// PrintBorderedLine prints content with a left vertical border
// Output: │  content here
func PrintBorderedLine(content string) {
	borderStyle := lipgloss.NewStyle().Foreground(DimCyan)
	fmt.Println(borderStyle.Render("  │ ") + content)
}

// PrintBorderedResult prints a success result inside the bordered section
// Output: │ ✓ Label: Value
func PrintBorderedResult(label, value string) {
	borderStyle := lipgloss.NewStyle().Foreground(DimCyan)
	successStyle := lipgloss.NewStyle().Foreground(Success)
	dimStyle := lipgloss.NewStyle().Foreground(Dim)
	valueStyle := lipgloss.NewStyle().Foreground(White)
	fmt.Println(borderStyle.Render("  │ ") + successStyle.Render("✓ ") + dimStyle.Render(label+": ") + valueStyle.Render(value))
}

// PrintBorderedEmpty prints an empty bordered line
// Output: │
func PrintBorderedEmpty() {
	borderStyle := lipgloss.NewStyle().Foreground(DimCyan)
	fmt.Println(borderStyle.Render("  │"))
}

// PrintBorderedBox renders content inside a lipgloss bordered box with an embedded title.
// lipgloss handles terminal-aware width calculation, so emojis and CJK characters align correctly.
func PrintBorderedBox(title string, lines []string, width int) {
	if width <= 0 {
		width = 56
	}

	titleStyle := lipgloss.NewStyle().Foreground(NeonPink).Bold(true)

	// Build content: join lines with newline
	content := strings.Join(lines, "\n")

	// Use lipgloss border with proper width - it handles emoji/CJK width internally
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DimCyan).
		Padding(1, 2).
		Width(width).
		MarginLeft(2)

	// Render the box
	box := boxStyle.Render(content)

	// Replace top-left corner area to embed the title
	// Find the first line (top border) and inject the title
	boxLines := strings.Split(box, "\n")
	if len(boxLines) > 0 {
		topBorder := boxLines[0]
		titleText := " " + title + " "
		styledTitle := titleStyle.Render(titleText)

		// Find position after "╭──" (margin + corner + 2 dashes)
		// We'll replace characters in the top border to embed the title
		runes := []rune(StripAnsi(topBorder))
		if len(runes) > 6 {
			// Rebuild: keep margin+corner+2 dashes, insert title, then rest of dashes+corner
			borderColor := lipgloss.NewStyle().Foreground(DimCyan)
			prefix := strings.Repeat(" ", 2) + borderColor.Render("╭──")
			titleVisLen := len([]rune(titleText))
			// Calculate remaining dashes after title
			// Total inner width = width + padding*2 (lipgloss adds padding inside)
			totalInner := len(runes) - 2 - 2 // subtract margin(2) and corners(2)
			remaining := totalInner - 2 - titleVisLen // subtract prefix dashes(2) and title
			if remaining < 0 {
				remaining = 0
			}
			boxLines[0] = prefix + styledTitle + borderColor.Render(strings.Repeat("─", remaining)+"╮")
		}
	}

	fmt.Println()
	fmt.Println(strings.Join(boxLines, "\n"))
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
