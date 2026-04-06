package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"jikime-adk/templates"
	"jikime-adk/version"
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
	configDir := filepath.Join(caretDir, "config")

	reinitialized := false
	if exists(caretDir) {
		if !force {
			return InitializeResult{}, fmt.Errorf("project already initialized at %s", pi.path)
		}
		reinitialized = true
	}

	// Phase 1: Create directory structure
	if err := os.MkdirAll(configDir, 0o755); err != nil {
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
	locale := languageCodeToLocale(answers.Locale)
	context := map[string]string{
		"PROJECT_NAME":               answers.ProjectName,
		"CONVERSATION_LANGUAGE":      answers.Locale,
		"CONVERSATION_LANGUAGE_NAME": conversationLanguageName,
		"GIT_COMMIT_LANG":            answers.GitCommitLang,
		"CODE_COMMENT_LANG":          answers.CodeCommentLang,
		"DOCUMENTATION_LANG":         answers.DocLang,
		"GIT_MODE":                   answers.GitMode,
		"GITHUB_USER":                answers.GitHubUser,
		"USER_NAME":                  answers.UserName,
		"HONORIFIC":                  answers.Honorific,
		"TONE_PRESET":                answers.TonePreset,
		"LOCALE":                     locale,
		"JIKIME_VERSION":             version.String(),
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

	// Phase 3: Initialize git repository (if not already initialized)
	initializeGit(pi.path)

	return InitializeResult{
		ProjectPath:   pi.path,
		Locale:        answers.Locale,
		Language:      answers.ProjectName,
		CreatedFiles:  created,
		Duration:      time.Since(start),
		Reinitialized: reinitialized,
	}, nil
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

// languageCodeToLocale converts a language code to its full locale
func languageCodeToLocale(code string) string {
	locales := map[string]string{
		"en": "en_US",
		"ko": "ko_KR",
		"ja": "ja_JP",
		"zh": "zh_CN",
		"es": "es_ES",
		"fr": "fr_FR",
		"de": "de_DE",
	}
	if locale, ok := locales[code]; ok {
		return locale
	}
	return "en_US"
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

