// Package team provides core data types and logic for jikime team orchestration.
// It supports multi-agent parallel workflows, task management, messaging,
// and cost tracking across distributed Claude Code agents.
package team

import "time"

// --- Team Configuration ---

// TeamConfig holds all configuration for a named team workspace.
// Persisted to ~/.jikime/teams/<team-name>/config.json.
type TeamConfig struct {
	// Name is the unique identifier for this team.
	Name string `json:"name"`

	// Template is the template name used to create this team (e.g. "leader-worker").
	Template string `json:"template,omitempty"`

	// BaseDir is the absolute path to the team workspace directory.
	// Default: ~/.jikime/teams/<name>/
	BaseDir string `json:"base_dir"`

	// MaxAgents is the maximum number of concurrent agents allowed.
	// 0 means unlimited.
	MaxAgents int `json:"max_agents,omitempty"`

	// Budget is the token budget limit across all agents in this team.
	// 0 means no limit.
	Budget int `json:"budget,omitempty"`

	// TimeoutSeconds is the overall team execution timeout in seconds.
	// 0 means no timeout.
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`

	// CreatedAt is when this team was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this team configuration was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	// Metadata holds arbitrary key-value pairs for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// --- Task ---

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	// TaskStatusPending means the task is waiting to be claimed.
	TaskStatusPending TaskStatus = "pending"

	// TaskStatusInProgress means an agent has claimed and is working on the task.
	TaskStatusInProgress TaskStatus = "in_progress"

	// TaskStatusDone means the task has been completed successfully.
	TaskStatusDone TaskStatus = "done"

	// TaskStatusBlocked means the task is waiting on one or more dependencies.
	TaskStatusBlocked TaskStatus = "blocked"

	// TaskStatusFailed means the task encountered an unrecoverable error.
	TaskStatusFailed TaskStatus = "failed"
)

// Task represents a unit of work in the team task queue.
// Persisted to ~/.jikime/teams/<team-name>/tasks/<id>.json.
type Task struct {
	// ID is the unique identifier for this task (UUID v4).
	ID string `json:"id"`

	// TeamName is the name of the team this task belongs to.
	TeamName string `json:"team_name"`

	// Title is a short human-readable description of the task.
	Title string `json:"title"`

	// Description provides additional context about what needs to be done.
	Description string `json:"description,omitempty"`

	// DoD (Definition of Done) describes the measurable acceptance criteria.
	DoD string `json:"dod,omitempty"`

	// Status is the current lifecycle state of this task.
	Status TaskStatus `json:"status"`

	// Owner is the agent ID pre-assigned to this task at creation time.
	// Empty means any worker can claim it (first-come, first-served).
	Owner string `json:"owner,omitempty"`

	// AgentID is the identifier of the agent that has claimed this task.
	// Empty when the task is pending.
	AgentID string `json:"agent_id,omitempty"`

	// DependsOn lists the IDs of tasks that must be done before this task can start.
	DependsOn []string `json:"depends_on,omitempty"`

	// Priority is a numeric priority value (higher = more important).
	// 0 is the default priority.
	Priority int `json:"priority,omitempty"`

	// Tags are arbitrary labels for filtering and grouping tasks.
	Tags []string `json:"tags,omitempty"`

	// Result holds the output or summary produced when the task completes.
	Result string `json:"result,omitempty"`

	// ErrorMsg holds the error message if the task failed.
	ErrorMsg string `json:"error_msg,omitempty"`

	// CreatedAt is when this task was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this task was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	// ClaimedAt is when an agent first claimed this task.
	ClaimedAt *time.Time `json:"claimed_at,omitempty"`

	// CompletedAt is when this task transitioned to done or failed.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// --- Message ---

// MessageKind classifies the purpose of a message.
type MessageKind string

const (
	// MessageKindDirect is a point-to-point message to a specific agent.
	MessageKindDirect MessageKind = "direct"

	// MessageKindBroadcast is sent to all agents in the team.
	MessageKindBroadcast MessageKind = "broadcast"

	// MessageKindSystem is sent by the orchestrator or CLI for control purposes.
	MessageKindSystem MessageKind = "system"
)

// Message represents a unit of communication between agents or the orchestrator.
// Persisted to ~/.jikime/teams/<team-name>/inbox/<recipient>/<id>.json.
type Message struct {
	// ID is the unique identifier for this message (UUID v4).
	ID string `json:"id"`

	// TeamName is the team this message belongs to.
	TeamName string `json:"team_name"`

	// Kind classifies the message (direct, broadcast, system).
	Kind MessageKind `json:"kind"`

	// From is the sender agent ID, or "orchestrator" for system messages.
	From string `json:"from"`

	// To is the recipient agent ID. Empty for broadcast messages.
	To string `json:"to,omitempty"`

	// Subject is an optional short summary of the message content.
	Subject string `json:"subject,omitempty"`

	// Body is the main content of the message.
	Body string `json:"body"`

	// ReplyTo is the ID of a previous message this is in response to.
	ReplyTo string `json:"reply_to,omitempty"`

	// Read indicates whether the recipient has read this message.
	Read bool `json:"read"`

	// SentAt is when this message was created and queued.
	SentAt time.Time `json:"sent_at"`

	// ReadAt is when the recipient marked this message as read.
	ReadAt *time.Time `json:"read_at,omitempty"`
}

// --- Agent Registry ---

// AgentStatus represents the liveness state of a registered agent.
type AgentStatus string

const (
	// AgentStatusActive means the agent is running and responsive.
	AgentStatusActive AgentStatus = "active"

	// AgentStatusIdle means the agent is running but has no current task.
	AgentStatusIdle AgentStatus = "idle"

	// AgentStatusOffline means the agent process is no longer alive.
	AgentStatusOffline AgentStatus = "offline"

	// AgentStatusShuttingDown means the agent has received a shutdown request.
	AgentStatusShuttingDown AgentStatus = "shutting_down"
)

// AgentInfo holds registration and liveness information for a single agent.
// Persisted to ~/.jikime/teams/<team-name>/registry/<agent-id>.json.
type AgentInfo struct {
	// ID is the unique identifier for this agent (e.g. "worker-1").
	ID string `json:"id"`

	// TeamName is the team this agent belongs to.
	TeamName string `json:"team_name"`

	// Role describes the agent's function (e.g. "leader", "worker", "reviewer").
	Role string `json:"role,omitempty"`

	// Status is the current liveness state of this agent.
	Status AgentStatus `json:"status"`

	// PID is the OS process ID of the agent, used for liveness checks.
	// 0 if the agent was spawned via tmux and PID is unavailable.
	PID int `json:"pid,omitempty"`

	// TmuxSession is the tmux session name if the agent runs in tmux.
	TmuxSession string `json:"tmux_session,omitempty"`

	// CurrentTaskID is the ID of the task currently being worked on, if any.
	CurrentTaskID string `json:"current_task_id,omitempty"`

	// LastHeartbeat is the last time the agent reported it was alive.
	LastHeartbeat time.Time `json:"last_heartbeat"`

	// JoinedAt is when this agent first registered with the team.
	JoinedAt time.Time `json:"joined_at"`

	// Metadata holds arbitrary key-value pairs for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// --- Cost Event ---

// CostEvent records a single token usage event from an agent tool call.
// Persisted to ~/.jikime/teams/<team-name>/costs/<agent-id>-<timestamp>.json.
type CostEvent struct {
	// ID is a unique identifier for this cost event (UUID v4).
	ID string `json:"id"`

	// TeamName is the team this event belongs to.
	TeamName string `json:"team_name"`

	// AgentID is the agent that incurred this cost.
	AgentID string `json:"agent_id"`

	// TaskID is the task being worked on when this cost was incurred, if known.
	TaskID string `json:"task_id,omitempty"`

	// ToolName is the name of the tool call that produced this usage.
	ToolName string `json:"tool_name,omitempty"`

	// InputTokens is the number of input (prompt) tokens used.
	InputTokens int `json:"input_tokens"`

	// OutputTokens is the number of output (completion) tokens used.
	OutputTokens int `json:"output_tokens"`

	// TotalTokens is InputTokens + OutputTokens, stored for convenience.
	TotalTokens int `json:"total_tokens"`

	// Model is the model name used for this tool call (e.g. "claude-sonnet-4-6").
	Model string `json:"model,omitempty"`

	// OccurredAt is when this cost event was recorded.
	OccurredAt time.Time `json:"occurred_at"`
}

// CostSummary aggregates cost events for reporting purposes.
type CostSummary struct {
	// TeamName is the team these costs belong to.
	TeamName string `json:"team_name"`

	// AgentID is the agent being summarized. Empty means team-wide summary.
	AgentID string `json:"agent_id,omitempty"`

	// TotalInputTokens is the sum of all input tokens.
	TotalInputTokens int `json:"total_input_tokens"`

	// TotalOutputTokens is the sum of all output tokens.
	TotalOutputTokens int `json:"total_output_tokens"`

	// TotalTokens is TotalInputTokens + TotalOutputTokens.
	TotalTokens int `json:"total_tokens"`

	// EventCount is the number of cost events included in this summary.
	EventCount int `json:"event_count"`

	// Budget is the configured budget limit. 0 means no limit.
	Budget int `json:"budget,omitempty"`

	// BudgetUsedPercent is TotalTokens / Budget * 100. 0 if no budget set.
	BudgetUsedPercent float64 `json:"budget_used_percent,omitempty"`
}

// --- Session ---

// Session captures a snapshot of team state for persistence and restoration.
// Persisted to ~/.jikime/sessions/<session-id>.json.
type Session struct {
	// ID is the unique identifier for this session (UUID v4).
	ID string `json:"id"`

	// TeamName is the team this session belongs to.
	TeamName string `json:"team_name"`

	// Description is an optional human-readable note about this session.
	Description string `json:"description,omitempty"`

	// Tasks is the snapshot of all tasks at save time.
	Tasks []Task `json:"tasks"`

	// Agents is the snapshot of all registered agents at save time.
	Agents []AgentInfo `json:"agents"`

	// SavedAt is when this session snapshot was taken.
	SavedAt time.Time `json:"saved_at"`

	// RestoredAt is when this session was last restored, if ever.
	RestoredAt *time.Time `json:"restored_at,omitempty"`
}

// --- Plan ---

// PlanStatus represents the approval lifecycle of a plan.
type PlanStatus string

const (
	// PlanStatusPending means the plan has been submitted and awaits review.
	PlanStatusPending PlanStatus = "pending"

	// PlanStatusApproved means the leader has approved this plan.
	PlanStatusApproved PlanStatus = "approved"

	// PlanStatusRejected means the leader has rejected this plan.
	PlanStatusRejected PlanStatus = "rejected"
)

// Plan represents a proposed set of tasks submitted by a worker agent for leader approval.
// Persisted to ~/.jikime/plans/<id>.json.
type Plan struct {
	// ID is the unique identifier for this plan (UUID v4).
	ID string `json:"id"`

	// TeamName is the team this plan belongs to.
	TeamName string `json:"team_name"`

	// SubmittedBy is the agent ID that submitted this plan.
	SubmittedBy string `json:"submitted_by"`

	// Title is a short human-readable summary of what this plan does.
	Title string `json:"title"`

	// Body is the full description of the proposed approach.
	Body string `json:"body"`

	// Tasks is the list of tasks proposed by this plan.
	Tasks []Task `json:"tasks,omitempty"`

	// Status is the current approval state of this plan.
	Status PlanStatus `json:"status"`

	// ReviewedBy is the agent ID (usually "leader") that reviewed this plan.
	ReviewedBy string `json:"reviewed_by,omitempty"`

	// RejectionReason is the reason provided when the plan was rejected.
	RejectionReason string `json:"rejection_reason,omitempty"`

	// SubmittedAt is when this plan was submitted.
	SubmittedAt time.Time `json:"submitted_at"`

	// ReviewedAt is when the plan was approved or rejected.
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

// --- Template ---

// TemplateAgentDef defines a single agent role within a team template.
type TemplateAgentDef struct {
	// ID is the agent identifier within the template (e.g. "leader", "worker-1").
	ID string `json:"id" yaml:"id"`

	// Role is the functional role label (e.g. "leader", "worker", "reviewer").
	Role string `json:"role" yaml:"role"`

	// Description is a short human-readable summary of this agent's purpose.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Task is the full task/goal prompt injected into this agent at spawn time.
	// Supports placeholders: {{goal}}, {{team_name}}, {{agent_id}}, {{leader_id}}.
	// If empty, a default prompt is generated based on Role.
	Task string `json:"task,omitempty" yaml:"task,omitempty"`

	// SystemPromptFile is an optional path (relative to template dir) to a
	// CLAUDE.md or system prompt file to inject for this agent.
	SystemPromptFile string `json:"system_prompt_file,omitempty" yaml:"system_prompt_file,omitempty"`

	// AutoSpawn indicates whether this agent should be spawned automatically
	// when the team is launched.
	AutoSpawn bool `json:"auto_spawn" yaml:"auto_spawn"`

	// Metadata holds arbitrary extensible key-value configuration.
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TemplateTaskDef defines a task that is pre-created when the team launches.
type TemplateTaskDef struct {
	// Subject is the task title.
	Subject string `json:"subject" yaml:"subject"`

	// Description provides additional context.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// DoD (Definition of Done) describes the acceptance criteria.
	DoD string `json:"dod,omitempty" yaml:"dod,omitempty"`

	// Owner is the agent ID pre-assigned to this task.
	// Empty means any worker can claim it.
	Owner string `json:"owner,omitempty" yaml:"owner,omitempty"`
}

// TemplateDef describes a reusable team configuration blueprint.
// Loaded from YAML files in the templates/teams/ directory.
type TemplateDef struct {
	// Name is the unique template identifier (e.g. "leader-worker").
	Name string `json:"name" yaml:"name"`

	// Description is a human-readable explanation of this template's purpose.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Version is the semantic version of this template definition.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Agents is the list of agent roles defined by this template.
	Agents []TemplateAgentDef `json:"agents" yaml:"agents"`

	// Tasks is a list of tasks pre-created when the team launches.
	// Useful for structured workflows where tasks are known upfront.
	Tasks []TemplateTaskDef `json:"tasks,omitempty" yaml:"tasks,omitempty"`

	// DefaultBudget is the default token budget for teams created from this template.
	// 0 means no limit.
	DefaultBudget int `json:"default_budget,omitempty" yaml:"default_budget,omitempty"`

	// DefaultMaxAgents is the default maximum concurrent agents.
	// 0 means unlimited.
	DefaultMaxAgents int `json:"default_max_agents,omitempty" yaml:"default_max_agents,omitempty"`

	// DefaultTimeoutSeconds is the default execution timeout.
	// 0 means no timeout.
	DefaultTimeoutSeconds int `json:"default_timeout_seconds,omitempty" yaml:"default_timeout_seconds,omitempty"`

	// Metadata holds arbitrary extensible key-value configuration.
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
