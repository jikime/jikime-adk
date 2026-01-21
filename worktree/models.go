// Package worktree provides Git worktree management for parallel SPEC development.
package worktree

import (
	"time"
)

// WorktreeInfo represents metadata about a Git worktree.
type WorktreeInfo struct {
	SpecID       string    `json:"spec_id"`
	Path         string    `json:"path"`
	Branch       string    `json:"branch"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	Status       string    `json:"status"` // active, inactive, recovered
}

// ToMap converts WorktreeInfo to a map for JSON serialization.
func (w *WorktreeInfo) ToMap() map[string]any {
	return map[string]any{
		"spec_id":       w.SpecID,
		"path":          w.Path,
		"branch":        w.Branch,
		"created_at":    w.CreatedAt.Format(time.RFC3339),
		"last_accessed": w.LastAccessed.Format(time.RFC3339),
		"status":        w.Status,
	}
}

// DoneResult represents the result of completing a worktree.
type DoneResult struct {
	MergedBranch string `json:"merged_branch"`
	BaseBranch   string `json:"base_branch"`
	Pushed       bool   `json:"pushed"`
}
