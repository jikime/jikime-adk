package team

import (
	"context"
	"fmt"
	"time"
)

// WaitResult summarises the outcome of a Wait call.
type WaitResult struct {
	Status    string // "completed" | "timeout" | "cancelled" | "error"
	Elapsed   time.Duration
	Total     int
	Done      int
	InProgress int
	Pending   int
	Blocked   int
	Failed    int
}

// WaiterCallbacks groups optional callbacks invoked during Wait.
type WaiterCallbacks struct {
	// OnProgress is called whenever task counts change.
	OnProgress func(r WaitResult)

	// OnAgentDead is called when a dead agent is detected.
	// The agentID and the IDs of tasks it was working on are provided.
	OnAgentDead func(agentID string, taskIDs []string)

	// OnMessage is called for each message drained from the leader inbox.
	OnMessage func(m *Message)
}

// Waiter polls the task store and registry until all tasks are done (or timeout).
type Waiter struct {
	store    *Store
	registry *Registry
	inbox    *Inbox    // leader's inbox, may be nil
	teamName string
	interval time.Duration
	cb       WaiterCallbacks
}

// NewWaiter creates a Waiter. leaderInbox may be nil to skip message draining.
func NewWaiter(store *Store, registry *Registry, leaderInbox *Inbox, teamName string, interval time.Duration, cb WaiterCallbacks) *Waiter {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Waiter{
		store:    store,
		registry: registry,
		inbox:    leaderInbox,
		teamName: teamName,
		interval: interval,
		cb:       cb,
	}
}

// Wait blocks until all tasks reach a terminal state (done or failed),
// the context is cancelled, or an internal error occurs.
func (w *Waiter) Wait(ctx context.Context) (WaitResult, error) {
	start := time.Now()
	var prev WaitResult

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		// Drain leader inbox messages first.
		if w.inbox != nil && w.cb.OnMessage != nil {
			msgs, _ := w.inbox.Receive(50)
			for _, m := range msgs {
				w.cb.OnMessage(m)
			}
		}

		// Check for dead agents and release their tasks.
		if err := w.recoverDeadAgents(); err != nil {
			return WaitResult{Status: "error", Elapsed: time.Since(start)}, err
		}

		// Tally task statuses.
		cur, err := w.tally()
		if err != nil {
			return WaitResult{Status: "error", Elapsed: time.Since(start)}, err
		}
		cur.Elapsed = time.Since(start)

		if cur != prev && w.cb.OnProgress != nil {
			w.cb.OnProgress(cur)
			prev = cur
		}

		// Terminal: all tasks in done or failed (no pending / in_progress / blocked left).
		if cur.Total > 0 && cur.Pending == 0 && cur.InProgress == 0 && cur.Blocked == 0 {
			cur.Status = "completed"
			return cur, nil
		}

		select {
		case <-ctx.Done():
			cur.Status = "cancelled"
			if ctx.Err() == context.DeadlineExceeded {
				cur.Status = "timeout"
			}
			return cur, nil
		case <-ticker.C:
			// next iteration
		}
	}
}

// tally counts tasks by status.
func (w *Waiter) tally() (WaitResult, error) {
	tasks, err := w.store.List("", "")
	if err != nil {
		return WaitResult{}, fmt.Errorf("team/waiter: list tasks: %w", err)
	}
	var r WaitResult
	r.Total = len(tasks)
	for _, t := range tasks {
		switch t.Status {
		case TaskStatusDone:
			r.Done++
		case TaskStatusInProgress:
			r.InProgress++
		case TaskStatusPending:
			r.Pending++
		case TaskStatusBlocked:
			r.Blocked++
		case TaskStatusFailed:
			r.Failed++
		}
	}
	return r, nil
}

// recoverDeadAgents finds agents that are no longer alive and releases
// any in_progress tasks they were holding.
func (w *Waiter) recoverDeadAgents() error {
	deadIDs, err := w.registry.DeadAgents()
	if err != nil {
		return nil // non-fatal
	}
	for _, agentID := range deadIDs {
		// Find tasks locked by this agent.
		tasks, err := w.store.List(TaskStatusInProgress, agentID)
		if err != nil {
			continue
		}
		var taskIDs []string
		for _, t := range tasks {
			if _, err := w.store.Release(t.ID); err == nil {
				taskIDs = append(taskIDs, t.ID)
			}
		}
		_ = w.registry.MarkDead(agentID)
		if w.cb.OnAgentDead != nil && len(taskIDs) > 0 {
			w.cb.OnAgentDead(agentID, taskIDs)
		}
	}
	return nil
}
