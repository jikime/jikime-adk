package hookscmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// PostToolAstGrepCmd represents the post-tool-ast-grep hook command
var PostToolAstGrepCmd = &cobra.Command{
	Use:   "post-tool-ast-grep",
	Short: "Automatic AST-Grep security scanning after file modifications",
	Long: `PostToolUse hook that automatically runs AST-Grep security scanning
after Claude writes or edits files to detect potential security vulnerabilities.

Supports multiple languages:
- Python (.py, .pyi)
- JavaScript/TypeScript (.js, .jsx, .ts, .tsx)
- Go (.go)
- Rust (.rs)
- Java (.java)
- And more...

Exit code 2 indicates security issues that need attention.`,
	RunE: runPostToolAstGrep,
}

// Supported extensions for AST-Grep scanning
var astGrepSupportedExtensions = map[string]bool{
	".py": true, ".pyi": true,
	".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
	".ts": true, ".tsx": true, ".mts": true, ".cts": true,
	".go":   true,
	".rs":   true,
	".java": true,
	".kt":   true, ".kts": true,
	".c": true, ".cpp": true, ".cc": true, ".h": true, ".hpp": true,
	".rb":  true,
	".php": true,
}

type astGrepInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type astGrepOutput struct {
	HookSpecificOutput *astGrepHookOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool               `json:"suppressOutput,omitempty"`
}

type astGrepHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

type astGrepFinding struct {
	RuleID   string `json:"ruleId"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Range    struct {
		Start struct {
			Line int `json:"line"`
		} `json:"start"`
	} `json:"range"`
}

type scanResult struct {
	Scanned      bool
	IssuesFound  int
	ErrorCount   int
	WarningCount int
	InfoCount    int
	Details      []scanDetail
	Error        string
}

type scanDetail struct {
	Rule     string
	Severity string
	Message  string
	Line     int
}

func runPostToolAstGrep(cmd *cobra.Command, args []string) error {
	// Check if scanning is disabled
	if disabled := os.Getenv("JIKIME_DISABLE_AST_GREP_SCAN"); disabled != "" {
		if disabled == "1" || strings.ToLower(disabled) == "true" || strings.ToLower(disabled) == "yes" {
			return nil
		}
	}

	// Read JSON input from stdin
	var input astGrepInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil
	}

	// Only process Write and Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return nil
	}

	// Get file path from tool input
	filePathRaw, ok := input.ToolInput["file_path"]
	if !ok {
		return nil
	}
	filePath, ok := filePathRaw.(string)
	if !ok || filePath == "" {
		return nil
	}

	// Check if file is scannable
	if !isScannable(filePath) {
		output := astGrepOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(output)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		output := astGrepOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(output)
	}

	// Find rules configuration
	configPath := findAstGrepConfig()

	// Run scan
	result := runAstGrepScan(filePath, configPath)

	// Format output
	context := formatScanResult(result, filePath)

	// Prepare hook output
	output := astGrepOutput{
		HookSpecificOutput: &astGrepHookOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: context,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(output); err != nil {
		return err
	}

	// Exit with attention code if errors found
	if result.ErrorCount > 0 {
		os.Exit(2)
	}

	return nil
}

func isScannable(filePath string) bool {
	if filePath == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	return astGrepSupportedExtensions[ext]
}

func findAstGrepConfig() string {
	projectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}

	// Check common locations for sgconfig.yml
	possiblePaths := []string{
		filepath.Join(projectDir, ".claude", "skills", "jikime-tool-ast-grep", "rules", "sgconfig.yml"),
		filepath.Join(projectDir, ".jikime", "skills", "jikime-tool-ast-grep", "rules", "sgconfig.yml"),
		filepath.Join(projectDir, "sgconfig.yml"),
		filepath.Join(projectDir, ".ast-grep", "sgconfig.yml"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func runAstGrepScan(filePath, configPath string) scanResult {
	result := scanResult{
		Scanned: false,
		Details: []scanDetail{},
	}

	// Check if sg (ast-grep) is available
	sgPath, err := exec.LookPath("sg")
	if err != nil {
		result.Error = "ast-grep (sg) not installed"
		return result
	}

	// Build command
	cmdArgs := []string{"scan", "--json"}
	if configPath != "" {
		cmdArgs = append(cmdArgs, "--config", configPath)
	}
	cmdArgs = append(cmdArgs, filePath)

	// Run scan with timeout
	cmd := exec.Command(sgPath, cmdArgs...)

	// Set project directory
	projectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	cmd.Dir = projectDir

	// Use a channel for timeout handling
	type cmdResult struct {
		output []byte
		err    error
	}
	done := make(chan cmdResult, 1)

	go func() {
		output, err := cmd.CombinedOutput()
		done <- cmdResult{output, err}
	}()

	select {
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		result.Error = "AST-Grep scan timed out"
		return result
	case res := <-done:
		result.Scanned = true

		if len(res.output) > 0 {
			// Parse JSON output
			var findings []astGrepFinding
			if err := json.Unmarshal(res.output, &findings); err == nil {
				for _, finding := range findings {
					severity := strings.ToLower(finding.Severity)
					if severity == "" {
						severity = "info"
					}

					switch severity {
					case "error":
						result.ErrorCount++
					case "warning":
						result.WarningCount++
					default:
						result.InfoCount++
					}

					result.Details = append(result.Details, scanDetail{
						Rule:     finding.RuleID,
						Severity: severity,
						Message:  finding.Message,
						Line:     finding.Range.Start.Line,
					})
				}
				result.IssuesFound = len(findings)
			}
		}
	}

	return result
}

func formatScanResult(result scanResult, filePath string) string {
	if result.Error != "" {
		return "AST-Grep scan skipped: " + result.Error
	}

	if !result.Scanned {
		return "AST-Grep scan not performed"
	}

	filename := filepath.Base(filePath)

	if result.IssuesFound == 0 {
		return "AST-Grep: No security issues found in " + filename
	}

	// Build summary
	var parts []string
	if result.ErrorCount > 0 {
		parts = append(parts, strconv.Itoa(result.ErrorCount)+" error(s)")
	}
	if result.WarningCount > 0 {
		parts = append(parts, strconv.Itoa(result.WarningCount)+" warning(s)")
	}
	if result.InfoCount > 0 {
		parts = append(parts, strconv.Itoa(result.InfoCount)+" info")
	}

	summary := "AST-Grep found " + strings.Join(parts, ", ") + " in " + filename

	// Add top 3 issues
	if len(result.Details) > 0 {
		var issues []string
		limit := 3
		if len(result.Details) < limit {
			limit = len(result.Details)
		}

		for i := 0; i < limit; i++ {
			detail := result.Details[i]
			issues = append(issues, "  - ["+strings.ToUpper(detail.Severity)+"] "+
				detail.Rule+": "+detail.Message+" (line "+strconv.Itoa(detail.Line)+")")
		}
		summary += "\n" + strings.Join(issues, "\n")

		if len(result.Details) > 3 {
			summary += "\n  ... and " + strconv.Itoa(len(result.Details)-3) + " more"
		}
	}

	return summary
}
