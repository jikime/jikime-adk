// Package updatecmd provides the update command for jikime-adk.
package updatecmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"jikime-adk/backup"
	"jikime-adk/version"
)

const (
	// GitHub releases API URL for jikime-adk
	releasesURL = "https://api.github.com/repos/jikime/jikime-adk/releases/latest"
	// Timeout for HTTP requests
	httpTimeout = 10 * time.Second
)

// NewUpdate creates the update command.
func NewUpdate() *cobra.Command {
	var (
		checkOnly    bool
		forceUpdate  bool
		skipBackup   bool
		syncTemplates bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update jikime-adk to the latest version",
		Long: `Update jikime-adk to the latest version available.

Includes:
- Version check against GitHub releases
- Automatic installer detection (binary, go install, brew)
- Atomic binary update with SHA256 checksum verification
- Template and configuration updates
- Backup before update with automatic rollback on failure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(checkOnly, forceUpdate, skipBackup, syncTemplates)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	cmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Force update even if already up to date")
	cmd.Flags().BoolVar(&skipBackup, "skip-backup", false, "Skip creating backup before update")
	cmd.Flags().BoolVar(&syncTemplates, "sync-templates", false, "Only sync templates without updating binary")

	return cmd
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	PublishedAt string        `json:"published_at"`
	HTMLURL     string        `json:"html_url"`
	Body        string        `json:"body"`
	Assets      []GitHubAsset `json:"assets"`
}

func runUpdate(checkOnly, forceUpdate, skipBackup, syncTemplates bool) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	dim := color.New(color.Faint)

	cyan.Println("╔════════════════════════════════════════╗")
	cyan.Println("║       Jikime-ADK Update                ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	currentVersion := version.String()
	fmt.Printf("%-20s %s\n", "Current version:", green.Sprint(currentVersion))

	// Check for latest version
	fmt.Printf("%-20s ", "Checking updates:")
	latestRelease, err := getLatestRelease()
	if err != nil {
		red.Printf("Failed (%v)\n", err)
		dim.Println("Using offline mode - only template sync available")
		fmt.Println()

		if syncTemplates {
			return syncProjectTemplates()
		}
		return nil
	}

	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")
	green.Printf("%s\n", latestVersion)
	fmt.Println()

	// Compare versions
	needsUpdate := compareVersions(currentVersion, latestVersion) < 0

	if !needsUpdate && !forceUpdate {
		green.Println("You are already running the latest version!")
		fmt.Println()

		if syncTemplates {
			return syncProjectTemplates()
		}

		dim.Println("Use --sync-templates to update project templates only.")
		return nil
	}

	if needsUpdate {
		yellow.Printf("New version available: %s → %s\n", currentVersion, latestVersion)
	} else {
		dim.Println("Forcing update as requested...")
	}
	fmt.Println()

	if checkOnly {
		fmt.Println("Release notes:")
		fmt.Println("──────────────")
		if latestRelease.Body != "" {
			// Truncate release notes if too long
			body := latestRelease.Body
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			dim.Println(body)
		}
		fmt.Println()
		dim.Printf("URL: %s\n", latestRelease.HTMLURL)
		return nil
	}

	// Detect installation method
	installer := detectInstaller()
	fmt.Printf("%-20s %s\n", "Installer:", installer)
	fmt.Println()

	// Create backup if requested
	if !skipBackup {
		fmt.Println("Creating backup...")
		backupPath, err := createBackup()
		if err != nil {
			yellow.Printf("Warning: Failed to create backup: %v\n", err)
			dim.Println("Continuing without backup...")
		} else {
			green.Println("Backup created successfully")
			dim.Printf("Location: %s\n", backupPath)
		}
		fmt.Println()
	}

	// Perform update
	fmt.Println("Updating jikime-adk...")
	if err := performUpdate(installer, latestRelease); err != nil {
		red.Printf("Update failed: %v\n", err)
		return err
	}

	green.Println("Update completed successfully!")
	fmt.Println()

	// Sync templates if in a project directory
	if err := syncProjectTemplates(); err != nil {
		yellow.Printf("Warning: Template sync failed: %v\n", err)
	}

	return nil
}

func getLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest("GET", releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "jikime-adk/"+version.String())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

func compareVersions(v1, v2 string) int {
	// Simple version comparison
	// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &n1)
		}
		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

func detectInstaller() string {
	execPath, err := os.Executable()
	if err != nil {
		return "binary"
	}

	// Resolve symlinks to get real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Check if installed via go install (GOPATH/bin or GOBIN)
	gobin := os.Getenv("GOBIN")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}

	gopathBin := filepath.Join(gopath, "bin")
	if gobin != "" && strings.HasPrefix(realPath, gobin) {
		return "go install"
	}
	if strings.HasPrefix(realPath, gopathBin) {
		return "go install"
	}

	// Check if installed via brew
	if _, err := exec.LookPath("brew"); err == nil {
		cmd := exec.Command("brew", "list", "jikime-adk")
		if cmd.Run() == nil {
			return "brew"
		}
	}

	// Otherwise it's a standalone binary (from GitHub Releases)
	return "binary"
}

func performUpdate(installer string, release *GitHubRelease) error {
	switch installer {
	case "go install":
		cmd := exec.Command("go", "install", "github.com/jikime/jikime-adk@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "brew":
		cmd := exec.Command("brew", "upgrade", "jikime-adk")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "binary":
		return performBinaryUpdate(release)
	default:
		return fmt.Errorf("unknown installer: %s", installer)
	}
}

func performBinaryUpdate(release *GitHubRelease) error {
	dim := color.New(color.Faint)

	// Find the matching asset for current platform
	assetName := fmt.Sprintf("jikime-adk-%s-%s", runtime.GOOS, runtime.GOARCH)
	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	// Download to temp file
	tmpDir, err := os.MkdirTemp("", "jikime-adk-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpPath := filepath.Join(tmpDir, assetName)
	dim.Printf("  Downloading %s...\n", assetName)
	if err := downloadFile(assetURL, tmpPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify checksum
	checksumURL := findChecksumAssetURL(release.Assets)
	if checksumURL != "" {
		dim.Println("  Verifying checksum...")
		if err := verifyChecksum(tmpPath, assetName, checksumURL); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Make downloaded file executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("cannot set executable permission: %w", err)
	}

	// Atomic replace: current → .bak, new → current
	backupPath := execPath + ".bak"

	// Remove existing backup if present
	os.Remove(backupPath)

	// Rename current binary to .bak
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("cannot backup current binary: %w", err)
	}

	// Copy new binary to target path
	if err := copyFile(tmpPath, execPath); err != nil {
		// Rollback on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("cannot install new binary: %w", err)
	}

	// Verify the new binary works
	cmd := exec.Command(execPath, "--version")
	if err := cmd.Run(); err != nil {
		// Rollback on failure
		os.Remove(execPath)
		os.Rename(backupPath, execPath)
		return fmt.Errorf("new binary verification failed, rolled back: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

func downloadFile(url, destPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "jikime-adk/"+version.String())

	// Support GITHUB_TOKEN for rate limiting
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyChecksum(filePath, fileName, checksumURL string) error {
	// Download checksums.txt
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest("GET", checksumURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "jikime-adk/"+version.String())
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot download checksums: HTTP %d", resp.StatusCode)
	}

	// Parse checksums.txt to find expected hash
	var expectedHash string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == fileName {
			expectedHash = parts[0]
			break
		}
	}
	if expectedHash == "" {
		return fmt.Errorf("checksum not found for %s", fileName)
	}

	// Calculate actual hash
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actualHash := hex.EncodeToString(h.Sum(nil))

	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func findChecksumAssetURL(assets []GitHubAsset) string {
	for _, asset := range assets {
		if asset.Name == "checksums.txt" {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}

// createBackup creates a backup using the backup package.
// Returns the backup path on success.
func createBackup() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Create TemplateBackup instance
	tb, err := backup.NewTemplateBackup(cwd)
	if err != nil {
		return "", err
	}

	// Check if there are any files to backup
	if !tb.HasExistingFiles() {
		return "", fmt.Errorf("not a jikime project")
	}

	// Create backup with metadata
	backupPath, err := tb.CreateBackup()
	if err != nil {
		return "", err
	}

	// Auto-cleanup old backups (keep last 5)
	const keepCount = 5
	deletedCount, _ := tb.CleanupOldBackups(keepCount)
	if deletedCount > 0 {
		dim := color.New(color.Faint)
		dim.Printf("Cleaned up %d old backup(s)\n", deletedCount)
	}

	return backupPath, nil
}


func syncProjectTemplates() error {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check if we're in a jikime project
	jikimeDir := filepath.Join(cwd, ".jikime")
	claudeDir := filepath.Join(cwd, ".claude")

	if _, err := os.Stat(jikimeDir); os.IsNotExist(err) {
		if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
			dim.Println("Not a jikime project - skipping template sync")
			return nil
		}
	}

	cyan.Println("Syncing project templates...")

	// Find template source
	// Templates are typically installed with the package
	templateSrc := findTemplateSource()
	if templateSrc == "" {
		dim.Println("Template source not found - skipping sync")
		return nil
	}

	// Sync templates (preserving user customizations)
	// This is a simplified version - full implementation would merge carefully
	syncCount := 0

	// Sync .claude directory
	srcClaude := filepath.Join(templateSrc, ".claude")
	if _, err := os.Stat(srcClaude); err == nil {
		if err := syncTemplateDir(srcClaude, claudeDir); err == nil {
			syncCount++
		}
	}

	// Sync .jikime/config (but not user data)
	srcJikimeConfig := filepath.Join(templateSrc, ".jikime", "config")
	dstJikimeConfig := filepath.Join(jikimeDir, "config")
	if _, err := os.Stat(srcJikimeConfig); err == nil {
		if err := syncTemplateDir(srcJikimeConfig, dstJikimeConfig); err == nil {
			syncCount++
		}
	}

	if syncCount > 0 {
		green.Printf("Synced %d template directories\n", syncCount)
	} else {
		dim.Println("No templates to sync")
	}

	return nil
}

func findTemplateSource() string {
	// Check common locations for templates

	// 1. Check JIKIME_TEMPLATE_DIR environment variable
	if templateDir := os.Getenv("JIKIME_TEMPLATE_DIR"); templateDir != "" {
		if _, err := os.Stat(templateDir); err == nil {
			return templateDir
		}
	}

	// 2. Check relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		templateDir := filepath.Join(execDir, "templates")
		if _, err := os.Stat(templateDir); err == nil {
			return templateDir
		}
	}

	// 3. Check home directory
	home, err := os.UserHomeDir()
	if err == nil {
		templateDir := filepath.Join(home, ".jikime", "templates")
		if _, err := os.Stat(templateDir); err == nil {
			return templateDir
		}
	}

	return ""
}

func syncTemplateDir(src, dst string) error {
	// Ensure destination exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Walk through source and copy files that don't exist in destination
	// or are newer in source
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Skip if destination file exists and is newer
		dstInfo, err := os.Stat(dstPath)
		if err == nil {
			if dstInfo.ModTime().After(info.ModTime()) {
				return nil // Skip, destination is newer
			}
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
