// Package doctorcmd provides the doctor command for jikime-adk.
package doctorcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// CheckResult represents the result of a diagnostic check
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // pass, warn, fail
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// DiagnosticReport represents the full diagnostic report
type DiagnosticReport struct {
	System       []CheckResult `json:"system"`
	Tools        []CheckResult `json:"tools"`
	Project      []CheckResult `json:"project"`
	LanguageSDK  []CheckResult `json:"language_sdk,omitempty"`
}

// NewDoctor creates the doctor command.
func NewDoctor() *cobra.Command {
	var (
		verbose       bool
		fix           bool
		export        string
		checkCommands bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system environment and diagnose issues",
		Long:  "Run diagnostic checks on your development environment, tools, and project configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(verbose, fix, export, checkCommands)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	cmd.Flags().BoolVar(&fix, "fix", false, "Attempt to fix detected issues")
	cmd.Flags().StringVar(&export, "export", "", "Export report to file (json format)")
	cmd.Flags().BoolVar(&checkCommands, "check-commands", false, "Check available slash commands")

	return cmd
}

func runDoctor(verbose, fix bool, export string, checkCommands bool) error {
	cyan := color.New(color.FgCyan, color.Bold)

	cyan.Println("╔════════════════════════════════════════╗")
	cyan.Println("║       Jikime-ADK Doctor                ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	report := &DiagnosticReport{}

	// System checks
	fmt.Println("System Environment")
	fmt.Println("──────────────────")
	report.System = checkSystem(verbose)
	fmt.Println()

	// Tool checks
	fmt.Println("Required Tools")
	fmt.Println("──────────────")
	report.Tools = checkTools(verbose)
	fmt.Println()

	// Project checks
	fmt.Println("Project Structure")
	fmt.Println("─────────────────")
	report.Project = checkProject(verbose)
	fmt.Println()

	// Language-specific checks
	lang := detectProjectLanguage()
	if lang != "" {
		fmt.Printf("Language Tools (%s)\n", lang)
		fmt.Println("────────────────────")
		report.LanguageSDK = checkLanguageTools(lang, verbose)
		fmt.Println()
	}

	// Check commands if requested
	if checkCommands {
		fmt.Println("Slash Commands")
		fmt.Println("──────────────")
		checkSlashCommands(verbose)
		fmt.Println()
	}

	// Print summary
	printSummary(report)

	// Export if requested
	if export != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		if err := os.WriteFile(export, data, 0644); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		fmt.Printf("\nReport exported to: %s\n", export)
	}

	// Attempt fixes if requested
	if fix {
		fmt.Println("\nAttempting fixes...")
		attemptFixes(report)
	}

	return nil
}

func checkSystem(verbose bool) []CheckResult {
	results := []CheckResult{}

	// OS info
	osResult := CheckResult{
		Name:    "Operating System",
		Status:  "pass",
		Message: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	printCheck(osResult)
	results = append(results, osResult)

	// Go version
	goResult := checkCommand("go", "version", "Go Runtime")
	printCheck(goResult)
	results = append(results, goResult)

	// Home directory
	home, err := os.UserHomeDir()
	homeResult := CheckResult{Name: "Home Directory"}
	if err == nil {
		homeResult.Status = "pass"
		homeResult.Message = home
	} else {
		homeResult.Status = "fail"
		homeResult.Message = "Cannot determine home directory"
	}
	printCheck(homeResult)
	results = append(results, homeResult)

	return results
}

func checkTools(verbose bool) []CheckResult {
	results := []CheckResult{}

	tools := []struct {
		cmd     string
		args    string
		name    string
		required bool
	}{
		{"git", "--version", "Git", true},
		{"claude", "--version", "Claude CLI", true},
		{"node", "--version", "Node.js", false},
		{"npm", "--version", "npm", false},
		{"python3", "--version", "Python 3", false},
		{"uv", "--version", "uv (Python)", false},
		{"ruff", "--version", "Ruff (Linter)", false},
		{"prettier", "--version", "Prettier", false},
	}

	for _, tool := range tools {
		result := checkCommand(tool.cmd, tool.args, tool.name)
		if !tool.required && result.Status == "fail" {
			result.Status = "warn"
		}
		printCheck(result)
		results = append(results, result)
	}

	return results
}

func checkProject(verbose bool) []CheckResult {
	results := []CheckResult{}
	cwd, _ := os.Getwd()

	// Check .claude directory
	claudeResult := CheckResult{Name: ".claude directory"}
	claudeDir := filepath.Join(cwd, ".claude")
	if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
		claudeResult.Status = "pass"
		claudeResult.Message = "Found"
	} else {
		claudeResult.Status = "warn"
		claudeResult.Message = "Not found"
	}
	printCheck(claudeResult)
	results = append(results, claudeResult)

	// Check .jikime directory
	jikimeResult := CheckResult{Name: ".jikime directory"}
	jikimeDir := filepath.Join(cwd, ".jikime")
	if info, err := os.Stat(jikimeDir); err == nil && info.IsDir() {
		jikimeResult.Status = "pass"
		jikimeResult.Message = "Found"
	} else {
		jikimeResult.Status = "warn"
		jikimeResult.Message = "Not found (optional)"
	}
	printCheck(jikimeResult)
	results = append(results, jikimeResult)

	// Check settings.json
	settingsResult := CheckResult{Name: "settings.json"}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		settingsResult.Status = "pass"
		settingsResult.Message = "Found"
	} else {
		settingsResult.Status = "warn"
		settingsResult.Message = "Not found"
	}
	printCheck(settingsResult)
	results = append(results, settingsResult)

	// Check CLAUDE.md
	claudeMdResult := CheckResult{Name: "CLAUDE.md"}
	claudeMdPath := filepath.Join(cwd, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); err == nil {
		claudeMdResult.Status = "pass"
		claudeMdResult.Message = "Found"
	} else {
		claudeMdResult.Status = "warn"
		claudeMdResult.Message = "Not found (optional)"
	}
	printCheck(claudeMdResult)
	results = append(results, claudeMdResult)

	// Check for git repository
	gitResult := CheckResult{Name: "Git repository"}
	gitDir := filepath.Join(cwd, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		gitResult.Status = "pass"
		gitResult.Message = "Initialized"
	} else {
		gitResult.Status = "warn"
		gitResult.Message = "Not initialized"
	}
	printCheck(gitResult)
	results = append(results, gitResult)

	return results
}

func detectProjectLanguage() string {
	cwd, _ := os.Getwd()

	// Check for various project files
	if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		return "javascript"
	}
	if _, err := os.Stat(filepath.Join(cwd, "pyproject.toml")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(cwd, "requirements.txt")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(cwd, "Cargo.toml")); err == nil {
		return "rust"
	}
	if _, err := os.Stat(filepath.Join(cwd, "pom.xml")); err == nil {
		return "java"
	}
	if _, err := os.Stat(filepath.Join(cwd, "build.gradle")); err == nil {
		return "java"
	}

	return ""
}

func checkLanguageTools(lang string, verbose bool) []CheckResult {
	results := []CheckResult{}

	switch lang {
	case "python":
		tools := []struct {
			cmd  string
			args string
			name string
		}{
			{"python3", "--version", "Python 3"},
			{"pip3", "--version", "pip"},
			{"uv", "--version", "uv"},
			{"ruff", "--version", "Ruff"},
			{"mypy", "--version", "mypy"},
			{"pytest", "--version", "pytest"},
		}
		for _, tool := range tools {
			result := checkCommand(tool.cmd, tool.args, tool.name)
			if result.Status == "fail" {
				result.Status = "warn"
			}
			printCheck(result)
			results = append(results, result)
		}

	case "javascript":
		tools := []struct {
			cmd  string
			args string
			name string
		}{
			{"node", "--version", "Node.js"},
			{"npm", "--version", "npm"},
			{"npx", "--version", "npx"},
			{"prettier", "--version", "Prettier"},
			{"eslint", "--version", "ESLint"},
		}
		for _, tool := range tools {
			result := checkCommand(tool.cmd, tool.args, tool.name)
			if result.Status == "fail" {
				result.Status = "warn"
			}
			printCheck(result)
			results = append(results, result)
		}

	case "go":
		tools := []struct {
			cmd  string
			args string
			name string
		}{
			{"go", "version", "Go"},
			{"golangci-lint", "--version", "golangci-lint"},
		}
		for _, tool := range tools {
			result := checkCommand(tool.cmd, tool.args, tool.name)
			if result.Status == "fail" {
				result.Status = "warn"
			}
			printCheck(result)
			results = append(results, result)
		}

	case "rust":
		tools := []struct {
			cmd  string
			args string
			name string
		}{
			{"rustc", "--version", "Rust"},
			{"cargo", "--version", "Cargo"},
			{"clippy-driver", "--version", "Clippy"},
		}
		for _, tool := range tools {
			result := checkCommand(tool.cmd, tool.args, tool.name)
			if result.Status == "fail" {
				result.Status = "warn"
			}
			printCheck(result)
			results = append(results, result)
		}
	}

	return results
}

func checkSlashCommands(verbose bool) {
	cwd, _ := os.Getwd()
	commandsDir := filepath.Join(cwd, ".claude", "commands")

	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)

	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		dim.Println("No commands directory found")
		return
	}

	// Walk through commands directory
	count := 0
	err := filepath.Walk(commandsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			relPath, _ := filepath.Rel(commandsDir, path)
			// Convert path to command name
			cmdName := strings.TrimSuffix(relPath, ".md")
			cmdName = strings.ReplaceAll(cmdName, string(os.PathSeparator), ":")
			fmt.Printf("  %s /%s\n", green.Sprint("•"), cmdName)
			count++
		}
		return nil
	})

	if err != nil {
		dim.Printf("Error reading commands: %v\n", err)
		return
	}

	fmt.Printf("\nTotal: %d commands\n", count)
}

func checkCommand(cmd, args, name string) CheckResult {
	result := CheckResult{Name: name}

	// Check if command exists
	path, err := exec.LookPath(cmd)
	if err != nil {
		result.Status = "fail"
		result.Message = "Not found"
		return result
	}

	// Get version
	argList := strings.Fields(args)
	execCmd := exec.Command(path, argList...)
	output, err := execCmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = "Error getting version"
		return result
	}

	// Extract version from output
	version := strings.TrimSpace(string(output))
	// Truncate long version strings
	if len(version) > 50 {
		version = version[:50] + "..."
	}
	// Clean up version string - take first line only
	if idx := strings.Index(version, "\n"); idx != -1 {
		version = version[:idx]
	}

	result.Status = "pass"
	result.Message = version
	return result
}

func printCheck(result CheckResult) {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	dim := color.New(color.Faint)

	var status string
	switch result.Status {
	case "pass":
		status = green.Sprint("✓")
	case "warn":
		status = yellow.Sprint("⚠")
	case "fail":
		status = red.Sprint("✗")
	}

	fmt.Printf("  %s %-20s %s\n", status, result.Name+":", dim.Sprint(result.Message))
}

func printSummary(report *DiagnosticReport) {
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	passes := 0
	warnings := 0
	failures := 0

	// Count all results
	allResults := append(report.System, report.Tools...)
	allResults = append(allResults, report.Project...)
	allResults = append(allResults, report.LanguageSDK...)

	for _, r := range allResults {
		switch r.Status {
		case "pass":
			passes++
		case "warn":
			warnings++
		case "fail":
			failures++
		}
	}

	fmt.Println("Summary")
	fmt.Println("───────")
	if failures > 0 {
		red.Printf("  ✗ %d issues need attention\n", failures)
	}
	if warnings > 0 {
		yellow.Printf("  ⚠ %d warnings\n", warnings)
	}
	if passes > 0 {
		green.Printf("  ✓ %d checks passed\n", passes)
	}

	if failures == 0 && warnings == 0 {
		green.Println("\n  All checks passed!")
	} else if failures == 0 {
		yellow.Println("\n  System is functional with some warnings.")
	} else {
		red.Println("\n  Please address the issues above.")
	}
}

func attemptFixes(report *DiagnosticReport) {
	dim := color.New(color.Faint)

	// Currently just provides suggestions
	for _, result := range report.Project {
		if result.Status == "fail" || result.Status == "warn" {
			switch result.Name {
			case ".claude directory":
				dim.Println("  • Run 'jikime-adk init' to create project structure")
			case "Git repository":
				dim.Println("  • Run 'git init' to initialize git repository")
			}
		}
	}
}
