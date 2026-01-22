package hookscmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"jikime-adk/internal/hooks"
)

// LoopState represents the enhanced state for Ralph Loop
type LoopState struct {
	// Basic info
	Active    bool      `json:"active"`
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Iteration info
	Iteration     int `json:"iteration"`
	MaxIterations int `json:"max_iterations"`

	// Task info
	TaskDescription string   `json:"task_description"`
	TargetFiles     []string `json:"target_files,omitempty"`

	// Completion criteria
	Criteria CompletionCriteria `json:"completion_criteria"`

	// Diagnostic history
	Snapshots []DiagnosticSnapshot `json:"snapshots"`

	// Final result
	CompletionReason string `json:"completion_reason,omitempty"`
	FinalStatus      string `json:"final_status,omitempty"` // COMPLETE, STOPPED, CANCELLED
}

// CompletionCriteria defines conditions for loop completion
type CompletionCriteria struct {
	ZeroErrors      bool `json:"zero_errors"`       // Require zero errors
	ZeroWarnings    bool `json:"zero_warnings"`     // Require zero warnings
	ZeroSecurity    bool `json:"zero_security"`     // Require zero security issues
	TestsPass       bool `json:"tests_pass"`        // Require tests to pass
	StagnationLimit int  `json:"stagnation_limit"`  // Max iterations without improvement
}

// DiagnosticSnapshot captures state at a point in time
type DiagnosticSnapshot struct {
	Iteration int       `json:"iteration"`
	Timestamp time.Time `json:"timestamp"`

	// LSP diagnostics
	ErrorCount   int `json:"error_count"`
	WarningCount int `json:"warning_count"`
	InfoCount    int `json:"info_count"`

	// AST-grep results
	SecurityIssues int `json:"security_issues"`

	// Test results
	TestsPassed bool `json:"tests_passed"`
	TestsRun    int  `json:"tests_run"`
	TestsFailed int  `json:"tests_failed"`

	// File details
	FileDetails []FileDetail `json:"file_details,omitempty"`
}

// FileDetail contains per-file diagnostic info
type FileDetail struct {
	Path         string `json:"path"`
	ErrorCount   int    `json:"error_count"`
	WarningCount int    `json:"warning_count"`
}

// CompletionResult represents the result of completion check
type CompletionResult struct {
	Complete        bool    `json:"complete"`
	Reason          string  `json:"reason"`
	ImprovementRate float64 `json:"improvement_rate"`
	Guidance        string  `json:"guidance"`
}

// LoopStateFile path
const loopStateFileName = ".jikime_loop_state.json"

// LoadEnhancedLoopState loads the enhanced loop state
func LoadEnhancedLoopState() *LoopState {
	statePath := GetLoopStatePath()
	data, err := os.ReadFile(statePath)
	if err != nil {
		return &LoopState{
			MaxIterations: 10,
			Criteria: CompletionCriteria{
				ZeroErrors:      true,
				StagnationLimit: 3,
			},
		}
	}

	var state LoopState
	if err := json.Unmarshal(data, &state); err != nil {
		return &LoopState{
			MaxIterations: 10,
			Criteria: CompletionCriteria{
				ZeroErrors:      true,
				StagnationLimit: 3,
			},
		}
	}

	// Ensure defaults
	if state.MaxIterations == 0 {
		state.MaxIterations = 10
	}
	if state.Criteria.StagnationLimit == 0 {
		state.Criteria.StagnationLimit = 3
	}

	return &state
}

// SaveEnhancedLoopState saves the enhanced loop state
func SaveEnhancedLoopState(state *LoopState) error {
	state.UpdatedAt = time.Now()

	statePath := GetLoopStatePath()
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// ClearEnhancedLoopState removes the loop state file
func ClearEnhancedLoopState() error {
	statePath := GetLoopStatePath()
	return os.Remove(statePath)
}

// GetLoopStatePath returns the path to the loop state file
func GetLoopStatePath() string {
	jikimeDir, err := hooks.FindJikimeDir()
	if err != nil {
		return loopStateFileName
	}
	return filepath.Join(jikimeDir, "cache", loopStateFileName)
}

// AddSnapshot adds a diagnostic snapshot to the state
func (s *LoopState) AddSnapshot(snapshot DiagnosticSnapshot) {
	snapshot.Iteration = s.Iteration
	snapshot.Timestamp = time.Now()
	s.Snapshots = append(s.Snapshots, snapshot)
}

// GetLatestSnapshot returns the most recent snapshot
func (s *LoopState) GetLatestSnapshot() *DiagnosticSnapshot {
	if len(s.Snapshots) == 0 {
		return nil
	}
	return &s.Snapshots[len(s.Snapshots)-1]
}

// GetInitialSnapshot returns the first snapshot
func (s *LoopState) GetInitialSnapshot() *DiagnosticSnapshot {
	if len(s.Snapshots) == 0 {
		return nil
	}
	return &s.Snapshots[0]
}

// CalculateImprovementRate calculates the improvement from initial to current
func (s *LoopState) CalculateImprovementRate() float64 {
	if len(s.Snapshots) < 2 {
		return 0.0
	}

	initial := s.Snapshots[0]
	current := s.Snapshots[len(s.Snapshots)-1]

	initialIssues := float64(initial.ErrorCount + initial.WarningCount + initial.SecurityIssues)
	currentIssues := float64(current.ErrorCount + current.WarningCount + current.SecurityIssues)

	if initialIssues == 0 {
		return 1.0
	}

	rate := (initialIssues - currentIssues) / initialIssues
	if rate < 0 {
		return 0.0
	}
	if rate > 1 {
		return 1.0
	}
	return rate
}

// IsStagnant checks if the loop has stagnated (no improvement)
func (s *LoopState) IsStagnant() bool {
	limit := s.Criteria.StagnationLimit
	if limit <= 0 || len(s.Snapshots) < limit {
		return false
	}

	recent := s.Snapshots[len(s.Snapshots)-limit:]
	firstIssues := recent[0].ErrorCount + recent[0].WarningCount + recent[0].SecurityIssues

	for _, snap := range recent[1:] {
		currentIssues := snap.ErrorCount + snap.WarningCount + snap.SecurityIssues
		if currentIssues < firstIssues {
			return false // There was improvement
		}
	}

	return true // No improvement in recent iterations
}

// EvaluateCompletion checks if completion criteria are met
func (s *LoopState) EvaluateCompletion() CompletionResult {
	latest := s.GetLatestSnapshot()
	if latest == nil {
		return CompletionResult{
			Complete: false,
			Reason:   "No diagnostic data",
			Guidance: "Run diagnostics first",
		}
	}

	// Check zero errors
	if s.Criteria.ZeroErrors && latest.ErrorCount > 0 {
		return CompletionResult{
			Complete:        false,
			Reason:          "Errors remain",
			ImprovementRate: s.CalculateImprovementRate(),
			Guidance:        formatGuidance("Fix %d remaining error(s)", latest.ErrorCount),
		}
	}

	// Check zero warnings
	if s.Criteria.ZeroWarnings && latest.WarningCount > 0 {
		return CompletionResult{
			Complete:        false,
			Reason:          "Warnings remain",
			ImprovementRate: s.CalculateImprovementRate(),
			Guidance:        formatGuidance("Address %d warning(s)", latest.WarningCount),
		}
	}

	// Check zero security issues
	if s.Criteria.ZeroSecurity && latest.SecurityIssues > 0 {
		return CompletionResult{
			Complete:        false,
			Reason:          "Security issues remain",
			ImprovementRate: s.CalculateImprovementRate(),
			Guidance:        formatGuidance("Fix %d security issue(s)", latest.SecurityIssues),
		}
	}

	// Check tests pass
	if s.Criteria.TestsPass && !latest.TestsPassed {
		return CompletionResult{
			Complete:        false,
			Reason:          "Tests failing",
			ImprovementRate: s.CalculateImprovementRate(),
			Guidance:        formatGuidance("Fix %d failing test(s)", latest.TestsFailed),
		}
	}

	// All conditions met
	return CompletionResult{
		Complete:        true,
		Reason:          "All conditions satisfied",
		ImprovementRate: s.CalculateImprovementRate(),
	}
}

func formatGuidance(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return formatLoopString(format, args[0].(int))
}

func formatLoopString(format string, val int) string {
	// Simple format without fmt package to keep it lightweight
	result := ""
	inPercent := false
	for _, c := range format {
		if c == '%' {
			inPercent = true
			continue
		}
		if inPercent {
			if c == 'd' {
				result += strconv.Itoa(val)
			}
			inPercent = false
			continue
		}
		result += string(c)
	}
	return result
}

// AppendDiagnosticEntry is used by post-tool hooks to add diagnostic data
type DiagnosticEntry struct {
	Source       string `json:"source"` // lsp, ast-grep, test
	FilePath     string `json:"file_path"`
	ErrorCount   int    `json:"error_count"`
	WarningCount int    `json:"warning_count"`
	SecurityHits int    `json:"security_hits"`
}

// AppendDiagnosticToSnapshot adds diagnostic data to the current snapshot
func AppendDiagnosticToSnapshot(entry DiagnosticEntry) {
	state := LoadEnhancedLoopState()
	if !state.Active {
		return
	}

	// Get or create current snapshot
	var snapshot *DiagnosticSnapshot
	if len(state.Snapshots) > 0 && state.Snapshots[len(state.Snapshots)-1].Iteration == state.Iteration {
		snapshot = &state.Snapshots[len(state.Snapshots)-1]
	} else {
		newSnapshot := DiagnosticSnapshot{
			Iteration: state.Iteration,
			Timestamp: time.Now(),
		}
		state.Snapshots = append(state.Snapshots, newSnapshot)
		snapshot = &state.Snapshots[len(state.Snapshots)-1]
	}

	// Update snapshot based on source
	switch entry.Source {
	case "lsp":
		snapshot.ErrorCount += entry.ErrorCount
		snapshot.WarningCount += entry.WarningCount
		if entry.FilePath != "" {
			snapshot.FileDetails = append(snapshot.FileDetails, FileDetail{
				Path:         entry.FilePath,
				ErrorCount:   entry.ErrorCount,
				WarningCount: entry.WarningCount,
			})
		}
	case "ast-grep":
		snapshot.SecurityIssues += entry.SecurityHits
	case "test":
		snapshot.TestsRun = entry.ErrorCount + entry.WarningCount // Reuse fields
		snapshot.TestsFailed = entry.ErrorCount
		snapshot.TestsPassed = entry.ErrorCount == 0
	}

	SaveEnhancedLoopState(state)
}

// IsLoopActive checks if a loop is currently active
func IsLoopActive() bool {
	state := LoadEnhancedLoopState()
	return state.Active
}
