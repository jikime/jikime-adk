// Package tag provides atomic file operations for TAG System v2.0.
package tag

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AtomicWriteText atomically writes text content to a file.
// Uses write-to-temp-then-rename pattern to prevent race conditions
// and partial writes from corrupting data.
func AtomicWriteText(filePath, content string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp_*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomic rename (overwrites target if exists)
	if err := os.Rename(tmpPath, filePath); err != nil {
		return err
	}

	// Clear tmpPath so deferred cleanup doesn't remove the target
	tmpPath = ""
	return nil
}

// AtomicWriteJSON atomically writes JSON data to a file.
// Uses write-to-temp-then-rename pattern to prevent race conditions
// and partial writes from corrupting data.
func AtomicWriteJSON(filePath string, data interface{}) error {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp_*.json")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	// Write JSON to temp file
	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomic rename (overwrites target if exists)
	if err := os.Rename(tmpPath, filePath); err != nil {
		return err
	}

	// Clear tmpPath so deferred cleanup doesn't remove the target
	tmpPath = ""
	return nil
}
