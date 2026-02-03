package memory

import (
	"os"
	"path/filepath"
)

// FindProjectRoot finds the project root by searching for .jikime directory
// from startDir upward to filesystem root.
//
// Returns the directory containing .jikime if found,
// otherwise returns startDir for backward compatibility.
func FindProjectRoot(startDir string) string {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return ""
		}
	}

	dir := startDir
	for {
		jikimePath := filepath.Join(dir, ".jikime")
		if info, err := os.Stat(jikimePath); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			break
		}
		dir = parent
	}

	// .jikime not found - return original directory for backward compatibility
	return startDir
}
