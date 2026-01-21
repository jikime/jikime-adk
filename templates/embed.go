// Package templates provides embedded template files for jikime-adk.
// This allows templates to be distributed with the binary via go install.
package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:*
var EmbeddedFS embed.FS

// ExtractTemplates extracts embedded templates to a temporary directory
// and returns the path to that directory.
func ExtractTemplates() (string, error) {
	// Create temp directory for templates
	tempDir, err := os.MkdirTemp("", "jikime-templates-*")
	if err != nil {
		return "", err
	}

	// Walk through embedded files and extract them
	err = fs.WalkDir(EmbeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the embed.go file itself
		if path == "embed.go" || strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip root
		if path == "." {
			return nil
		}

		destPath := filepath.Join(tempDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read embedded file
		content, err := EmbeddedFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Embedded files have read-only permissions (0444)
		// Use writable permissions for extracted files
		perm := os.FileMode(0644)

		// Check if file should be executable (scripts, etc.)
		if strings.HasSuffix(path, ".sh") || strings.Contains(path, "/bin/") {
			perm = 0755
		}

		// Write file with writable permissions
		return os.WriteFile(destPath, content, perm)
	})

	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// HasEmbeddedTemplates checks if embedded templates are available
func HasEmbeddedTemplates() bool {
	entries, err := EmbeddedFS.ReadDir(".")
	if err != nil {
		return false
	}
	// Check for at least .claude or .jikime directory
	for _, entry := range entries {
		if entry.Name() == ".claude" || entry.Name() == ".jikime" {
			return true
		}
	}
	return false
}
