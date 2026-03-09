package serve

import "time"

// Issue is the normalized issue record used across orchestration.
// Maps to Symphony SPEC Section 4.1.1.
type Issue struct {
	ID          string
	Identifier  string
	Title       string
	Description string
	Priority    *int
	State       string
	BranchName  string
	URL         string
	Labels      []string
	BlockedBy   []BlockerRef
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// BlockerRef represents a blocker issue reference.
type BlockerRef struct {
	ID         string
	Identifier string
	State      string
}

// WorkflowDefinition is the parsed WORKFLOW.md payload.
// Maps to Symphony SPEC Section 4.1.2.
type WorkflowDefinition struct {
	Config         map[string]any
	PromptTemplate string
}

// RunStatus represents the status of a run attempt.
type RunStatus string

const (
	RunStatusPreparingWorkspace RunStatus = "preparing_workspace"
	RunStatusBuildingPrompt     RunStatus = "building_prompt"
	RunStatusLaunchingAgent     RunStatus = "launching_agent"
	RunStatusStreaming           RunStatus = "streaming"
	RunStatusSucceeded          RunStatus = "succeeded"
	RunStatusFailed             RunStatus = "failed"
	RunStatusTimedOut           RunStatus = "timed_out"
	RunStatusStalled            RunStatus = "stalled"
	RunStatusCanceled           RunStatus = "canceled"
)

// LiveSession tracks state while an agent subprocess is running.
// Maps to Symphony SPEC Section 4.1.6.
type LiveSession struct {
	SessionID   string
	LastEvent   string
	LastEventAt *time.Time
	LastMessage string
	InputTokens int
	OutputTokens int
	TotalTokens  int
	TurnCount    int
}

// RetryEntry is scheduled retry state for an issue.
// Maps to Symphony SPEC Section 4.1.7.
type RetryEntry struct {
	IssueID    string
	Identifier string
	Attempt    int
	DueAt      time.Time
	Error      string
}

// RunningEntry combines all running state for one issue.
type RunningEntry struct {
	Issue         Issue
	Attempt       *int
	WorkspacePath string
	StartedAt     time.Time
	Session       LiveSession
	CancelFunc    func()
	// TerminalState is set by reconcile when the issue reaches a terminal state.
	// runWorker reads this after runner.Run returns to decide cleanup vs. retry.
	TerminalState string
}

// TokenTotals accumulates aggregate token and runtime metrics.
type TokenTotals struct {
	InputTokens   int
	OutputTokens  int
	TotalTokens   int
	SecondsRunning float64
}

// AgentEvent is an event emitted by the agent runner back to the orchestrator.
type AgentEvent struct {
	Type      AgentEventType
	IssueID   string
	Message   string
	Tokens    *TokenUsage
	Timestamp time.Time
}

// AgentEventType categorizes agent events.
type AgentEventType string

const (
	AgentEventStarted   AgentEventType = "session_started"
	AgentEventCompleted AgentEventType = "turn_completed"
	AgentEventFailed    AgentEventType = "turn_failed"
	AgentEventStalled   AgentEventType = "stalled"
	AgentEventMessage   AgentEventType = "notification"
)

// TokenUsage holds token counts from an agent session.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}
