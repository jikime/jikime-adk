package hookscmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// autoMemory holds discovered project memory information.
type autoMemory struct {
	MemoryDir   string            // ~/.claude/projects/{hash}/memory/
	Files       map[string]string // filename -> content
	Available   bool              // memory directory exists and is readable
	TotalSize   int               // total bytes across all memory files
}

// discoverAutoMemory finds Claude Code's native auto-memory directory for the
// current project and reads all markdown files inside it.
//
// Claude Code stores project memories at:
//
//	~/.claude/projects/{path-hash}/memory/
//
// where {path-hash} is the project path with '/' replaced by '-'.
// Example: /Users/foo/myproject → -Users-foo-myproject
func discoverAutoMemory(cwd string) *autoMemory {
	mem := &autoMemory{
		Files: make(map[string]string),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return mem
	}

	// Compute project hash: replace all '/' with '-'
	// /Users/foo/project → -Users-foo-project
	projectHash := strings.ReplaceAll(cwd, "/", "-")

	memoryDir := filepath.Join(homeDir, ".claude", "projects", projectHash, "memory")

	info, err := os.Stat(memoryDir)
	if err != nil || !info.IsDir() {
		// Memory directory doesn't exist yet — this is normal for new sessions
		mem.MemoryDir = memoryDir
		return mem
	}

	mem.MemoryDir = memoryDir
	mem.Available = true

	// Read all .md files in the memory directory
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		return mem
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		filePath := filepath.Join(memoryDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		mem.Files[entry.Name()] = string(content)
		mem.TotalSize += len(content)
	}

	return mem
}

// formatMemorySection builds the SystemMessage section for auto-memory.
// Returns an empty string if no memory is available.
func formatMemorySection(mem *autoMemory) string {
	if !mem.Available || len(mem.Files) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n---\n")
	sb.WriteString("📚 **Auto-Memory Loaded**\n")
	sb.WriteString(fmt.Sprintf("   📁 Path: %s\n", mem.MemoryDir))
	sb.WriteString(fmt.Sprintf("   📄 Files: %d (%d bytes)\n", len(mem.Files), mem.TotalSize))

	// Priority file ordering: MEMORY.md first, then others alphabetically
	priority := []string{"MEMORY.md", "lessons.md", "context.md"}
	printed := map[string]bool{}

	for _, name := range priority {
		content, ok := mem.Files[name]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n### %s\n", name))
		sb.WriteString(truncateMemory(content, 800))
		printed[name] = true
	}

	// Remaining files
	for name, content := range mem.Files {
		if printed[name] {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n### %s\n", name))
		sb.WriteString(truncateMemory(content, 400))
	}

	sb.WriteString("\n---\n")

	return sb.String()
}

// truncateMemory limits memory content to maxLen bytes,
// appending "... (truncated)" if cut.
func truncateMemory(content string, maxLen int) string {
	content = strings.TrimSpace(content)
	if len(content) <= maxLen {
		return content + "\n"
	}
	return content[:maxLen] + "\n... (truncated)\n"
}

// ensureMemoryDir creates the project memory directory if it doesn't exist.
// This enables Claude Code's auto-memory to start persisting data.
func ensureMemoryDir(cwd string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectHash := strings.ReplaceAll(cwd, "/", "-")
	memoryDir := filepath.Join(homeDir, ".claude", "projects", projectHash, "memory")

	if err := os.MkdirAll(memoryDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create memory directory: %w", err)
	}

	return memoryDir, nil
}
