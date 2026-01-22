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

// PostToolLspCmd represents the post-tool-lsp hook command
var PostToolLspCmd = &cobra.Command{
	Use:   "post-tool-lsp",
	Short: "LSP diagnostics after file modifications",
	Long: `PostToolUse hook that runs diagnostics after file modifications.

Uses external tools as fallback when LSP is not available:
- Python: ruff check
- TypeScript: tsc --noEmit
- Go: go vet

Exit code 2 indicates errors that need attention.`,
	RunE: runPostToolLsp,
}

// Supported file extensions for LSP diagnostics
var lspSupportedExtensions = map[string]string{
	".py": "python", ".pyi": "python",
	".ts": "typescript", ".tsx": "typescriptreact",
	".js": "javascript", ".jsx": "javascriptreact",
	".mjs": "javascript", ".cjs": "javascript",
	".mts": "typescript", ".cts": "typescript",
	".go":   "go",
	".rs":   "rust",
	".java": "java",
	".kt":   "kotlin", ".kts": "kotlin",
	".rb":  "ruby",
	".php": "php",
	".c":   "c", ".cpp": "cpp",
	".h": "c", ".hpp": "cpp",
}

type lspInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type lspOutput struct {
	HookSpecificOutput *lspHookOutput `json:"hookSpecificOutput,omitempty"`
}

type lspHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

type lspDiagnosticResult struct {
	Available    bool
	ErrorCount   int
	WarningCount int
	InfoCount    int
	Diagnostics  []lspDiagnostic
	Error        string
	Fallback     bool
}

type lspDiagnostic struct {
	Severity string
	Message  string
	Line     int
	Source   string
	Code     string
}

func runPostToolLsp(cmd *cobra.Command, args []string) error {
	// Check if diagnostics are disabled
	if disabled := os.Getenv("JIKIME_DISABLE_LSP_DIAGNOSTIC"); disabled != "" {
		if disabled == "1" || strings.ToLower(disabled) == "true" || strings.ToLower(disabled) == "yes" {
			return nil
		}
	}

	// Read JSON input from stdin
	var input lspInput
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

	// Check if file is supported
	if !isLspSupported(filePath) {
		output := lspOutput{
			HookSpecificOutput: &lspHookOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: "LSP: File type not supported for diagnostics",
			},
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(output)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	// Run fallback diagnostics
	result := runFallbackDiagnostics(filePath)

	// Format output
	context := formatLspDiagnosticOutput(result, filePath)

	// Prepare hook output
	output := lspOutput{
		HookSpecificOutput: &lspHookOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: context,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(output); err != nil {
		return err
	}

	// Record to loop snapshot if loop is active
	if IsLoopActive() {
		AppendDiagnosticToSnapshot(DiagnosticEntry{
			Source:       "lsp",
			FilePath:     filePath,
			ErrorCount:   result.ErrorCount,
			WarningCount: result.WarningCount,
		})
	}

	// Exit with attention code if errors found
	if result.ErrorCount > 0 {
		os.Exit(2)
	}

	return nil
}

func isLspSupported(filePath string) bool {
	if filePath == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	_, ok := lspSupportedExtensions[ext]
	return ok
}

func getLspLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	return lspSupportedExtensions[ext]
}

func runFallbackDiagnostics(filePath string) lspDiagnosticResult {
	result := lspDiagnosticResult{
		Diagnostics: []lspDiagnostic{},
		Fallback:    true,
	}

	language := getLspLanguage(filePath)
	if language == "" {
		return result
	}

	switch language {
	case "python":
		runPythonDiagnostics(filePath, &result)
	case "typescript", "typescriptreact":
		runTypeScriptDiagnostics(filePath, &result)
	case "go":
		runGoDiagnostics(filePath, &result)
	}

	return result
}

func runPythonDiagnostics(filePath string, result *lspDiagnosticResult) {
	// Try ruff
	if _, err := exec.LookPath("ruff"); err == nil {
		cmd := exec.Command("ruff", "check", "--output-format=json", filePath)

		// Use timeout
		done := make(chan error, 1)
		var output []byte
		go func() {
			var err error
			output, err = cmd.CombinedOutput()
			done <- err
		}()

		select {
		case <-time.After(30 * time.Second):
			cmd.Process.Kill()
			result.Error = "Ruff check timed out"
			return
		case <-done:
			if len(output) > 0 {
				var issues []map[string]interface{}
				if err := json.Unmarshal(output, &issues); err == nil {
					result.Available = true
					for i, issue := range issues {
						if i >= 10 {
							break
						}
						code, _ := issue["code"].(string)
						severity := "warning"
						if strings.HasPrefix(code, "E") || strings.HasPrefix(code, "F") {
							severity = "error"
							result.ErrorCount++
						} else {
							result.WarningCount++
						}

						message, _ := issue["message"].(string)
						location, _ := issue["location"].(map[string]interface{})
						row := 0
						if location != nil {
							if r, ok := location["row"].(float64); ok {
								row = int(r)
							}
						}

						result.Diagnostics = append(result.Diagnostics, lspDiagnostic{
							Severity: severity,
							Message:  message,
							Line:     row,
							Source:   "ruff",
							Code:     code,
						})
					}
				}
			}
		}
	}
}

func runTypeScriptDiagnostics(filePath string, result *lspDiagnosticResult) {
	// Try tsc
	if _, err := exec.LookPath("tsc"); err == nil {
		cmd := exec.Command("tsc", "--noEmit", "--pretty", "false", filePath)

		// Use timeout
		done := make(chan error, 1)
		var output []byte
		go func() {
			var err error
			output, err = cmd.CombinedOutput()
			done <- err
		}()

		select {
		case <-time.After(60 * time.Second):
			cmd.Process.Kill()
			result.Error = "TypeScript check timed out"
			return
		case <-done:
			if len(output) > 0 {
				result.Available = true
				lines := strings.Split(string(output), "\n")
				for i, line := range lines {
					if i >= 10 {
						break
					}
					line = strings.TrimSpace(line)
					if strings.Contains(strings.ToLower(line), "error") {
						result.ErrorCount++
						result.Diagnostics = append(result.Diagnostics, lspDiagnostic{
							Severity: "error",
							Message:  line,
							Line:     0,
							Source:   "tsc",
						})
					}
				}
			}
		}
	}
}

func runGoDiagnostics(filePath string, result *lspDiagnosticResult) {
	// Try go vet
	if _, err := exec.LookPath("go"); err == nil {
		cmd := exec.Command("go", "vet", filePath)

		// Use timeout
		done := make(chan error, 1)
		var output []byte
		go func() {
			var err error
			output, err = cmd.CombinedOutput()
			done <- err
		}()

		select {
		case <-time.After(30 * time.Second):
			cmd.Process.Kill()
			result.Error = "Go vet timed out"
			return
		case <-done:
			result.Available = true
			if len(output) > 0 {
				lines := strings.Split(string(output), "\n")
				for i, line := range lines {
					if i >= 10 {
						break
					}
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "#") {
						result.ErrorCount++
						result.Diagnostics = append(result.Diagnostics, lspDiagnostic{
							Severity: "error",
							Message:  line,
							Line:     0,
							Source:   "go vet",
						})
					}
				}
			}
		}
	}
}

func formatLspDiagnosticOutput(result lspDiagnosticResult, filePath string) string {
	filename := filepath.Base(filePath)

	if result.Error != "" {
		return "LSP: Diagnostics unavailable for " + filename + " (" + result.Error + ")"
	}

	if !result.Available {
		return "LSP: No diagnostics available for " + filename
	}

	// No issues found
	total := result.ErrorCount + result.WarningCount + result.InfoCount
	if total == 0 {
		return "LSP: No issues in " + filename
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

	summary := "LSP: " + strings.Join(parts, ", ") + " in " + filename

	// Add top diagnostics
	if len(result.Diagnostics) > 0 {
		var issues []string
		limit := 5
		if len(result.Diagnostics) < limit {
			limit = len(result.Diagnostics)
		}

		for i := 0; i < limit; i++ {
			diag := result.Diagnostics[i]
			sev := strings.ToUpper(diag.Severity)
			msg := diag.Message
			if len(msg) > 100 {
				msg = msg[:100]
			}
			line := strconv.Itoa(diag.Line)
			if diag.Line == 0 {
				line = "?"
			}
			sourceInfo := ""
			if diag.Source != "" {
				sourceInfo = " [" + diag.Source + "]"
			}
			issues = append(issues, "  - ["+sev+"] Line "+line+": "+msg+sourceInfo)
		}
		summary += "\n" + strings.Join(issues, "\n")

		if len(result.Diagnostics) > 5 {
			summary += "\n  ... and " + strconv.Itoa(len(result.Diagnostics)-5) + " more"
		}
	}

	return summary
}
