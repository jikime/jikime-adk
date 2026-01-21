package hookscmd

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// PostToolFormatterCmd represents the post-tool-formatter hook command
var PostToolFormatterCmd = &cobra.Command{
	Use:   "post-tool-formatter",
	Short: "Automatic code formatting after file modifications",
	Long: `PostToolUse hook that automatically formats code after Claude writes or edits files.

Supports multiple languages with automatic tool detection:
- Python: ruff format, black
- JavaScript/TypeScript: prettier, biome
- Go: gofmt, goimports
- Rust: rustfmt
- And more...`,
	RunE: runPostToolFormatter,
}

// File extensions to skip
var skipExtensions = map[string]bool{
	".json": true, ".lock": true, ".min.js": true, ".min.css": true,
	".map": true, ".svg": true, ".png": true, ".jpg": true,
	".gif": true, ".ico": true, ".woff": true, ".woff2": true,
	".ttf": true, ".eot": true, ".pdf": true, ".zip": true,
}

// Directories to skip
var skipDirectories = map[string]bool{
	"node_modules": true, ".git": true, ".venv": true, "venv": true,
	"__pycache__": true, ".cache": true, "dist": true, "build": true,
	".next": true, ".nuxt": true, "target": true, "vendor": true,
}

// Formatter definitions by file extension
type formatterDef struct {
	command string
	args    []string
}

var formatters = map[string][]formatterDef{
	".py": {
		{command: "ruff", args: []string{"format", "--quiet"}},
		{command: "black", args: []string{"--quiet"}},
	},
	".js": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
		{command: "biome", args: []string{"format", "--write"}},
	},
	".jsx": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
		{command: "biome", args: []string{"format", "--write"}},
	},
	".ts": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
		{command: "biome", args: []string{"format", "--write"}},
	},
	".tsx": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
		{command: "biome", args: []string{"format", "--write"}},
	},
	".go": {
		{command: "gofmt", args: []string{"-w"}},
	},
	".rs": {
		{command: "rustfmt", args: []string{}},
	},
	".css": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
	},
	".scss": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
	},
	".html": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
	},
	".yaml": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
	},
	".yml": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
	},
	".md": {
		{command: "prettier", args: []string{"--write", "--log-level", "error"}},
	},
}

type postToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type postToolOutput struct {
	HookSpecificOutput *postHookSpecificOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool                    `json:"suppressOutput,omitempty"`
}

type postHookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runPostToolFormatter(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input postToolInput
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

	// Check if we should skip
	shouldSkip, _ := shouldSkipFormatting(filePath)
	if shouldSkip {
		output := postToolOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(output)
	}

	// Get formatters for this file
	ext := strings.ToLower(filepath.Ext(filePath))
	formatterList, ok := formatters[ext]
	if !ok {
		output := postToolOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(output)
	}

	// Try formatters until one succeeds
	var formattedWith string
	var formatError string

	for _, formatter := range formatterList {
		// Check if formatter exists
		if _, err := exec.LookPath(formatter.command); err != nil {
			continue
		}

		// Run formatter
		args := append(formatter.args, filePath)
		cmd := exec.Command(formatter.command, args...)
		output, err := cmd.CombinedOutput()

		if err == nil {
			formattedWith = formatter.command
			break
		} else {
			formatError = strings.TrimSpace(string(output))
		}
	}

	// Build context messages
	var contextMessages []string

	if formattedWith != "" {
		contextMessages = append(contextMessages, "Auto-formatted with "+formattedWith)
	} else if formatError != "" {
		contextMessages = append(contextMessages, "Format warning: "+formatError)
	}

	// Check for console.log in JS/TS files
	if isJavaScriptFile(ext) {
		if warnings := checkConsoleLog(filePath); len(warnings) > 0 {
			contextMessages = append(contextMessages, "⚠️ console.log found:\n"+strings.Join(warnings, "\n"))
		}
	}

	// Run TypeScript check for .ts/.tsx files
	if ext == ".ts" || ext == ".tsx" {
		if tscErrors := runTypeScriptCheck(filePath); len(tscErrors) > 0 {
			contextMessages = append(contextMessages, "⚠️ TypeScript errors:\n"+strings.Join(tscErrors, "\n"))
		}
	}

	// Build output
	var output postToolOutput

	if len(contextMessages) > 0 {
		output = postToolOutput{
			HookSpecificOutput: &postHookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: strings.Join(contextMessages, "\n"),
			},
		}
	} else {
		output = postToolOutput{SuppressOutput: true}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}

// isJavaScriptFile checks if the extension is a JS/TS file
func isJavaScriptFile(ext string) bool {
	jsExtensions := map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".mjs": true, ".cjs": true, ".mts": true, ".cts": true,
	}
	return jsExtensions[ext]
}

// consoleLogPattern matches console.log statements
var consoleLogPattern = regexp.MustCompile(`console\.(log|debug|info|warn|error)\s*\(`)

// checkConsoleLog scans a file for console.log statements
func checkConsoleLog(filePath string) []string {
	var warnings []string

	file, err := os.Open(filePath)
	if err != nil {
		return warnings
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	foundCount := 0
	maxWarnings := 5 // Limit warnings to avoid spam

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		if consoleLogPattern.MatchString(line) {
			foundCount++
			if foundCount <= maxWarnings {
				// Truncate long lines
				displayLine := trimmed
				if len(displayLine) > 60 {
					displayLine = displayLine[:57] + "..."
				}
				warnings = append(warnings,
					strings.Repeat(" ", 2)+filepath.Base(filePath)+":"+
						string(rune('0'+lineNum/100%10))+
						string(rune('0'+lineNum/10%10))+
						string(rune('0'+lineNum%10))+
						" "+displayLine)
			}
		}
	}

	if foundCount > maxWarnings {
		warnings = append(warnings,
			strings.Repeat(" ", 2)+"... and "+(string(rune('0'+((foundCount-maxWarnings)/10)%10))+
				string(rune('0'+(foundCount-maxWarnings)%10)))+" more")
	}

	return warnings
}

// runTypeScriptCheck runs tsc --noEmit on the file and returns errors
func runTypeScriptCheck(filePath string) []string {
	var errors []string

	// Find project root with tsconfig.json
	dir := filepath.Dir(filePath)
	projectRoot := ""

	for dir != "/" && dir != "." {
		tsconfigPath := filepath.Join(dir, "tsconfig.json")
		if _, err := os.Stat(tsconfigPath); err == nil {
			projectRoot = dir
			break
		}
		dir = filepath.Dir(dir)
	}

	if projectRoot == "" {
		return errors // No tsconfig.json found
	}

	// Check if tsc is available
	if _, err := exec.LookPath("npx"); err != nil {
		return errors
	}

	// Run tsc --noEmit
	cmd := exec.Command("npx", "tsc", "--noEmit", "--pretty", "false")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Parse output for errors related to this file
		lines := strings.Split(string(output), "\n")
		relPath, _ := filepath.Rel(projectRoot, filePath)
		maxErrors := 5

		for _, line := range lines {
			if strings.Contains(line, relPath) || strings.Contains(line, filepath.Base(filePath)) {
				if len(errors) < maxErrors {
					// Truncate long error messages
					if len(line) > 100 {
						line = line[:97] + "..."
					}
					errors = append(errors, strings.Repeat(" ", 2)+line)
				}
			}
		}

		if len(errors) >= maxErrors {
			totalErrors := 0
			for _, line := range lines {
				if strings.Contains(line, relPath) || strings.Contains(line, filepath.Base(filePath)) {
					totalErrors++
				}
			}
			if totalErrors > maxErrors {
				remaining := totalErrors - maxErrors
				errors = append(errors, strings.Repeat(" ", 2)+"... and "+
					string(rune('0'+remaining/10%10))+
					string(rune('0'+remaining%10))+" more errors")
			}
		}
	}

	return errors
}

func shouldSkipFormatting(filePath string) (bool, string) {
	// Check extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if skipExtensions[ext] {
		return true, "Skipped: " + ext + " files are not formatted"
	}

	// Check for minified files
	if strings.Contains(filepath.Base(filePath), ".min.") {
		return true, "Skipped: minified file"
	}

	// Check if in skip directory
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for _, part := range parts {
		if skipDirectories[part] {
			return true, "Skipped: file in " + part + "/ directory"
		}
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true, "Skipped: file does not exist"
	}

	// Check if file is binary
	file, err := os.Open(filePath)
	if err != nil {
		return true, "Skipped: cannot read file"
	}
	defer file.Close()

	buf := make([]byte, 8192)
	n, _ := file.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, "Skipped: binary file"
		}
	}

	return false, ""
}
