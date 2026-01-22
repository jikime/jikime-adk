package tagcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"jikime-adk/tag"
)

// ValidationConfig holds TAG validation settings.
type ValidationConfig struct {
	Enabled bool   // Whether validation is enabled
	Mode    string // warn | enforce | off
}

func newValidateCmd() *cobra.Command {
	var (
		mode    string
		specsDir string
		files   []string
	)

	cmd := &cobra.Command{
		Use:   "validate [files...]",
		Short: "Validate TAGs in files (pre-commit hook)",
		Long: `Validate TAGs in staged files for pre-commit validation.

Modes:
  warn     Show warnings but allow commit (default)
  enforce  Block commit on validation errors
  off      Skip validation

If no files specified, validates staged git files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get files to validate
			filesToValidate := args
			if len(filesToValidate) == 0 {
				// Get staged files from git
				staged, err := getStagedFiles()
				if err != nil {
					return nil // Not in git repo or no staged files
				}
				filesToValidate = staged
			}

			if len(filesToValidate) == 0 {
				fmt.Println("No files to validate")
				return nil
			}

			// Find specs directory
			if specsDir == "" {
				cwd, _ := os.Getwd()
				specsDir = filepath.Join(cwd, ".jikime", "specs")
			}

			// Validate files
			return validateFiles(filesToValidate, specsDir, mode)
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "warn", "Validation mode: warn, enforce, off")
	cmd.Flags().StringVar(&specsDir, "specs-dir", "", "SPEC documents directory")
	cmd.Flags().StringSliceVarP(&files, "files", "f", nil, "Files to validate")

	return cmd
}

// getStagedFiles returns a list of staged files from git.
func getStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" && isSupportedFile(line) {
			files = append(files, line)
		}
	}

	return files, nil
}

// isSupportedFile checks if a file supports TAG comments.
func isSupportedFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExts := map[string]bool{
		".py":     true,
		".sh":     true,
		".bash":   true,
		".yaml":   true,
		".yml":    true,
		".toml":   true,
		".rb":     true,
		".pl":     true,
		".r":      true,
		".coffee": true,
	}
	return supportedExts[ext]
}

// ValidationResult holds the result of TAG validation.
type ValidationResult struct {
	File     string
	Line     int
	Level    string // error, warning, hint
	Message  string
	TAG      *tag.TAG
}

// validateFiles validates TAGs in the given files.
func validateFiles(files []string, specsDir, mode string) error {
	if mode == "off" {
		fmt.Println("TAG validation mode: off")
		return nil
	}

	var results []ValidationResult
	hasErrors := false

	for _, file := range files {
		if !fileExists(file) {
			continue
		}

		tags := tag.ExtractTagsFromFile(file)
		if len(tags) == 0 {
			continue
		}

		for _, t := range tags {
			// Validate TAG format
			if !tag.ValidateSpecIDFormat(t.SpecID) {
				results = append(results, ValidationResult{
					File:    file,
					Line:    t.Line,
					Level:   "error",
					Message: fmt.Sprintf("Invalid SPEC-ID format: %s", t.SpecID),
					TAG:     t,
				})
				hasErrors = true
				continue
			}

			if !tag.ValidateVerb(t.Verb) {
				results = append(results, ValidationResult{
					File:    file,
					Line:    t.Line,
					Level:   "error",
					Message: fmt.Sprintf("Invalid verb: %s", t.Verb),
					TAG:     t,
				})
				hasErrors = true
				continue
			}

			// Check SPEC existence
			if !tag.SpecDocumentExists(t.SpecID, specsDir) {
				level := getValidationLevel(mode)
				message := fmt.Sprintf("SPEC document not found: %s", t.SpecID)

				switch t.Verb {
				case "related":
					level = "hint"
					message = fmt.Sprintf("Consider creating SPEC for %s (related TAG - no SPEC required)", t.SpecID)
				case "verify":
					if isTestFile(file) {
						level = "warning"
						message = fmt.Sprintf("TEST references missing SPEC: %s (verify TAG in test file)", t.SpecID)
					}
				case "impl":
					if mode == "enforce" {
						level = "error"
						hasErrors = true
						message = fmt.Sprintf("Implementation references missing SPEC: %s (commit blocked in enforce mode)", t.SpecID)
					}
				}

				results = append(results, ValidationResult{
					File:    file,
					Line:    t.Line,
					Level:   level,
					Message: message,
					TAG:     t,
				})
			}
		}
	}

	// Display results
	if len(results) > 0 {
		fmt.Println()
		color.Cyan("TAG Validation Results:")
		fmt.Println(strings.Repeat("=", 60))

		for _, r := range results {
			var prefix string
			switch r.Level {
			case "error":
				prefix = color.RedString("ERROR")
			case "warning":
				prefix = color.YellowString("WARNING")
			case "hint":
				prefix = color.HiBlackString("HINT")
			}
			fmt.Printf("  [%s] %s:%d: %s\n", prefix, r.File, r.Line, r.Message)
		}

		fmt.Println(strings.Repeat("=", 60))

		if mode == "enforce" && hasErrors {
			color.Red("\nCommit blocked due to TAG validation errors (enforce mode)")
			fmt.Println("Fix the errors and try again, or use --no-verify to bypass")
			return fmt.Errorf("TAG validation failed")
		}

		color.Yellow("\nCommit allowed with warnings (warn mode)")
		fmt.Println("Consider fixing the TAG validation errors")
	} else {
		color.Green("TAG validation: All TAGs valid")
	}

	return nil
}

// getValidationLevel determines the validation level based on context.
func getValidationLevel(mode string) string {
	if mode == "enforce" {
		return "error"
	}
	return "warning"
}

// isTestFile checks if a file is a test file.
func isTestFile(file string) bool {
	name := filepath.Base(file)
	return strings.HasPrefix(name, "test_") ||
		strings.HasSuffix(name, "_test.py") ||
		strings.Contains(file, "/tests/") ||
		strings.Contains(file, "/test/")
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
