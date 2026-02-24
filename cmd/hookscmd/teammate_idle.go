package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// TeammateIdleCmd represents the teammate-idle hook command
var TeammateIdleCmd = &cobra.Command{
	Use:   "teammate-idle",
	Short: "Validate quality gates before accepting teammate idle",
	Long: `TeammateIdle hook that checks quality gates before allowing a teammate to go idle.

Features:
- Check LSP diagnostics baseline (errors/warnings) against quality gate thresholds
- Check test coverage against configured threshold
- Exit code 0: Accept idle (quality gates pass)
- Exit code 2: Keep working (quality gates fail)

Only active in Agent Teams mode (CLAUDE_HOOK_EVENT_TEAM_NAME is set).
Based on JikiME-ADK hook patterns.`,
	RunE: runTeammateIdle,
}

// --- Types (inlined from moai-adk internal/lsp/hook/types.go) ---

type severityCounts struct {
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
	Information int `json:"information"`
	Hints       int `json:"hints"`
}

type qualityGate struct {
	MaxErrors      int  `json:"maxErrors"`
	MaxWarnings    int  `json:"maxWarnings"`
	BlockOnError   bool `json:"blockOnError"`
	BlockOnWarning bool `json:"blockOnWarning"`
}

type coverageData struct {
	Overall float64 `json:"overall"`
}

type teammateIdleOutput struct {
	Continue       bool   `json:"continue"`
	SystemMessage  string `json:"systemMessage,omitempty"`
	SuppressOutput bool   `json:"suppressOutput,omitempty"`
}

// --- Main handler ---

func runTeammateIdle(cmd *cobra.Command, args []string) error {
	// Only enforce in team mode
	teamName := os.Getenv("CLAUDE_HOOK_EVENT_TEAM_NAME")
	if teamName == "" {
		return outputTeammateIdle(teammateIdleOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	teammateName := os.Getenv("CLAUDE_HOOK_EVENT_TEAMMATE_NAME")

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		// Cannot determine project root - accept idle
		return outputTeammateIdle(teammateIdleOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	// Load quality gate config
	gate := loadQualityGateConfig(projectRoot)
	if !gate.BlockOnError {
		// Quality gates disabled - accept idle
		return outputTeammateIdle(teammateIdleOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	// Check LSP diagnostics baseline
	counts := loadBaselineCounts(projectRoot)
	if shouldBlock(counts, gate) {
		msg := fmt.Sprintf(
			"[Hook] TeammateIdle blocked for %s: LSP diagnostics exceed quality gate "+
				"(errors: %d/%d, warnings: %d/%d). Fix issues before going idle.",
			teammateName,
			counts.Errors, gate.MaxErrors,
			counts.Warnings, gate.MaxWarnings,
		)
		// Exit code 2 = keep working
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(2)
	}

	// Check test coverage
	coverage := loadCoverageData(projectRoot)
	threshold := loadCoverageThreshold(projectRoot)
	if coverage.Overall > 0 && coverage.Overall < threshold {
		msg := fmt.Sprintf(
			"[Hook] TeammateIdle blocked for %s: Test coverage %.1f%% is below threshold %.1f%%. "+
				"Improve coverage before going idle.",
			teammateName,
			coverage.Overall,
			threshold,
		)
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(2)
	}

	// All quality gates pass - accept idle
	return outputTeammateIdle(teammateIdleOutput{
		Continue:       true,
		SuppressOutput: true,
	})
}

// --- Helper functions ---

func outputTeammateIdle(out teammateIdleOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(out)
}

func loadQualityGateConfig(projectRoot string) qualityGate {
	gate := qualityGate{
		MaxErrors:      0,
		MaxWarnings:    10,
		BlockOnError:   true,
		BlockOnWarning: false,
	}

	configPath := filepath.Join(projectRoot, ".jikime", "config", "quality.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return gate
	}

	// Simple YAML parsing for quality gate settings
	// Look for lsp_quality_gates section
	parseQualityGateYAML(string(data), &gate)
	return gate
}

func parseQualityGateYAML(content string, gate *qualityGate) {
	// Simple line-by-line YAML parsing for the fields we need
	// This avoids importing a full YAML library
	lines := splitLines(content)
	inLSPSection := false
	inRunSection := false

	for _, line := range lines {
		trimmed := trimString(line)

		if trimmed == "lsp_quality_gates:" {
			inLSPSection = true
			continue
		}
		if inLSPSection && trimmed == "run:" {
			inRunSection = true
			continue
		}

		// Reset section tracking on non-indented lines
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			if trimmed != "constitution:" {
				inLSPSection = false
				inRunSection = false
			}
			continue
		}

		if inLSPSection && !inRunSection {
			if key, val, ok := parseYAMLKeyValue(trimmed); ok {
				switch key {
				case "enabled":
					if val == "false" {
						gate.BlockOnError = false
						gate.BlockOnWarning = false
					}
				}
			}
		}

		if inRunSection {
			if key, val, ok := parseYAMLKeyValue(trimmed); ok {
				switch key {
				case "max_errors":
					gate.MaxErrors = parseIntDefault(val, 0)
				case "max_warnings":
					// not used for blocking in run phase typically
				case "allow_regression":
					if val == "true" {
						gate.BlockOnError = false
					}
				}
			}
		}
	}
}

func loadBaselineCounts(projectRoot string) severityCounts {
	var counts severityCounts

	baselinePath := filepath.Join(projectRoot, ".jikime", "memory", "diagnostics-baseline.json")
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return counts
	}

	_ = json.Unmarshal(data, &counts)
	return counts
}

func shouldBlock(counts severityCounts, gate qualityGate) bool {
	if gate.BlockOnError && counts.Errors > gate.MaxErrors {
		return true
	}
	if gate.BlockOnWarning && counts.Warnings > gate.MaxWarnings {
		return true
	}
	return false
}

func loadCoverageData(projectRoot string) coverageData {
	var cov coverageData

	coveragePath := filepath.Join(projectRoot, ".jikime", "memory", "coverage.json")
	data, err := os.ReadFile(coveragePath)
	if err != nil {
		return cov
	}

	_ = json.Unmarshal(data, &cov)
	return cov
}

func loadCoverageThreshold(projectRoot string) float64 {
	defaultThreshold := 85.0

	configPath := filepath.Join(projectRoot, ".jikime", "config", "quality.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultThreshold
	}

	// Parse test_coverage_target from quality.yaml
	lines := splitLines(string(data))
	for _, line := range lines {
		trimmed := trimString(line)
		if key, val, ok := parseYAMLKeyValue(trimmed); ok {
			if key == "test_coverage_target" {
				if v := parseFloatDefault(val, defaultThreshold); v > 0 {
					return v
				}
			}
		}
	}

	return defaultThreshold
}

// --- Utility functions ---

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimString(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func parseYAMLKeyValue(line string) (string, string, bool) {
	// Skip comments
	if len(line) == 0 || line[0] == '#' {
		return "", "", false
	}

	colonIdx := -1
	for i := 0; i < len(line); i++ {
		if line[i] == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return "", "", false
	}

	key := trimString(line[:colonIdx])
	val := ""
	if colonIdx+1 < len(line) {
		val = trimString(line[colonIdx+1:])
	}

	// Remove inline comments
	for i := 0; i < len(val); i++ {
		if val[i] == '#' {
			val = trimString(val[:i])
			break
		}
	}

	return key, val, key != ""
}

func parseIntDefault(s string, def int) int {
	n := 0
	negative := false
	i := 0

	if len(s) == 0 {
		return def
	}
	if s[0] == '-' {
		negative = true
		i = 1
	}

	parsed := false
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int(s[i]-'0')
			parsed = true
		} else {
			break
		}
	}

	if !parsed {
		return def
	}
	if negative {
		n = -n
	}
	return n
}

func parseFloatDefault(s string, def float64) float64 {
	if len(s) == 0 {
		return def
	}

	// Simple float parser for positive numbers (e.g., "85.0", "100")
	whole := 0
	frac := 0.0
	fracDiv := 1.0
	inFrac := false
	parsed := false

	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			inFrac = true
			continue
		}
		if s[i] >= '0' && s[i] <= '9' {
			parsed = true
			if inFrac {
				fracDiv *= 10
				frac += float64(s[i]-'0') / fracDiv
			} else {
				whole = whole*10 + int(s[i]-'0')
			}
		} else {
			break
		}
	}

	if !parsed {
		return def
	}
	return float64(whole) + frac
}
