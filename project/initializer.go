package project

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"jikime-adk/templates"
)

// ProjectInitializer orchestrates project scaffolding.
type ProjectInitializer struct {
	path string
}

// InitializeResult exposes the outcome of initialization.
type InitializeResult struct {
	ProjectPath   string
	Locale        string
	Language      string
	CreatedFiles  []string
	Duration      time.Duration
	Reinitialized bool
}

// NewInitializer returns a new ProjectInitializer for the given path.
func NewInitializer(path string) *ProjectInitializer {
	return &ProjectInitializer{path: path}
}

// Initialize scaffolds the project directories and configuration.
func (pi *ProjectInitializer) Initialize(answers SetupAnswers, force bool) (InitializeResult, error) {
	start := time.Now()
	created := []string{}
	caretDir := filepath.Join(pi.path, ".jikime")
	sectionsDir := filepath.Join(caretDir, "config", "sections")

	reinitialized := false
	if exists(caretDir) {
		if !force {
			return InitializeResult{}, fmt.Errorf("project already initialized at %s", pi.path)
		}
		reinitialized = true
	}

	// Phase 1: Create directory structure
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		return InitializeResult{}, err
	}

	// Create additional required directories
	claudeDir := filepath.Join(pi.path, ".claude")
	requiredDirs := []string{
		filepath.Join(caretDir, "project"),
		filepath.Join(caretDir, "specs"),
		filepath.Join(caretDir, "reports"),
		filepath.Join(caretDir, "memory"),
		filepath.Join(claudeDir, "logs"),
	}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return InitializeResult{}, err
		}
	}

	// Phase 2: Copy templates from jikime-adk/templates/ to project directory
	templateRoot, err := getTemplateRoot()
	if err != nil {
		return InitializeResult{}, fmt.Errorf("failed to find template directory: %w", err)
	}

	// Prepare variable substitution context
	conversationLanguageName := languageCodeToName(answers.Locale)
	context := map[string]string{
		"PROJECT_NAME":               answers.ProjectName,
		"CONVERSATION_LANGUAGE":      answers.Locale,
		"CONVERSATION_LANGUAGE_NAME": conversationLanguageName,
		"GIT_COMMIT_LANG":            answers.GitCommitLang,
		"CODE_COMMENT_LANG":          answers.CodeCommentLang,
		"DOCUMENTATION_LANG":         answers.DocLang,
		"GIT_MODE":                   answers.GitMode,
		"GITHUB_USER":                answers.GitHubUser,
	}

	// Copy template directories and files
	templateDirs := []string{".claude", ".jikime", ".github"}
	for _, dir := range templateDirs {
		src := filepath.Join(templateRoot, dir)
		dst := filepath.Join(pi.path, dir)
		if exists(src) {
			copiedFiles, err := copyDirectoryWithSubstitution(src, dst, context)
			if err != nil {
				return InitializeResult{}, fmt.Errorf("failed to copy %s: %w", dir, err)
			}
			created = append(created, copiedFiles...)
		}
	}

	// Copy template files
	templateFiles := []string{"CLAUDE.md", ".gitignore", ".mcp.json"}
	for _, file := range templateFiles {
		src := filepath.Join(templateRoot, file)
		dst := filepath.Join(pi.path, file)
		if exists(src) {
			if err := copyFileWithSubstitution(src, dst, context); err != nil {
				return InitializeResult{}, fmt.Errorf("failed to copy %s: %w", file, err)
			}
			created = append(created, dst)
		}
	}

	// Phase 3: Update configuration section files
	entries := []struct {
		filename string
		content  map[string]interface{}
	}{
		{"language.yaml", map[string]interface{}{"language": map[string]string{"conversation_language": answers.Locale, "conversation_language_name": conversationLanguageName, "git_commit_messages": answers.GitCommitLang, "code_comments": answers.CodeCommentLang, "documentation": answers.DocLang}}},
		{"git-strategy.yaml", map[string]interface{}{"git_strategy": map[string]string{"mode": answers.GitMode}}},
		{"project.yaml", map[string]interface{}{"project": map[string]string{"name": answers.ProjectName}, "github": map[string]string{"profile_name": answers.GitHubUser}}},
		{"user.yaml", map[string]interface{}{"user": map[string]string{"name": answers.UserName}}},
	}

	for _, entry := range entries {
		if entry.filename == "project.yaml" && answers.GitHubUser == "" {
			delete(entry.content, "github")
		}
		filePath := filepath.Join(sectionsDir, entry.filename)

		// Skip if file already exists (preserve user customizations)
		if exists(filePath) {
			continue
		}

		if err := writeYAML(filePath, entry.content); err != nil {
			return InitializeResult{}, err
		}
		created = append(created, filePath)
	}

	// Phase 4: Initialize git repository (if not already initialized)
	initializeGit(pi.path)

	// Phase 5: Update settings.json with companyAnnouncements
	if err := updateSettingsWithAnnouncements(pi.path, templateRoot, answers.Locale); err != nil {
		// Non-fatal error - continue even if announcements update fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to update announcements: %v\n", err)
	}

	return InitializeResult{
		ProjectPath:   pi.path,
		Locale:        answers.Locale,
		Language:      answers.ProjectName,
		CreatedFiles:  created,
		Duration:      time.Since(start),
		Reinitialized: reinitialized,
	}, nil
}

func writeYAML(path string, content map[string]interface{}) error {
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// languageCodeToName converts a language code to its display name
func languageCodeToName(code string) string {
	names := map[string]string{
		"en": "English",
		"ko": "Korean (한국어)",
		"ja": "Japanese (日本語)",
		"zh": "Chinese (中文)",
		"es": "Spanish (Español)",
		"fr": "French (Français)",
		"de": "German (Deutsch)",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "English"
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getTemplateRoot finds the templates directory path
func getTemplateRoot() (string, error) {
	// 1. Check environment variable (highest priority)
	if envPath := os.Getenv("JIKIME_TEMPLATE_DIR"); envPath != "" {
		if exists(envPath) {
			return envPath, nil
		}
	}

	// 2. Try to find templates directory relative to executable
	execPath, err := os.Executable()
	if err == nil {
		// Resolve symlinks (for development)
		execPath, err = filepath.EvalSymlinks(execPath)
		if err != nil {
			execPath, _ = os.Executable()
		}

		// Check in same directory as executable
		execDir := filepath.Dir(execPath)
		templatesDir := filepath.Join(execDir, "templates")
		if exists(templatesDir) {
			return templatesDir, nil
		}

		// Check in parent directory (for development)
		parentDir := filepath.Dir(execDir)
		templatesDir = filepath.Join(parentDir, "templates")
		if exists(templatesDir) {
			return templatesDir, nil
		}

		// 3. Check in source directory structure (for go run)
		// Look for jikime-adk/templates from current path
		current := execDir
		for i := 0; i < 5; i++ {
			templatesDir = filepath.Join(current, "templates")
			if exists(templatesDir) {
				return templatesDir, nil
			}
			current = filepath.Dir(current)
		}
	}

	// 4. Check in current working directory (for development)
	cwd, err := os.Getwd()
	if err == nil {
		templatesDir := filepath.Join(cwd, "templates")
		if exists(templatesDir) {
			return templatesDir, nil
		}
	}

	// 5. Use embedded templates (for go install)
	if templates.HasEmbeddedTemplates() {
		return templates.ExtractTemplates()
	}

	return "", fmt.Errorf("templates directory not found")
}

// copyDirectoryWithSubstitution copies a directory recursively with variable substitution
// Returns a list of all files (not directories) that were copied
// Handles OS-specific files:
// - settings.json.unix: copied as settings.json on macOS/Linux
// - settings.json.windows: copied as settings.json on Windows
func copyDirectoryWithSubstitution(src, dst string, context map[string]string) ([]string, error) {
	copiedFiles := []string{}
	isWindows := runtime.GOOS == "windows"

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// Create directories
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		fileName := info.Name()

		// Handle OS-specific settings.json files
		if strings.HasSuffix(fileName, ".unix") {
			if isWindows {
				// Skip .unix files on Windows
				return nil
			}
			// On Unix (macOS/Linux), rename to remove .unix suffix
			dstPath = strings.TrimSuffix(dstPath, ".unix")
		} else if strings.HasSuffix(fileName, ".windows") {
			if !isWindows {
				// Skip .windows files on Unix
				return nil
			}
			// On Windows, rename to remove .windows suffix
			dstPath = strings.TrimSuffix(dstPath, ".windows")
		}

		// Copy files with substitution
		if err := copyFileWithSubstitution(path, dstPath, context); err != nil {
			return err
		}
		copiedFiles = append(copiedFiles, dstPath)
		return nil
	})
	return copiedFiles, err
}

// copyFileWithSubstitution copies a file with variable substitution
func copyFileWithSubstitution(src, dst string, context map[string]string) error {
	// Read source file
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Perform variable substitution
	text := string(content)
	for key, value := range context {
		placeholder := "{{" + key + "}}"
		text = strings.ReplaceAll(text, placeholder, value)
	}

	// Create parent directory if not exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// Determine file permissions
	// Default to writable (0644), make executable if source was executable
	perm := os.FileMode(0644)
	if info, err := os.Stat(src); err == nil {
		if info.Mode()&0111 != 0 { // Has any execute bit
			perm = 0755
		}
	}

	// Write to destination with writable permissions
	return os.WriteFile(dst, []byte(text), perm)
}

// copyFile copies a file without modification
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create parent directory if not exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// initializeGit initializes a git repository if not already initialized
func initializeGit(projectPath string) {
	// Check if .git directory already exists
	gitDir := filepath.Join(projectPath, ".git")
	if exists(gitDir) {
		// Git already initialized, skip
		return
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = projectPath

	// Run git init (errors are non-fatal)
	// Git might not be installed or available, but that's okay
	_ = cmd.Run()
}

// updateSettingsWithAnnouncements updates .claude/settings.json with companyAnnouncements
func updateSettingsWithAnnouncements(projectPath, templateRoot, locale string) error {
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")

	// Check if settings.json exists
	if !exists(settingsPath) {
		return fmt.Errorf("settings.json not found")
	}

	// Read current settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings.json: %w", err)
	}

	// Parse JSON
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings.json: %w", err)
	}

	// Load announcements for the specified language
	announcements, err := LoadAnnouncements(templateRoot, locale)
	if err != nil {
		// Use default announcements on error
		announcements = getDefaultAnnouncements()
	}

	// Update companyAnnouncements
	settings["companyAnnouncements"] = announcements

	// Marshal back to JSON with indentation
	updatedData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings.json: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(settingsPath, updatedData, 0o644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	return nil
}
