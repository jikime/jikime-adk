// Package tag provides TAG parsing for TAG System v2.0.
package tag

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTagsFromSource extracts all @SPEC TAGs from source code.
// Works with any language that uses # for comments (Python, shell, YAML, etc.)
func ExtractTagsFromSource(source, filePath string) []*TAG {
	var tags []*TAG

	scanner := bufio.NewScanner(strings.NewReader(source))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Find comment portion (from # to end of line)
		commentStart := strings.Index(line, "#")
		if commentStart == -1 {
			continue
		}

		// Extract the comment portion
		commentPortion := line[commentStart:]

		// Parse TAG from comment
		tag := ParseTAGString(commentPortion, filePath, lineNumber)
		if tag != nil {
			tags = append(tags, tag)
		}
	}

	return tags
}

// ExtractTagsFromFile extracts all @SPEC TAGs from a file.
// Returns empty slice if file doesn't exist or has errors.
func ExtractTagsFromFile(filePath string) []*TAG {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return []*TAG{}
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []*TAG{}
	}

	return ExtractTagsFromSource(string(content), filePath)
}

// ExtractTagsFromFiles extracts TAGs from multiple files.
func ExtractTagsFromFiles(filePaths []string) []*TAG {
	var allTags []*TAG

	for _, filePath := range filePaths {
		tags := ExtractTagsFromFile(filePath)
		allTags = append(allTags, tags...)
	}

	return allTags
}

// ExtractTagsFromDirectory extracts TAGs from all matching files in a directory.
// pattern: glob pattern to match files (e.g., "*.py", "*.go")
// recursive: whether to search subdirectories
func ExtractTagsFromDirectory(directory, pattern string, recursive bool) []*TAG {
	// Check if directory exists
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return []*TAG{}
	}

	var files []string

	if recursive {
		// Walk through directory recursively
		err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue on error
			}
			if info.IsDir() {
				return nil
			}
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				files = append(files, path)
			}
			return nil
		})
	} else {
		// Non-recursive glob
		matches, err := filepath.Glob(filepath.Join(directory, pattern))
		if err == nil {
			for _, match := range matches {
				info, err := os.Stat(match)
				if err == nil && !info.IsDir() {
					files = append(files, match)
				}
			}
		}
	}

	return ExtractTagsFromFiles(files)
}

// SupportedExtensions lists file extensions that support # comments.
var SupportedExtensions = []string{
	".py",     // Python
	".sh",     // Shell
	".bash",   // Bash
	".yaml",   // YAML
	".yml",    // YAML
	".toml",   // TOML
	".rb",     // Ruby
	".pl",     // Perl
	".r",      // R
	".coffee", // CoffeeScript
}

// IsSupportedFile checks if a file extension supports # comments.
func IsSupportedFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supported := range SupportedExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

// ExtractTagsFromSupportedFiles extracts TAGs only from files with supported extensions.
func ExtractTagsFromSupportedFiles(directory string, recursive bool) []*TAG {
	var allTags []*TAG

	for _, ext := range SupportedExtensions {
		pattern := "*" + ext
		tags := ExtractTagsFromDirectory(directory, pattern, recursive)
		allTags = append(allTags, tags...)
	}

	return allTags
}
