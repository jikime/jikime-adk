// Package orchestrator implements the core scheduling and state machine.
// Implements Symphony SPEC Sections 7, 8, and 16.
package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"jikime-adk/internal/serve"
	"jikime-adk/internal/serve/agent"
	"jikime-adk/internal/serve/tracker"
	"jikime-adk/internal/serve/workspace"
	"jikime-adk/internal/serve/workflow"
)

// Orchestrator owns the poll loop and in-memory runtime state.
// Single authority for all state mutations.
type Orchestrator struct {
	mu sync.Mutex

	// Runtime state
	running  map[string]*serve.RunningEntry // issueID → entry
	claimed  map[string]bool                // issueID → reserved
	retries  map[string]*serve.RetryEntry   // issueID → retry
	totals   serve.TokenTotals

	// Dependencies
	cfg       *workflow.Config
	tracker   tracker.Client
	workspace *workspace.Manager
	runner    *agent.Runner
	logger    *slog.Logger

	// Config snapshot (updated on WORKFLOW.md reload)
	pollIntervalMS      int
	maxConcurrentAgents int
	stallTimeoutMS      int
	terminalStates      map[string]bool
	activeStates        map[string]bool
}

// New creates a new Orchestrator.
func New(
	cfg *workflow.Config,
	t tracker.Client,
	ws *workspace.Manager,
	r *agent.Runner,
	logger *slog.Logger,
) *Orchestrator {
	o := &Orchestrator{
		running:  make(map[string]*serve.RunningEntry),
		claimed:  make(map[string]bool),
		retries:  make(map[string]*serve.RetryEntry),
		tracker:  t,
		workspace: ws,
		runner:   r,
		logger:   logger,
	}
	o.applyConfig(cfg)
	return o
}

// ApplyConfig updates runtime config from a reloaded workflow definition.
// Called on WORKFLOW.md change — does not affect in-flight sessions.
func (o *Orchestrator) ApplyConfig(cfg *workflow.Config) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.applyConfig(cfg)
}

func (o *Orchestrator) applyConfig(cfg *workflow.Config) {
	o.cfg = cfg
	o.pollIntervalMS = cfg.PollIntervalMS()
	o.maxConcurrentAgents = cfg.MaxConcurrentAgents()
	o.stallTimeoutMS = cfg.StallTimeoutMS()

	o.terminalStates = make(map[string]bool)
	for _, s := range cfg.TrackerTerminalStates() {
		o.terminalStates[normalize(s)] = true
	}
	o.activeStates = make(map[string]bool)
	for _, s := range cfg.TrackerActiveStates() {
		o.activeStates[normalize(s)] = true
	}
}

// StartupCleanup removes workspaces for issues already in terminal states.
func (o *Orchestrator) StartupCleanup(ctx context.Context) {
	o.mu.Lock()
	termStates := o.cfg.TrackerTerminalStates()
	o.mu.Unlock()

	issues, err := o.tracker.FetchIssuesByStates(termStates)
	if err != nil {
		o.logger.Warn("startup cleanup: fetch failed (continuing)", "error", err)
		return
	}
	for _, issue := range issues {
		o.workspace.CleanupForIssue(issue.Identifier)
	}
}

// Run starts the poll loop and blocks until ctx is cancelled.
func (o *Orchestrator) Run(ctx context.Context) {
	o.logger.Info("orchestrator started",
		"poll_interval_ms", o.pollIntervalMS,
		"max_concurrent_agents", o.maxConcurrentAgents,
	)

	// Immediate first tick
	o.tick(ctx)

	for {
		o.mu.Lock()
		interval := time.Duration(o.pollIntervalMS) * time.Millisecond
		o.mu.Unlock()

		select {
		case <-ctx.Done():
			o.logger.Info("orchestrator shutting down")
			return
		case <-time.After(interval):
			o.tick(ctx)
		}
	}
}

// Snapshot returns a read-only view of current state for the HTTP API.
func (o *Orchestrator) Snapshot() Snapshot {
	o.mu.Lock()
	defer o.mu.Unlock()

	running := make([]RunningRow, 0, len(o.running))
	for _, e := range o.running {
		running = append(running, RunningRow{
			IssueID:         e.Issue.ID,
			IssueIdentifier: e.Issue.Identifier,
			State:           e.Issue.State,
			SessionID:       e.Session.SessionID,
			TurnCount:       e.Session.TurnCount,
			LastEvent:       e.Session.LastEvent,
			LastMessage:     e.Session.LastMessage,
			StartedAt:       e.StartedAt,
			LastEventAt:     e.Session.LastEventAt,
			Tokens: serve.TokenUsage{
				InputTokens:  e.Session.InputTokens,
				OutputTokens: e.Session.OutputTokens,
				TotalTokens:  e.Session.TotalTokens,
			},
		})
	}

	retrying := make([]RetryRow, 0, len(o.retries))
	for _, r := range o.retries {
		retrying = append(retrying, RetryRow{
			IssueID:    r.IssueID,
			Identifier: r.Identifier,
			Attempt:    r.Attempt,
			DueAt:      r.DueAt,
			Error:      r.Error,
		})
	}

	return Snapshot{
		GeneratedAt: time.Now(),
		Running:     running,
		Retrying:    retrying,
		Totals:      o.totals,
	}
}

// --- Internal: tick ---

func (o *Orchestrator) tick(ctx context.Context) {
	o.reconcile(ctx)

	o.mu.Lock()
	cfg := o.cfg
	o.mu.Unlock()

	if err := cfg.Validate(); err != nil {
		o.logger.Error("dispatch preflight failed (skipping dispatch)", "error", err)
		return
	}

	candidates, err := o.tracker.FetchCandidateIssues()
	if err != nil {
		o.logger.Error("tracker fetch failed (skipping dispatch)", "error", err)
		return
	}

	sortForDispatch(candidates)

	o.mu.Lock()
	defer o.mu.Unlock()

	for _, issue := range candidates {
		if o.availableSlots() <= 0 {
			break
		}
		if o.shouldDispatch(issue) {
			o.dispatch(ctx, issue, nil)
		}
	}
}

// --- Internal: dispatch ---

func (o *Orchestrator) dispatch(ctx context.Context, issue serve.Issue, attempt *int) {
	issueCtx, cancel := context.WithCancel(ctx)

	entry := &serve.RunningEntry{
		Issue:     issue,
		Attempt:   attempt,
		StartedAt: time.Now(),
		CancelFunc: cancel,
	}

	o.running[issue.ID] = entry
	o.claimed[issue.ID] = true
	delete(o.retries, issue.ID)

	o.logger.Info("dispatching issue",
		"issue_id", issue.ID,
		"issue_identifier", issue.Identifier,
		"attempt", fmtAttempt(attempt),
	)

	go o.runWorker(issueCtx, issue, attempt, entry)
}

func (o *Orchestrator) runWorker(ctx context.Context, issue serve.Issue, attempt *int, entry *serve.RunningEntry) {
	defer func() {
		if r := recover(); r != nil {
			o.logger.Error("worker panic", "issue_id", issue.ID, "panic", r)
			o.onWorkerExit(issue, attempt, false, fmt.Sprintf("panic: %v", r))
		}
	}()

	cfg := o.snapshotConfig()

	// 1. Create workspace
	ws, err := o.workspace.CreateForIssue(issue.Identifier)
	if err != nil {
		o.logger.Error("workspace creation failed", "issue_id", issue.ID, "error", err)
		o.onWorkerExit(issue, attempt, false, err.Error())
		return
	}
	entry.WorkspacePath = ws.Path

	// 2. before_run hook
	if err := o.workspace.BeforeRun(ws.Path); err != nil {
		o.logger.Error("before_run hook failed", "issue_id", issue.ID, "error", err)
		o.workspace.AfterRun(ws.Path)
		o.onWorkerExit(issue, attempt, false, err.Error())
		return
	}

	// 3. Multi-turn loop
	turnNumber := 1
	maxTurns := cfg.MaxTurns()
	success := false
	var runErr string

	for turnNumber <= maxTurns {
		// Build prompt
		prompt, err := cfg.RenderPrompt(&issue, attempt)
		if err != nil {
			o.logger.Error("prompt render failed", "issue_id", issue.ID, "error", err)
			runErr = err.Error()
			break
		}

		o.logger.Info("starting turn",
			"issue_id", issue.ID,
			"issue_identifier", issue.Identifier,
			"turn", turnNumber,
		)

		result := o.runner.Run(ctx, &issue, prompt, ws.Path)

		// Update session metrics
		o.mu.Lock()
		if e, ok := o.running[issue.ID]; ok {
			e.Session.TurnCount = turnNumber
			e.Session.LastMessage = result.LastMessage
			if result.Tokens != nil {
				e.Session.InputTokens += result.Tokens.InputTokens
				e.Session.OutputTokens += result.Tokens.OutputTokens
				e.Session.TotalTokens += result.Tokens.TotalTokens
			}
		}
		o.mu.Unlock()

		if !result.Success {
			runErr = result.Error
			break
		}

		// After successful turn: re-check issue state
		refreshed, err := o.tracker.FetchIssueStatesByIDs([]string{issue.ID})
		if err != nil || len(refreshed) == 0 {
			o.logger.Warn("issue state refresh failed", "issue_id", issue.ID)
			break
		}

		issue = refreshed[0]
		if !o.isActiveState(issue.State) {
			success = true
			break
		}

		if turnNumber >= maxTurns {
			success = true
			break
		}
		turnNumber++
	}

	// If reconcile detected a terminal state and cancelled our context,
	// treat this as a graceful exit: run AfterRun, clean up, and return
	// without scheduling a retry.
	if entry.TerminalState != "" {
		o.workspace.AfterRun(ws.Path)
		go o.workspace.CleanupForIssue(issue.Identifier)
		o.mu.Lock()
		o.totals.SecondsRunning += time.Since(entry.StartedAt).Seconds()
		o.totals.InputTokens += entry.Session.InputTokens
		o.totals.OutputTokens += entry.Session.OutputTokens
		o.totals.TotalTokens += entry.Session.TotalTokens
		o.mu.Unlock()
		o.logger.Info("worker terminated: issue reached terminal state",
			"issue_id", issue.ID,
			"issue_identifier", issue.Identifier,
			"state", entry.TerminalState,
		)
		return
	}

	o.workspace.AfterRun(ws.Path)
	o.onWorkerExit(issue, attempt, success, runErr)
}

func (o *Orchestrator) onWorkerExit(issue serve.Issue, attempt *int, success bool, errMsg string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	entry, ok := o.running[issue.ID]
	if ok {
		o.totals.SecondsRunning += time.Since(entry.StartedAt).Seconds()
		o.totals.InputTokens += entry.Session.InputTokens
		o.totals.OutputTokens += entry.Session.OutputTokens
		o.totals.TotalTokens += entry.Session.TotalTokens
		entry.CancelFunc()
	}
	delete(o.running, issue.ID)

	if success || errMsg == "" {
		o.logger.Info("worker exited normally",
			"issue_id", issue.ID,
			"issue_identifier", issue.Identifier,
		)
		// Schedule short continuation retry (1 second)
		o.scheduleRetry(issue.ID, issue.Identifier, 1, 1000, "")
	} else {
		o.logger.Error("worker exited with error",
			"issue_id", issue.ID,
			"issue_identifier", issue.Identifier,
			"error", errMsg,
		)
		nextAttempt := nextAttemptNum(attempt)
		delayMS := o.backoffDelayMS(nextAttempt)
		o.scheduleRetry(issue.ID, issue.Identifier, nextAttempt, delayMS, errMsg)
	}
}

// --- Internal: retry ---

func (o *Orchestrator) scheduleRetry(issueID, identifier string, attempt, delayMS int, errMsg string) {
	// Cancel existing retry timer if any
	delete(o.retries, issueID)

	entry := &serve.RetryEntry{
		IssueID:    issueID,
		Identifier: identifier,
		Attempt:    attempt,
		DueAt:      time.Now().Add(time.Duration(delayMS) * time.Millisecond),
		Error:      errMsg,
	}
	o.retries[issueID] = entry

	o.logger.Info("retry scheduled",
		"issue_id", issueID,
		"issue_identifier", identifier,
		"attempt", attempt,
		"delay_ms", delayMS,
		"error", errMsg,
	)

	go func() {
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
		o.fireRetry(issueID)
	}()
}

func (o *Orchestrator) fireRetry(issueID string) {
	o.mu.Lock()
	retryEntry, ok := o.retries[issueID]
	if !ok {
		o.mu.Unlock()
		return
	}
	delete(o.retries, issueID)
	o.mu.Unlock()

	candidates, err := o.tracker.FetchCandidateIssues()
	if err != nil {
		o.mu.Lock()
		o.scheduleRetry(issueID, retryEntry.Identifier,
			retryEntry.Attempt+1,
			o.backoffDelayMS(retryEntry.Attempt+1),
			"retry poll failed",
		)
		o.mu.Unlock()
		return
	}

	var found *serve.Issue
	for _, c := range candidates {
		if c.ID == issueID {
			found = &c
			break
		}
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if found == nil {
		// Issue no longer active — release claim
		delete(o.claimed, issueID)
		o.logger.Info("retry: issue no longer active, releasing", "issue_id", issueID)
		return
	}

	if o.availableSlots() <= 0 {
		o.scheduleRetry(issueID, found.Identifier,
			retryEntry.Attempt+1,
			o.backoffDelayMS(retryEntry.Attempt+1),
			"no available orchestrator slots",
		)
		return
	}

	attempt := retryEntry.Attempt
	o.dispatch(context.Background(), *found, &attempt)
}

// --- Internal: reconciliation ---

func (o *Orchestrator) reconcile(ctx context.Context) {
	o.mu.Lock()
	// Stall detection
	stallMS := o.stallTimeoutMS
	for issueID, entry := range o.running {
		if stallMS <= 0 {
			break
		}
		lastActivity := entry.StartedAt
		if entry.Session.LastEventAt != nil {
			lastActivity = *entry.Session.LastEventAt
		}
		if time.Since(lastActivity) > time.Duration(stallMS)*time.Millisecond {
			o.logger.Warn("stall detected, killing worker",
				"issue_id", issueID,
				"issue_identifier", entry.Issue.Identifier,
			)
			entry.CancelFunc()
		}
	}

	runningIDs := make([]string, 0, len(o.running))
	for id := range o.running {
		runningIDs = append(runningIDs, id)
	}
	o.mu.Unlock()

	if len(runningIDs) == 0 {
		return
	}

	refreshed, err := o.tracker.FetchIssueStatesByIDs(runningIDs)
	if err != nil {
		o.logger.Warn("reconciliation state refresh failed (keeping workers)", "error", err)
		return
	}

	refreshMap := make(map[string]serve.Issue, len(refreshed))
	for _, issue := range refreshed {
		refreshMap[issue.ID] = issue
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	for issueID, entry := range o.running {
		issue, ok := refreshMap[issueID]
		if !ok {
			continue
		}
		stateNorm := normalize(issue.State)
		if o.terminalStates[stateNorm] {
			o.logger.Info("reconcile: terminal state, stopping worker",
				"issue_id", issueID,
				"state", issue.State,
			)
			// Signal runWorker to handle AfterRun + cleanup before exiting.
			// Do NOT call CleanupForIssue here — runWorker calls it after AfterRun.
			entry.TerminalState = issue.State
			entry.CancelFunc()
			delete(o.running, issueID)
			delete(o.claimed, issueID)
		} else if o.activeStates[stateNorm] {
			entry.Issue = issue // update snapshot
		} else {
			o.logger.Info("reconcile: non-active state, stopping worker (no cleanup)",
				"issue_id", issueID,
				"state", issue.State,
			)
			entry.CancelFunc()
			delete(o.running, issueID)
			delete(o.claimed, issueID)
		}
	}
}

// --- Internal: helpers ---

func (o *Orchestrator) shouldDispatch(issue serve.Issue) bool {
	if o.claimed[issue.ID] || o.running[issue.ID] != nil {
		return false
	}
	if !o.isActiveState(issue.State) {
		return false
	}
	// Blocker check for "Todo" state
	if normalize(issue.State) == "todo" {
		for _, b := range issue.BlockedBy {
			if !o.terminalStates[normalize(b.State)] {
				return false
			}
		}
	}
	return true
}

func (o *Orchestrator) availableSlots() int {
	avail := o.maxConcurrentAgents - len(o.running)
	if avail < 0 {
		return 0
	}
	return avail
}

func (o *Orchestrator) isActiveState(state string) bool {
	return o.activeStates[normalize(state)]
}

func (o *Orchestrator) snapshotConfig() *workflow.Config {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.cfg
}

// backoffDelayMS computes exponential backoff delay.
// Formula: min(10000 * 2^(attempt-1), max_retry_backoff_ms)
func (o *Orchestrator) backoffDelayMS(attempt int) int {
	maxMS := o.cfg.MaxRetryBackoffMS()
	if attempt <= 1 {
		return 10000
	}
	delay := 10000 * math.Pow(2, float64(attempt-1))
	if delay > float64(maxMS) {
		return maxMS
	}
	return int(delay)
}

// sortForDispatch sorts issues by priority (asc), then created_at (oldest first).
func sortForDispatch(issues []serve.Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		pi := issuePriority(issues[i])
		pj := issuePriority(issues[j])
		if pi != pj {
			return pi < pj
		}
		if issues[i].CreatedAt != nil && issues[j].CreatedAt != nil {
			return issues[i].CreatedAt.Before(*issues[j].CreatedAt)
		}
		return issues[i].Identifier < issues[j].Identifier
	})
}

func issuePriority(issue serve.Issue) int {
	if issue.Priority == nil {
		return 9999
	}
	return *issue.Priority
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func nextAttemptNum(attempt *int) int {
	if attempt == nil {
		return 1
	}
	return *attempt + 1
}

func fmtAttempt(attempt *int) string {
	if attempt == nil {
		return "first"
	}
	return fmt.Sprintf("%d", *attempt)
}

// --- Snapshot types for HTTP API ---

// Snapshot is a point-in-time view of orchestrator state.
type Snapshot struct {
	GeneratedAt time.Time
	Running     []RunningRow
	Retrying    []RetryRow
	Totals      serve.TokenTotals
}

// RunningRow is a summary of one running issue session.
type RunningRow struct {
	IssueID         string
	IssueIdentifier string
	State           string
	SessionID       string
	TurnCount       int
	LastEvent       string
	LastMessage     string
	StartedAt       time.Time
	LastEventAt     *time.Time
	Tokens          serve.TokenUsage
}

// RetryRow is a summary of one queued retry.
type RetryRow struct {
	IssueID    string
	Identifier string
	Attempt    int
	DueAt      time.Time
	Error      string
}
