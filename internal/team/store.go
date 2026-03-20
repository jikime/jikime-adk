package team

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Store manages tasks for a team with atomic file-based locking.
// Each task is persisted to ~/.jikime/teams/<team>/tasks/<id>.json.
type Store struct {
	mu      sync.Mutex
	taskDir string
}

// NewStore creates a Store rooted at the given task directory.
// The directory is created if it does not exist.
func NewStore(taskDir string) (*Store, error) {
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return nil, fmt.Errorf("team/store: mkdir %s: %w", taskDir, err)
	}
	return &Store{taskDir: taskDir}, nil
}

// Create adds a new task to the store and returns it.
// If any dependsOn IDs are provided the task starts as Blocked.
func (s *Store) Create(title, description, dod string, dependsOn []string, priority int, tags []string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := TaskStatusPending
	if len(dependsOn) > 0 {
		status = TaskStatusBlocked
	}

	t := &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		DoD:         dod,
		Status:      status,
		DependsOn:   dependsOn,
		Priority:    priority,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return t, s.save(t)
}

// Get returns the task with the given ID, or (nil, nil) if not found.
func (s *Store) Get(id string) (*Task, error) {
	// Try exact match first.
	data, err := os.ReadFile(s.path(id))
	if err == nil {
		var t Task
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("team/store: unmarshal %s: %w", id, err)
		}
		return &t, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("team/store: read %s: %w", id, err)
	}
	// Exact file not found — try prefix match for short IDs (first 8 chars).
	entries, rerr := os.ReadDir(s.taskDir)
	if rerr != nil {
		return nil, nil
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		fullID := e.Name()[:len(e.Name())-5]
		if strings.HasPrefix(fullID, id) {
			data, err = os.ReadFile(filepath.Join(s.taskDir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("team/store: read %s: %w", fullID, err)
			}
			var t Task
			if err := json.Unmarshal(data, &t); err != nil {
				return nil, fmt.Errorf("team/store: unmarshal %s: %w", fullID, err)
			}
			return &t, nil
		}
	}
	return nil, nil
}

// List returns all tasks, optionally filtered by status or agentID.
// Pass empty strings to skip a filter.
func (s *Store) List(filterStatus TaskStatus, filterAgentID string) ([]*Task, error) {
	entries, err := os.ReadDir(s.taskDir)
	if err != nil {
		return nil, fmt.Errorf("team/store: readdir: %w", err)
	}
	var tasks []*Task
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		t, err := s.Get(id)
		if err != nil || t == nil {
			continue
		}
		if filterStatus != "" && t.Status != filterStatus {
			continue
		}
		if filterAgentID != "" && t.AgentID != filterAgentID {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Claim atomically assigns a task to agentID and marks it in_progress.
// Returns ErrTaskNotFound, ErrTaskNotClaimable, or ErrTaskLocked on failure.
func (s *Store) Claim(taskID, agentID string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, &ErrTaskNotFound{ID: taskID}
	}
	if t.Status == TaskStatusBlocked {
		return nil, &ErrTaskNotClaimable{ID: taskID, Status: t.Status}
	}
	if t.Status == TaskStatusInProgress {
		if t.AgentID == agentID {
			return t, nil // idempotent
		}
		return nil, &ErrTaskLocked{ID: taskID, LockedBy: t.AgentID}
	}
	if t.Status != TaskStatusPending {
		return nil, &ErrTaskNotClaimable{ID: taskID, Status: t.Status}
	}

	now := time.Now()
	t.Status = TaskStatusInProgress
	t.AgentID = agentID
	t.ClaimedAt = &now
	t.UpdatedAt = now
	return t, s.save(t)
}

// Complete marks a task as done and unblocks any tasks that depended on it.
// result is a short summary of what was produced.
func (s *Store) Complete(taskID, agentID, result string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, &ErrTaskNotFound{ID: taskID}
	}
	if t.AgentID != agentID {
		return nil, fmt.Errorf("team/store: agent %q does not own task %s (owner: %s)", agentID, taskID, t.AgentID)
	}

	now := time.Now()
	t.Status = TaskStatusDone
	t.Result = result
	t.CompletedAt = &now
	t.UpdatedAt = now
	if err := s.save(t); err != nil {
		return nil, err
	}

	// Unblock tasks that depended on this one.
	if err := s.unblock(taskID); err != nil {
		return t, fmt.Errorf("team/store: unblock after complete: %w", err)
	}
	return t, nil
}

// Fail marks a task as failed and releases the claim so it can be retried.
func (s *Store) Fail(taskID, agentID, errMsg string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, &ErrTaskNotFound{ID: taskID}
	}

	now := time.Now()
	t.Status = TaskStatusFailed
	t.ErrorMsg = errMsg
	t.CompletedAt = &now
	t.UpdatedAt = now
	return t, s.save(t)
}

// Release drops the claim on a task and resets it to pending.
// Used when an agent dies before finishing a task.
func (s *Store) Release(taskID string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, &ErrTaskNotFound{ID: taskID}
	}
	if t.Status != TaskStatusInProgress {
		return t, nil // nothing to release
	}

	t.Status = TaskStatusPending
	t.AgentID = ""
	t.ClaimedAt = nil
	t.UpdatedAt = time.Now()
	return t, s.save(t)
}

// Update allows updating title, description, dod, or priority of any task.
func (s *Store) Update(taskID string, title, description, dod string, priority int) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, &ErrTaskNotFound{ID: taskID}
	}
	if title != "" {
		t.Title = title
	}
	if description != "" {
		t.Description = description
	}
	if dod != "" {
		t.DoD = dod
	}
	if priority != 0 {
		t.Priority = priority
	}
	t.UpdatedAt = time.Now()
	return t, s.save(t)
}

// Delete removes a task file from the store.
func (s *Store) Delete(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path(taskID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// unblock scans all tasks and removes taskID from their DependsOn lists.
// If a task's DependsOn becomes empty it is transitioned to pending.
// Caller must hold s.mu.
func (s *Store) unblock(completedID string) error {
	entries, err := os.ReadDir(s.taskDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		t, err := s.Get(id)
		if err != nil || t == nil || t.Status != TaskStatusBlocked {
			continue
		}
		newDeps := make([]string, 0, len(t.DependsOn))
		removed := false
		for _, dep := range t.DependsOn {
			// Support both full UUID and short-prefix (first 8 chars) matching.
			if dep == completedID || strings.HasPrefix(completedID, dep) {
				removed = true
			} else {
				newDeps = append(newDeps, dep)
			}
		}
		if !removed {
			continue
		}
		t.DependsOn = newDeps
		if len(newDeps) == 0 {
			t.Status = TaskStatusPending
		}
		t.UpdatedAt = time.Now()
		if err := s.save(t); err != nil {
			return err
		}
	}
	return nil
}

// save atomically writes a task to disk via temp-file + rename.
// Caller must hold s.mu.
func (s *Store) save(t *Task) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("team/store: marshal %s: %w", t.ID, err)
	}
	tmp := s.path(t.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("team/store: write tmp %s: %w", t.ID, err)
	}
	if err := os.Rename(tmp, s.path(t.ID)); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("team/store: rename %s: %w", t.ID, err)
	}
	return nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.taskDir, id+".json")
}

// ForceStatus persists a task's status (and any other fields already set on t)
// directly, bypassing ownership and transition checks. Intended for operator/
// admin overrides (e.g. manually completing a task that was never claimed).
func (s *Store) ForceStatus(t *Task) error {
	return s.save(t)
}

// --- Sentinel errors ---

// ErrTaskNotFound is returned when a task ID does not exist.
type ErrTaskNotFound struct{ ID string }

func (e *ErrTaskNotFound) Error() string { return fmt.Sprintf("task %s not found", e.ID) }

// ErrTaskNotClaimable is returned when a task cannot be claimed due to its status.
type ErrTaskNotClaimable struct {
	ID     string
	Status TaskStatus
}

func (e *ErrTaskNotClaimable) Error() string {
	return fmt.Sprintf("task %s is not claimable (status: %s)", e.ID, e.Status)
}

// ErrTaskLocked is returned when a task is already claimed by another agent.
type ErrTaskLocked struct {
	ID       string
	LockedBy string
}

func (e *ErrTaskLocked) Error() string {
	return fmt.Sprintf("task %s is locked by agent %s", e.ID, e.LockedBy)
}
