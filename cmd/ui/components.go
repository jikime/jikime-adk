package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
		titleText := " " + title + " "
		styledTitle := titleStyle.Render(titleText)

		// Find position after "╭──" (margin + corner + 2 dashes)
		// We'll replace characters in the top border to embed the title
		runes := []rune(StripAnsi(boxLines[0]))
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
