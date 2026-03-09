package orchestrator

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"jikime-adk/internal/serve"
	"jikime-adk/internal/serve/workflow"
	"jikime-adk/internal/serve/workspace"
)

// --- fake tracker (implements tracker.Client) ---

type fakeTracker struct {
	candidates []serve.Issue
	states     []serve.Issue
	err        error
}

func (f *fakeTracker) FetchCandidateIssues() ([]serve.Issue, error) {
	return f.candidates, f.err
}

func (f *fakeTracker) FetchIssueStatesByIDs(_ []string) ([]serve.Issue, error) {
	return f.states, f.err
}

func (f *fakeTracker) FetchIssuesByStates(_ []string) ([]serve.Issue, error) {
	return f.states, f.err
}

// --- helpers ---

func makeTestOrchestrator(t *testing.T) *Orchestrator {
	t.Helper()
	def, err := workflow.Parse([]byte("---\ntracker:\n  kind: github\n  api_key: token\n  project_slug: owner/repo\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := workflow.NewConfig(def)
	wsm := workspace.NewManager(t.TempDir())
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return New(cfg, &fakeTracker{}, wsm, nil, logger)
}

func intPtr(n int) *int        { return &n }
func timePtr(t time.Time) *time.Time { return &t }

// prepareRunning inserts an entry so onWorkerExit can find it.
func prepareRunning(o *Orchestrator, issue serve.Issue) {
	o.running[issue.ID] = &serve.RunningEntry{
		Issue:      issue,
		StartedAt:  time.Now(),
		CancelFunc: func() {},
	}
}

// --- sortForDispatch ---

func TestSortForDispatch_PriorityAscending(t *testing.T) {
	now := time.Now()
	issues := []serve.Issue{
		{ID: "3", Identifier: "PROJ-3", Priority: nil, CreatedAt: timePtr(now)},
		{ID: "1", Identifier: "PROJ-1", Priority: intPtr(1), CreatedAt: timePtr(now)},
		{ID: "2", Identifier: "PROJ-2", Priority: intPtr(2), CreatedAt: timePtr(now)},
	}
	sortForDispatch(issues)

	want := []string{"1", "2", "3"}
	for i, wantID := range want {
		if issues[i].ID != wantID {
			t.Errorf("index %d: got %q, want %q", i, issues[i].ID, wantID)
		}
	}
}

func TestSortForDispatch_CreatedAtTiebreak(t *testing.T) {
	base := time.Now()
	older := base.Add(-1 * time.Hour)
	newer := base

	issues := []serve.Issue{
		{ID: "B", Identifier: "PROJ-B", Priority: intPtr(1), CreatedAt: timePtr(newer)},
		{ID: "A", Identifier: "PROJ-A", Priority: intPtr(1), CreatedAt: timePtr(older)},
	}
	sortForDispatch(issues)

	if issues[0].ID != "A" {
		t.Errorf("expected older issue A first, got %q", issues[0].ID)
	}
}

func TestSortForDispatch_IdentifierLexicographic(t *testing.T) {
	// No CreatedAt — falls back to identifier lexicographic order
	issues := []serve.Issue{
		{ID: "2", Identifier: "PROJ-B", Priority: intPtr(1)},
		{ID: "1", Identifier: "PROJ-A", Priority: intPtr(1)},
	}
	sortForDispatch(issues)

	if issues[0].Identifier != "PROJ-A" {
		t.Errorf("expected PROJ-A first (lexicographic), got %q", issues[0].Identifier)
	}
}

// --- shouldDispatch ---

func TestShouldDispatch_SkipsClaimed(t *testing.T) {
	o := makeTestOrchestrator(t)
	o.claimed["issue-1"] = true

	issue := serve.Issue{ID: "issue-1", State: "In Progress"}
	if o.shouldDispatch(issue) {
		t.Error("shouldDispatch should return false for claimed issue")
	}
}

func TestShouldDispatch_SkipsRunning(t *testing.T) {
	o := makeTestOrchestrator(t)
	o.running["issue-1"] = &serve.RunningEntry{
		Issue:      serve.Issue{ID: "issue-1"},
		CancelFunc: func() {},
	}

	issue := serve.Issue{ID: "issue-1", State: "In Progress"}
	if o.shouldDispatch(issue) {
		t.Error("shouldDispatch should return false for already-running issue")
	}
}

func TestShouldDispatch_InactiveState(t *testing.T) {
	o := makeTestOrchestrator(t)

	// "Done" is terminal (default), not active
	issue := serve.Issue{ID: "issue-1", State: "Done"}
	if o.shouldDispatch(issue) {
		t.Error("shouldDispatch should return false for non-active state")
	}
}

func TestShouldDispatch_TodoWithNonTerminalBlocker(t *testing.T) {
	o := makeTestOrchestrator(t)

	issue := serve.Issue{
		ID:    "issue-2",
		State: "Todo",
		BlockedBy: []serve.BlockerRef{
			{ID: "blocker-1", Identifier: "PROJ-1", State: "In Progress"}, // not terminal
		},
	}
	if o.shouldDispatch(issue) {
		t.Error("shouldDispatch should return false for Todo issue with non-terminal blocker")
	}
}

func TestShouldDispatch_TodoWithAllTerminalBlockers(t *testing.T) {
	o := makeTestOrchestrator(t)

	issue := serve.Issue{
		ID:    "issue-3",
		State: "Todo",
		BlockedBy: []serve.BlockerRef{
			{ID: "blocker-1", Identifier: "PROJ-1", State: "Done"},
			{ID: "blocker-2", Identifier: "PROJ-2", State: "Closed"},
		},
	}
	if !o.shouldDispatch(issue) {
		t.Error("shouldDispatch should return true for Todo issue with all-terminal blockers")
	}
}

func TestShouldDispatch_NonTodoActiveStateIgnoresBlockers(t *testing.T) {
	o := makeTestOrchestrator(t)

	// "In Progress" is active but NOT "Todo" → blocker check doesn't apply
	issue := serve.Issue{
		ID:    "issue-4",
		State: "In Progress",
		BlockedBy: []serve.BlockerRef{
			{ID: "blocker-1", State: "In Progress"}, // non-terminal
		},
	}
	if !o.shouldDispatch(issue) {
		t.Error("shouldDispatch should return true for non-Todo active state regardless of blockers")
	}
}

// --- backoffDelayMS ---

func TestBackoffDelayMS_Formula(t *testing.T) {
	o := makeTestOrchestrator(t)

	// Formula: min(10000 * 2^(attempt-1), max_retry_backoff_ms)
	tests := []struct {
		attempt int
		wantMS  int
	}{
		{1, 10000},          // 10000 * 2^0 = 10000
		{2, 20000},          // 10000 * 2^1 = 20000
		{3, 40000},          // 10000 * 2^2 = 40000
		{4, 80000},          // 10000 * 2^3 = 80000
		{5, 160000},         // 10000 * 2^4 = 160000
	}
	for _, tt := range tests {
		got := o.backoffDelayMS(tt.attempt)
		if got != tt.wantMS {
			t.Errorf("backoffDelayMS(%d) = %d, want %d", tt.attempt, got, tt.wantMS)
		}
	}
}

func TestBackoffDelayMS_CappedAtMaxRetryBackoff(t *testing.T) {
	o := makeTestOrchestrator(t)

	// Default max_retry_backoff_ms = 300000
	// attempt=6: 10000 * 2^5 = 320000 → capped at 300000
	got := o.backoffDelayMS(6)
	if got != 300000 {
		t.Errorf("backoffDelayMS(6) = %d, want 300000 (capped at max)", got)
	}

	// attempt=10: 10000 * 2^9 = 5120000 → capped at 300000
	got = o.backoffDelayMS(10)
	if got != 300000 {
		t.Errorf("backoffDelayMS(10) = %d, want 300000 (capped at max)", got)
	}
}

// --- availableSlots ---

func TestAvailableSlots_AllFree(t *testing.T) {
	o := makeTestOrchestrator(t)
	// maxConcurrentAgents = 10 (default), running = 0
	got := o.availableSlots()
	if got != 10 {
		t.Errorf("availableSlots() = %d, want 10", got)
	}
}

func TestAvailableSlots_AtCapacity(t *testing.T) {
	o := makeTestOrchestrator(t)
	o.maxConcurrentAgents = 3
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("issue-%d", i)
		o.running[id] = &serve.RunningEntry{
			Issue:      serve.Issue{ID: id},
			CancelFunc: func() {},
		}
	}
	got := o.availableSlots()
	if got != 0 {
		t.Errorf("availableSlots() = %d, want 0 (at capacity)", got)
	}
}

func TestAvailableSlots_NeverNegative(t *testing.T) {
	o := makeTestOrchestrator(t)
	o.maxConcurrentAgents = 1
	for i := 0; i < 2; i++ { // over capacity
		id := fmt.Sprintf("issue-%d", i)
		o.running[id] = &serve.RunningEntry{
			Issue:      serve.Issue{ID: id},
			CancelFunc: func() {},
		}
	}
	got := o.availableSlots()
	if got != 0 {
		t.Errorf("availableSlots() = %d, want 0 (never negative)", got)
	}
}

// --- onWorkerExit retry scheduling ---

func TestOnWorkerExit_NormalExit_SchedulesContinuationRetry(t *testing.T) {
	o := makeTestOrchestrator(t)
	issue := serve.Issue{ID: "issue-1", Identifier: "PROJ-1", State: "Todo"}
	prepareRunning(o, issue)

	before := time.Now()
	o.onWorkerExit(issue, nil, true, "") // success = true, no error
	after := time.Now()

	o.mu.Lock()
	retry, ok := o.retries[issue.ID]
	o.mu.Unlock()

	if !ok {
		t.Fatal("expected retry entry scheduled after normal exit")
	}
	if retry.Attempt != 1 {
		t.Errorf("retry.Attempt = %d, want 1 (continuation retry)", retry.Attempt)
	}
	if retry.Error != "" {
		t.Errorf("retry.Error = %q, want empty for normal exit", retry.Error)
	}
	// DueAt should be ~1000ms in the future
	minDue := before.Add(900 * time.Millisecond)
	maxDue := after.Add(1100 * time.Millisecond)
	if retry.DueAt.Before(minDue) || retry.DueAt.After(maxDue) {
		t.Errorf("retry.DueAt = %v, want in range [%v, %v]", retry.DueAt, minDue, maxDue)
	}
}

func TestOnWorkerExit_FailureFirstDispatch_Backoff10s(t *testing.T) {
	o := makeTestOrchestrator(t)
	issue := serve.Issue{ID: "issue-2", Identifier: "PROJ-2", State: "In Progress"}
	prepareRunning(o, issue)

	// attempt=nil (first dispatch) → nextAttempt=1 → delay=10000ms
	before := time.Now()
	o.onWorkerExit(issue, nil, false, "claude exited non-zero")
	after := time.Now()

	o.mu.Lock()
	retry, ok := o.retries[issue.ID]
	o.mu.Unlock()

	if !ok {
		t.Fatal("expected retry entry scheduled after failure exit")
	}
	if retry.Attempt != 1 {
		t.Errorf("retry.Attempt = %d, want 1", retry.Attempt)
	}
	if retry.Error != "claude exited non-zero" {
		t.Errorf("retry.Error = %q, want %q", retry.Error, "claude exited non-zero")
	}
	// DueAt should be ~10000ms in the future
	minDue := before.Add(9900 * time.Millisecond)
	maxDue := after.Add(10100 * time.Millisecond)
	if retry.DueAt.Before(minDue) || retry.DueAt.After(maxDue) {
		t.Errorf("retry.DueAt = %v, want in range [%v, %v]", retry.DueAt, minDue, maxDue)
	}
}

func TestOnWorkerExit_FailureSecondDispatch_Backoff20s(t *testing.T) {
	o := makeTestOrchestrator(t)
	issue := serve.Issue{ID: "issue-3", Identifier: "PROJ-3", State: "In Progress"}
	prepareRunning(o, issue)

	// attempt=1 (second dispatch) → nextAttempt=2 → delay=20000ms
	attempt := 1
	before := time.Now()
	o.onWorkerExit(issue, &attempt, false, "stall timeout")
	after := time.Now()

	o.mu.Lock()
	retry, ok := o.retries[issue.ID]
	o.mu.Unlock()

	if !ok {
		t.Fatal("expected retry entry scheduled after failure exit (attempt=1)")
	}
	if retry.Attempt != 2 {
		t.Errorf("retry.Attempt = %d, want 2 (nextAttempt)", retry.Attempt)
	}
	// DueAt should be ~20000ms in the future
	minDue := before.Add(19900 * time.Millisecond)
	maxDue := after.Add(20100 * time.Millisecond)
	if retry.DueAt.Before(minDue) || retry.DueAt.After(maxDue) {
		t.Errorf("retry.DueAt = %v, want in range [%v, %v]", retry.DueAt, minDue, maxDue)
	}
}

func TestOnWorkerExit_RemovesFromRunning(t *testing.T) {
	o := makeTestOrchestrator(t)
	issue := serve.Issue{ID: "issue-4", Identifier: "PROJ-4", State: "In Progress"}
	prepareRunning(o, issue)

	o.onWorkerExit(issue, nil, true, "")

	o.mu.Lock()
	_, stillRunning := o.running[issue.ID]
	o.mu.Unlock()

	if stillRunning {
		t.Error("expected issue to be removed from running map after exit")
	}
}
