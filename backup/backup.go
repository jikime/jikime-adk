// Package backup provides template backup management for jikime-adk.
// Implements backup creation, listing, cleanup, and restoration with metadata tracking.
package backup

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

// BackupExcludeDirs defines paths excluded from backups (protect user data).
// These are relative paths within .jikime/ directory.
var BackupExcludeDirs = []string{
	"specs",   // User SPEC documents
	"reports", // User reports
	"project", // User project documents (product/structure/tech.md)
	"config",  // User configuration files (YAML)
}

// TrackedItems defines the items to be backed up.
var TrackedItems = []string{
	".jikime",
	".claude",
	".github",
	"CLAUDE.md",
	".mcp.json",
	".lsp.json",
	".git-hooks",
}

// BackupMetadata contains information about a backup.
type BackupMetadata struct {
	Timestamp     string   `json:"timestamp"`
	Description   string   `json:"description"`
	BackedUpItems []string `json:"backed_up_items"`
	ExcludedItems []string `json:"excluded_items"`
	ExcludedDirs  []string `json:"excluded_dirs"`
	ProjectRoot   string   `json:"project_root"`
	BackupType    string   `json:"backup_type"`
}

// TemplateBackup manages template backups for a project.
type TemplateBackup struct {
	TargetPath string // Project path (absolute)
}

// NewTemplateBackup creates a new TemplateBackup instance.
func NewTemplateBackup(targetPath string) (*TemplateBackup, error) {
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}
	return &TemplateBackup{TargetPath: absPath}, nil
}

// BackupDir returns the backup directory path.
func (tb *TemplateBackup) BackupDir() string {
	return filepath.Join(tb.TargetPath, ".jikime-backups")
}

// HasExistingFiles checks whether backup-worthy files already exist.
func (tb *TemplateBackup) HasExistingFiles() bool {
	for _, item := range TrackedItems {
		path := filepath.Join(tb.TargetPath, item)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// CreateBackup creates a timestamped backup under .jikime-backups/.
// Returns the path to the created backup directory.
func (tb *TemplateBackup) CreateBackup() (string, error) {
	// Generate timestamp for backup directory name
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(tb.BackupDir(), timestamp)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", err
	}

	// Track backed up items for metadata
	backedUpItems := []string{}
	excludedItems := []string{}

	// Copy backup targets
	for _, item := range TrackedItems {
		src := filepath.Join(tb.TargetPath, item)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		dst := filepath.Join(backupPath, item)

		if item == ".jikime" {
			// Copy while skipping protected paths
			excluded, err := tb.copyExcludeProtected(src, dst)
			if err != nil {
				return "", err
			}
			backedUpItems = append(backedUpItems, item)
			excludedItems = append(excludedItems, excluded...)
		} else {
			info, err := os.Stat(src)
			if err != nil {
				continue
			}

			if info.IsDir() {
				if err := copyDir(src, dst); err != nil {
					return "", err
				}
			} else {
				if err := copyFile(src, dst); err != nil {
					return "", err
				}
			}
			backedUpItems = append(backedUpItems, item)
		}
	}

	// Create backup metadata
	metadata := BackupMetadata{
		Timestamp:     timestamp,
		Description:   "template_backup",
		BackedUpItems: backedUpItems,
		ExcludedItems: excludedItems,
		ExcludedDirs:  BackupExcludeDirs,
		ProjectRoot:   tb.TargetPath,
		BackupType:    "template",
	}

	metadataPath := filepath.Join(backupPath, "backup_metadata.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

// GetLatestBackup returns the most recent backup path.
// Supports both new timestamped and legacy backup structures.
func (tb *TemplateBackup) GetLatestBackup() string {
	backupDir := tb.BackupDir()
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return ""
	}

	// Match pattern: YYYYMMDD_HHMMSS
	timestampPattern := regexp.MustCompile(`^\d{8}_\d{6}$`)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return ""
	}

	var timestampedBackups []string
	for _, entry := range entries {
		if entry.IsDir() && timestampPattern.MatchString(entry.Name()) {
			timestampedBackups = append(timestampedBackups, entry.Name())
		}
	}

	if len(timestampedBackups) > 0 {
		// Sort and return the latest
		sort.Strings(timestampedBackups)
		return filepath.Join(backupDir, timestampedBackups[len(timestampedBackups)-1])
	}

	// Fall back to legacy backup/ directory
	legacyBackup := filepath.Join(backupDir, "backup")
	if _, err := os.Stat(legacyBackup); err == nil {
		return legacyBackup
	}

	return ""
}

// ListBackups returns all timestamped backup directories sorted by timestamp (newest first).
func (tb *TemplateBackup) ListBackups() []string {
	backupDir := tb.BackupDir()
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return nil
	}

	// Match pattern: YYYYMMDD_HHMMSS
	timestampPattern := regexp.MustCompile(`^\d{8}_\d{6}$`)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && timestampPattern.MatchString(entry.Name()) {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	// Sort in descending order (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))
	return backups
}

// CleanupOldBackups removes old backups, keeping only the most recent ones.
// Returns the number of backups deleted.
func (tb *TemplateBackup) CleanupOldBackups(keepCount int) (int, error) {
	if keepCount <= 0 {
		keepCount = 5
	}

	backups := tb.ListBackups()
	if len(backups) <= keepCount {
		return 0, nil
	}

	deletedCount := 0
	for _, backupPath := range backups[keepCount:] {
		if err := os.RemoveAll(backupPath); err == nil {
			deletedCount++
		}
		// Ignore deletion errors and continue with other backups
	}

	return deletedCount, nil
}

// RestoreBackup restores project files from a backup.
// If backupPath is empty, it restores from the latest backup.
func (tb *TemplateBackup) RestoreBackup(backupPath string) error {
	if backupPath == "" {
		backupPath = tb.GetLatestBackup()
	}

	if backupPath == "" {
		return os.ErrNotExist
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return err
	}

	// Restore each item from backup
	for _, item := range TrackedItems {
		src := filepath.Join(backupPath, item)
		dst := filepath.Join(tb.TargetPath, item)

		// Skip if not in backup
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		// Remove current version
		if _, err := os.Stat(dst); err == nil {
			if err := os.RemoveAll(dst); err != nil {
				return err
			}
		}

		// Restore from backup
		info, err := os.Stat(src)
		if err != nil {
			continue
		}

		if info.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return err
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReadMetadata reads the backup metadata from a backup directory.
func (tb *TemplateBackup) ReadMetadata(backupPath string) (*BackupMetadata, error) {
	metadataPath := filepath.Join(backupPath, "backup_metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// copyExcludeProtected copies backup content while excluding protected paths.
func (tb *TemplateBackup) copyExcludeProtected(src, dst string) ([]string, error) {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return nil, err
	}

	var excluded []string

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip excluded paths
		for _, excludeDir := range BackupExcludeDirs {
			if relPath == excludeDir || hasPathPrefix(relPath, excludeDir) {
				excluded = append(excluded, relPath)
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		dstItem := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstItem, info.Mode())
		}

		return copyFile(path, dstItem)
	})

	return excluded, err
}

// hasPathPrefix checks if path starts with prefix as a directory component.
func hasPathPrefix(path, prefix string) bool {
	if path == prefix {
		return true
	}
	return len(path) > len(prefix) && path[:len(prefix)] == prefix && path[len(prefix)] == filepath.Separator
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
