package worktree

import "fmt"

// WorktreeExistsError is returned when a worktree already exists.
type WorktreeExistsError struct {
	SpecID string
	Path   string
}

func (e *WorktreeExistsError) Error() string {
	return fmt.Sprintf("worktree already exists for %s at %s", e.SpecID, e.Path)
}

// WorktreeNotFoundError is returned when a worktree is not found.
type WorktreeNotFoundError struct {
	SpecID string
}

func (e *WorktreeNotFoundError) Error() string {
	return fmt.Sprintf("worktree not found: %s", e.SpecID)
}

// UncommittedChangesError is returned when a worktree has uncommitted changes.
type UncommittedChangesError struct {
	SpecID string
}

func (e *UncommittedChangesError) Error() string {
	return fmt.Sprintf("worktree has uncommitted changes: %s", e.SpecID)
}

// MergeConflictError is returned when a merge conflict occurs.
type MergeConflictError struct {
	SpecID          string
	ConflictedFiles []string
}

func (e *MergeConflictError) Error() string {
	return fmt.Sprintf("merge conflict in %s: %v", e.SpecID, e.ConflictedFiles)
}

// GitOperationError is returned when a Git operation fails.
type GitOperationError struct {
	Operation string
	Message   string
}

func (e *GitOperationError) Error() string {
	return fmt.Sprintf("git %s failed: %s", e.Operation, e.Message)
}
