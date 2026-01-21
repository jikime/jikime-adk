package hooks

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// RunCommand executes a command and returns its output
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RunCommandInDir executes a command in a specific directory
func RunCommandInDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RunCommandWithTimeout executes a command with a timeout
func RunCommandWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		return "", &TimeoutError{timeout}
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Duration time.Duration
}

func (e *TimeoutError) Error() string {
	return "command timed out after " + e.Duration.String()
}

// CommandExists checks if a command is available in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetFileExtension returns the file extension without the dot
func GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// IsPythonFile checks if a file is a Python file
func IsPythonFile(filename string) bool {
	ext := GetFileExtension(filename)
	return ext == "py" || ext == "pyw"
}

// IsJavaScriptFile checks if a file is a JavaScript/TypeScript file
func IsJavaScriptFile(filename string) bool {
	ext := GetFileExtension(filename)
	return ext == "js" || ext == "jsx" || ext == "ts" || ext == "tsx" || ext == "mjs" || ext == "cjs"
}

// IsGoFile checks if a file is a Go file
func IsGoFile(filename string) bool {
	return GetFileExtension(filename) == "go"
}

// SupportedLanguages returns the list of supported languages
func SupportedLanguages() []string {
	return []string{"en", "ko", "ja", "zh"}
}
