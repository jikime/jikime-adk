package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// typeHeadingMap maps DailyLogEntry.Type values to markdown section headings.
// Only structured memory types use section-based grouping.
// Session data types (user_prompt, assistant_response, tool_usage) are appended
// chronologically without section grouping.
var typeHeadingMap = map[string]string{
	"decision":        "Decision",
	"learning":        "Learning",
	"error_fix":       "Error Fix",
	"session_summary": "Session Summary",
}

// chronologicalTypes are appended in time order without section grouping.
var chronologicalTypes = map[string]bool{
	"user_prompt":        true,
	"assistant_response": true,
	"tool_usage":         true,
}

// AppendDailyLog appends an entry to the daily log file.
// File path: <projectDir>/.jikime/memory/<YYYY-MM-DD>.md
// Creates the file with a date header if it doesn't exist.
// Returns the relative path from projectDir (e.g. ".jikime/memory/2026-01-27.md").
func AppendDailyLog(projectDir string, entry DailyLogEntry) (string, error) {
	today := time.Now().Format("2006-01-02")
	relPath := filepath.Join(".jikime", "memory", today+".md")
	absPath := filepath.Join(projectDir, relPath)

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create memory directory: %w", err)
	}

	content, err := readFileOrEmpty(absPath)
	if err != nil {
		return "", fmt.Errorf("read daily log: %w", err)
	}

	// If file is empty/new, add the date header.
	if content == "" {
		content = "# " + today + "\n"
	}

	if chronologicalTypes[entry.Type] {
		// Session data: append chronologically at file end with timestamp and type label.
		ts := time.Now().Format("15:04:05")
		label := titleCase(strings.ReplaceAll(entry.Type, "_", " "))
		item := fmt.Sprintf("- [%s] **%s**: %s", ts, label, entry.Content)
		if entry.Metadata != "" {
			item += " (" + entry.Metadata + ")"
		}
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += item + "\n"
	} else {
		// Structured memory: group under ## section heading.
		heading, ok := typeHeadingMap[entry.Type]
		if !ok {
			heading = strings.ReplaceAll(entry.Type, "_", " ")
			heading = titleCase(heading)
		}
		item := "- " + entry.Content
		if entry.Metadata != "" {
			item += " (" + entry.Metadata + ")"
		}
		content = appendUnderSection(content, heading, item)
	}

	if err := writeFileAtomic(absPath, content); err != nil {
		return "", fmt.Errorf("write daily log: %w", err)
	}

	return relPath, nil
}

// AppendMemoryMD appends an entry to the MEMORY.md file.
// File path: <projectDir>/.jikime/memory/MEMORY.md
// Creates the file with "# Project Memory" header if it doesn't exist.
func AppendMemoryMD(projectDir string, entry MemoryMDEntry) error {
	absPath := filepath.Join(projectDir, ".jikime", "memory", "MEMORY.md")

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create memory directory: %w", err)
	}

	content, err := readFileOrEmpty(absPath)
	if err != nil {
		return fmt.Errorf("read MEMORY.md: %w", err)
	}

	if content == "" {
		content = "# Project Memory\n"
	}

	item := "- " + entry.Content
	content = appendUnderSection(content, entry.Section, item)

	if err := writeFileAtomic(absPath, content); err != nil {
		return fmt.Errorf("write MEMORY.md: %w", err)
	}

	return nil
}

// DailyLogPath returns the filename for today's daily log.
func DailyLogPath() string {
	return time.Now().Format("2006-01-02") + ".md"
}

// appendUnderSection finds the "## <heading>" section in content and appends item
// under it. If the section doesn't exist, it is created at the end.
func appendUnderSection(content, heading, item string) string {
	sectionHeader := "## " + heading
	lines := strings.Split(content, "\n")

	// Find the section.
	sectionIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			sectionIdx = i
			break
		}
	}

	if sectionIdx == -1 {
		// Section doesn't exist — append at the end.
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + sectionHeader + "\n\n" + item + "\n"
		return content
	}

	// Section exists. Find the insertion point: just before the next ## heading
	// or at the end of file, after existing list items.
	insertIdx := len(lines)
	for i := sectionIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "## ") {
			insertIdx = i
			break
		}
	}

	// Walk backwards from insertIdx to skip trailing blank lines within the section,
	// so we insert right after the last content line in the section.
	insertAt := insertIdx
	for insertAt > sectionIdx+1 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}

	// Insert the item.
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, item)
	newLines = append(newLines, lines[insertAt:]...)

	return strings.Join(newLines, "\n")
}

// readFileOrEmpty reads a file and returns its content as a string.
// Returns an empty string if the file doesn't exist.
func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// writeFileAtomic writes content to a file using a temp file + rename pattern
// for crash safety.
func writeFileAtomic(path, content string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Clean up on failure.
	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	tmpPath = "" // prevent deferred cleanup
	return nil
}

// titleCase converts "error fix" to "Error Fix".
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
