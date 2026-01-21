package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Manager manages Git worktrees for parallel SPEC development.
type Manager struct {
	RepoPath     string
	WorktreeRoot string
	ProjectName  string
	Registry     *Registry
}

// NewManager creates a new worktree Manager.
func NewManager(repoPath, worktreeRoot, projectName string) *Manager {
	if projectName == "" {
		projectName = filepath.Base(repoPath)
	}
	return &Manager{
		RepoPath:     repoPath,
		WorktreeRoot: worktreeRoot,
		ProjectName:  projectName,
		Registry:     NewRegistry(worktreeRoot),
	}
}

// Create creates a new worktree for a SPEC.
func (m *Manager) Create(specID, branchName, baseBranch string, force bool, llmConfigPath string) (*WorktreeInfo, error) {
	// Check if worktree already exists
	existing := m.Registry.Get(specID, m.ProjectName)
	if existing != nil && !force {
		return nil, &WorktreeExistsError{SpecID: specID, Path: existing.Path}
	}

	// If force and exists, remove first
	if existing != nil && force {
		_ = m.Remove(specID, true)
	}

	// Determine branch name
	if branchName == "" {
		branchName = "feature/" + specID
	}

	// Create worktree path with project namespace
	worktreePath := filepath.Join(m.WorktreeRoot, m.ProjectName, specID)

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return nil, &GitOperationError{Operation: "mkdir", Message: err.Error()}
	}

	// Fetch latest from remote (network issues shouldn't block local operations)
	if _, err := m.gitCmd("fetch", "origin"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to fetch from origin: %v\n", err)
	}

	// Create branch if it doesn't exist
	branches, err := m.gitCmd("branch", "--list", branchName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to list branches: %v\n", err)
		branches = "" // Continue with empty result
	}
	if strings.TrimSpace(branches) == "" {
		if _, err := m.gitCmd("branch", branchName, baseBranch); err != nil {
			return nil, &GitOperationError{Operation: "branch", Message: err.Error()}
		}
	}

	// Create worktree
	if _, err := m.gitCmd("worktree", "add", worktreePath, branchName); err != nil {
		return nil, &GitOperationError{Operation: "worktree add", Message: err.Error()}
	}

	// Create WorktreeInfo
	now := time.Now()
	info := &WorktreeInfo{
		SpecID:       specID,
		Path:         worktreePath,
		Branch:       branchName,
		CreatedAt:    now,
		LastAccessed: now,
		Status:       "active",
	}

	// Register in registry
	if err := m.Registry.Register(info, m.ProjectName); err != nil {
		return nil, err
	}

	// Copy LLM config to worktree if provided
	if llmConfigPath != "" {
		if err := m.copyLLMConfig(worktreePath, llmConfigPath); err != nil {
			// Log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: Failed to copy LLM config: %v\n", err)
		}
	}

	return info, nil
}

// Remove removes a worktree.
func (m *Manager) Remove(specID string, force bool) error {
	info := m.Registry.Get(specID, m.ProjectName)
	if info == nil {
		return &WorktreeNotFoundError{SpecID: specID}
	}

	// Check for uncommitted changes
	if !force {
		if hasChanges, _ := m.hasUncommittedChanges(info.Path); hasChanges {
			return &UncommittedChangesError{SpecID: specID}
		}
	}

	// Remove worktree using git command
	args := []string{"worktree", "remove", info.Path}
	if force {
		args = append(args, "--force")
	}

	if _, err := m.gitCmd(args...); err != nil {
		// Try to remove directory manually
		if err := os.RemoveAll(info.Path); err != nil {
			return &GitOperationError{Operation: "worktree remove", Message: err.Error()}
		}
	}

	// Unregister from registry
	return m.Registry.Unregister(specID, m.ProjectName)
}

// List returns all worktrees for the current project.
func (m *Manager) List() []*WorktreeInfo {
	return m.Registry.ListAll(m.ProjectName)
}

// Sync synchronizes a worktree with the base branch.
func (m *Manager) Sync(specID, baseBranch string, rebase, ffOnly, autoResolve bool) error {
	info := m.Registry.Get(specID, m.ProjectName)
	if info == nil {
		return &WorktreeNotFoundError{SpecID: specID}
	}

	// Fetch latest (network issues shouldn't block local operations)
	if _, err := m.gitCmdInDir(info.Path, "fetch", "origin"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to fetch from origin: %v\n", err)
	}

	// Determine target branch
	targetBranch := "origin/" + baseBranch
	if _, err := m.gitCmdInDir(info.Path, "rev-parse", targetBranch); err != nil {
		targetBranch = baseBranch
		if _, err := m.gitCmdInDir(info.Path, "rev-parse", targetBranch); err != nil {
			return &GitOperationError{Operation: "rev-parse", Message: fmt.Sprintf("base branch '%s' not found", baseBranch)}
		}
	}

	// Perform sync
	var err error
	if ffOnly {
		_, err = m.gitCmdInDir(info.Path, "merge", targetBranch, "--ff-only")
	} else if rebase {
		_, err = m.gitCmdInDir(info.Path, "rebase", targetBranch)
	} else {
		_, err = m.gitCmdInDir(info.Path, "merge", targetBranch)
	}

	if err != nil {
		// Check for conflicts
		conflicted := m.getConflictedFiles(info.Path)
		if len(conflicted) > 0 {
			if autoResolve {
				if resolveErr := m.autoResolveConflicts(info.Path, specID, conflicted); resolveErr != nil {
					// Abort and return error
					m.abortMerge(info.Path, rebase)
					return &MergeConflictError{SpecID: specID, ConflictedFiles: conflicted}
				}
			} else {
				m.abortMerge(info.Path, rebase)
				return &MergeConflictError{SpecID: specID, ConflictedFiles: conflicted}
			}
		} else {
			return &GitOperationError{Operation: "sync", Message: err.Error()}
		}
	}

	// Update last accessed time
	info.LastAccessed = time.Now()
	return m.Registry.Register(info, m.ProjectName)
}

// CleanMerged removes worktrees for merged branches.
func (m *Manager) CleanMerged() []string {
	var cleaned []string

	// Get merged branches
	output, err := m.gitCmd("branch", "--merged", "main")
	if err != nil {
		return cleaned
	}

	mergedBranches := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		branch := strings.TrimSpace(strings.TrimPrefix(line, "*"))
		if branch != "" {
			mergedBranches[branch] = true
		}
	}

	// Check each worktree
	for _, info := range m.List() {
		if mergedBranches[info.Branch] {
			if err := m.Remove(info.SpecID, true); err == nil {
				cleaned = append(cleaned, info.SpecID)
			}
		}
	}

	return cleaned
}

// Done completes the worktree workflow: merge to base and cleanup.
func (m *Manager) Done(specID, baseBranch string, push, force bool) (*DoneResult, error) {
	info := m.Registry.Get(specID, m.ProjectName)
	if info == nil {
		return nil, &WorktreeNotFoundError{SpecID: specID}
	}

	mergedBranch := info.Branch
	pushed := false

	// Fetch latest (network issues shouldn't block local operations)
	if _, err := m.gitCmd("fetch", "origin"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to fetch from origin: %v\n", err)
	}

	// Checkout base branch
	if _, err := m.gitCmd("checkout", baseBranch); err != nil {
		return nil, &GitOperationError{Operation: "checkout", Message: err.Error()}
	}

	// Merge worktree branch
	if _, err := m.gitCmd("merge", mergedBranch, "--no-ff", "-m", fmt.Sprintf("Merge %s into %s", mergedBranch, baseBranch)); err != nil {
		// Check for conflicts
		conflicted := m.getConflictedFiles(m.RepoPath)
		if len(conflicted) > 0 {
			// Best effort cleanup - merge abort errors are non-critical
			if _, abortErr := m.gitCmd("merge", "--abort"); abortErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to abort merge: %v\n", abortErr)
			}
			return nil, &MergeConflictError{SpecID: specID, ConflictedFiles: conflicted}
		}
		return nil, &GitOperationError{Operation: "merge", Message: err.Error()}
	}

	// Push if requested
	if push {
		if _, err := m.gitCmd("push", "origin", baseBranch); err != nil {
			return nil, &GitOperationError{Operation: "push", Message: err.Error()}
		}
		pushed = true
	}

	// Remove worktree
	if err := m.Remove(specID, force); err != nil {
		return nil, err
	}

	// Delete branch (cleanup task, not critical)
	if _, err := m.gitCmd("branch", "-d", mergedBranch); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to delete branch %s: %v\n", mergedBranch, err)
	}

	return &DoneResult{
		MergedBranch: mergedBranch,
		BaseBranch:   baseBranch,
		Pushed:       pushed,
	}, nil
}

// gitCmd runs a git command in the main repository.
func (m *Manager) gitCmd(args ...string) (string, error) {
	return m.gitCmdInDir(m.RepoPath, args...)
}

// gitCmdInDir runs a git command in a specific directory.
func (m *Manager) gitCmdInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// hasUncommittedChanges checks if a directory has uncommitted changes.
func (m *Manager) hasUncommittedChanges(path string) (bool, error) {
	output, err := m.gitCmdInDir(path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// getConflictedFiles returns files with merge conflicts.
func (m *Manager) getConflictedFiles(path string) []string {
	output, err := m.gitCmdInDir(path, "status", "--porcelain")
	if err != nil {
		return nil
	}

	var conflicted []string
	for _, line := range strings.Split(output, "\n") {
		if len(line) >= 2 {
			status := line[:2]
			if status == "UU" || status == "DD" || status == "AA" || status == "DU" {
				conflicted = append(conflicted, strings.TrimSpace(line[3:]))
			}
		}
	}
	return conflicted
}

// abortMerge aborts an in-progress merge or rebase (best effort cleanup).
func (m *Manager) abortMerge(path string, isRebase bool) {
	if isRebase {
		if _, err := m.gitCmdInDir(path, "rebase", "--abort"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to abort rebase: %v\n", err)
		}
	} else {
		if _, err := m.gitCmdInDir(path, "merge", "--abort"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to abort merge: %v\n", err)
		}
	}
}

// autoResolveConflicts attempts to automatically resolve merge conflicts.
func (m *Manager) autoResolveConflicts(path, specID string, conflicted []string) error {
	for _, file := range conflicted {
		// Try to accept our changes
		if _, err := m.gitCmdInDir(path, "checkout", "--ours", file); err == nil {
			if _, addErr := m.gitCmdInDir(path, "add", file); addErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to stage %s after checkout --ours: %v\n", file, addErr)
			}
			continue
		}

		// Try to accept their changes
		if _, err := m.gitCmdInDir(path, "checkout", "--theirs", file); err == nil {
			if _, addErr := m.gitCmdInDir(path, "add", file); addErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to stage %s after checkout --theirs: %v\n", file, addErr)
			}
			continue
		}

		// Try to remove conflict markers
		filePath := filepath.Join(path, file)
		if err := m.removeConflictMarkers(filePath); err == nil {
			if _, addErr := m.gitCmdInDir(path, "add", file); addErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to stage %s after removing conflict markers: %v\n", file, addErr)
			}
			continue
		}

		return fmt.Errorf("failed to resolve conflict in %s", file)
	}

	// Commit the resolution
	_, err := m.gitCmdInDir(path, "commit", "-m", fmt.Sprintf("Auto-resolved conflicts during sync of %s", specID))
	return err
}

// removeConflictMarkers removes Git conflict markers from a file.
func (m *Manager) removeConflictMarkers(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Remove conflict markers using regex
	re := regexp.MustCompile(`(?m)^<<<<<<<.*\n|^=======\n|^>>>>>>>.*\n`)
	cleaned := re.ReplaceAll(content, []byte{})

	return os.WriteFile(filePath, cleaned, 0644)
}

// copyLLMConfig copies LLM config to the worktree with environment variable substitution.
func (m *Manager) copyLLMConfig(worktreePath, llmConfigPath string) error {
	content, err := os.ReadFile(llmConfigPath)
	if err != nil {
		return err
	}

	// Substitute environment variables (${VAR_NAME} pattern)
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	substituted := re.ReplaceAllFunc(content, func(match []byte) []byte {
		varName := string(match[2 : len(match)-1])
		if value := os.Getenv(varName); value != "" {
			return []byte(value)
		}
		return match
	})

	// Create .claude directory
	claudeDir := filepath.Join(worktreePath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}

	// Write to settings.local.json
	targetPath := filepath.Join(claudeDir, "settings.local.json")
	return os.WriteFile(targetPath, substituted, 0644)
}
