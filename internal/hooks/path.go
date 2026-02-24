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

// FindJikimeDir returns the .jikime directory path
func FindJikimeDir() (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, ".jikime"), nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
