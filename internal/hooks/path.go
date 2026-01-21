package hooks

import (
	"os"
	"path/filepath"
)

// FindProjectRoot finds the project root by looking for .jikime directory
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		jikimePath := filepath.Join(dir, ".jikime")
		if info, err := os.Stat(jikimePath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, return current working directory
			return os.Getwd()
		}
		dir = parent
	}
}

// FindClaudeDir returns the .claude directory path
func FindClaudeDir() (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, ".claude"), nil
}

// FindJikimeDir returns the .jikime directory path
func FindJikimeDir() (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, ".jikime"), nil
}

// FindConfigDir returns the .jikime/config directory path
func FindConfigDir() (string, error) {
	jikimeDir, err := FindJikimeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(jikimeDir, "config"), nil
}

// FindSectionsDir returns the .jikime/config/sections directory path
func FindSectionsDir() (string, error) {
	configDir, err := FindConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sections"), nil
}

// FindCacheDir returns the .jikime/cache directory path
func FindCacheDir() (string, error) {
	jikimeDir, err := FindJikimeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(jikimeDir, "cache")
	// Create if not exists
	os.MkdirAll(cacheDir, 0o755)
	return cacheDir, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
