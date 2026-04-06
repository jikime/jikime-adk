package worktree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Registry manages worktree metadata persistence.
type Registry struct {
	root     string
	filePath string
	mu       sync.RWMutex
}

// registryData represents the JSON structure of the registry file.
type registryData struct {
	Worktrees map[string]map[string]*WorktreeInfo `json:"worktrees"` // project -> spec_id -> info
}

// NewRegistry creates a new Registry instance.
func NewRegistry(worktreeRoot string) *Registry {
	return &Registry{
		root:     worktreeRoot,
		filePath: filepath.Join(worktreeRoot, ".jikime-worktree-registry.json"),
	}
}

// Register adds or updates a worktree in the registry.
func (r *Registry) Register(info *WorktreeInfo, projectName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.load()
	if err != nil {
		data = &registryData{Worktrees: make(map[string]map[string]*WorktreeInfo)}
	}

	if data.Worktrees[projectName] == nil {
		data.Worktrees[projectName] = make(map[string]*WorktreeInfo)
	}

	data.Worktrees[projectName][info.SpecID] = info

	return r.save(data)
}

// Unregister removes a worktree from the registry.
func (r *Registry) Unregister(specID, projectName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.load()
	if err != nil {
		return nil // Nothing to unregister
	}

	if data.Worktrees[projectName] != nil {
		delete(data.Worktrees[projectName], specID)
	}

	return r.save(data)
}

// Get retrieves a worktree info by spec ID.
func (r *Registry) Get(specID, projectName string) *WorktreeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.load()
	if err != nil {
		return nil
	}

	if data.Worktrees[projectName] == nil {
		return nil
	}

	return data.Worktrees[projectName][specID]
}

// ListAll returns all worktrees for a project.
func (r *Registry) ListAll(projectName string) []*WorktreeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.load()
	if err != nil {
		return nil
	}

	if data.Worktrees[projectName] == nil {
		return nil
	}

	result := make([]*WorktreeInfo, 0, len(data.Worktrees[projectName]))
	for _, info := range data.Worktrees[projectName] {
		result = append(result, info)
	}

	return result
}

// SyncWithGit synchronizes the registry with actual Git worktrees.
func (r *Registry) SyncWithGit(repoPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.load()
	if err != nil {
		data = &registryData{Worktrees: make(map[string]map[string]*WorktreeInfo)}
	}

	// Verify each registered worktree still exists
	for projectName, worktrees := range data.Worktrees {
		for specID, info := range worktrees {
			if _, err := os.Stat(info.Path); os.IsNotExist(err) {
				// Worktree directory doesn't exist, mark as inactive
				info.Status = "inactive"
				data.Worktrees[projectName][specID] = info
			}
		}
	}

	return r.save(data)
}

// RecoverFromDisk scans the worktree root directory and recovers worktrees.
func (r *Registry) RecoverFromDisk() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.load()
	if err != nil {
		data = &registryData{Worktrees: make(map[string]map[string]*WorktreeInfo)}
	}

	recovered := 0

	// Scan root directory for project directories
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}

		projectName := entry.Name()
		projectPath := filepath.Join(r.root, projectName)

		// Scan project directory for worktrees
		worktreeEntries, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, wtEntry := range worktreeEntries {
			if !wtEntry.IsDir() {
				continue
			}

			specID := wtEntry.Name()
			worktreePath := filepath.Join(projectPath, specID)
			gitPath := filepath.Join(worktreePath, ".git")

			// Check if it's a valid worktree
			if _, err := os.Stat(gitPath); os.IsNotExist(err) {
				continue
			}

			// Check if already registered
			if data.Worktrees[projectName] != nil {
				if _, exists := data.Worktrees[projectName][specID]; exists {
					continue
				}
			}

			// Try to get branch name from .git file
			branch := r.getBranchFromWorktree(worktreePath)
			if branch == "" {
				branch = "feature/" + specID
			}

			// Register the recovered worktree
			if data.Worktrees[projectName] == nil {
				data.Worktrees[projectName] = make(map[string]*WorktreeInfo)
			}

			now := time.Now()
			data.Worktrees[projectName][specID] = &WorktreeInfo{
				SpecID:       specID,
				Path:         worktreePath,
				Branch:       branch,
				CreatedAt:    now,
				LastAccessed: now,
				Status:       "recovered",
			}
			recovered++
		}
	}

	if recovered > 0 {
		if err := r.save(data); err != nil {
			return recovered, err
		}
	}

	return recovered, nil
}

// getBranchFromWorktree extracts the branch name from a worktree's .git file.
func (r *Registry) getBranchFromWorktree(worktreePath string) string {
	gitPath := filepath.Join(worktreePath, ".git")

	content, err := os.ReadFile(gitPath)
	if err != nil {
		return ""
	}

	// .git file in worktree contains: gitdir: /path/to/main/.git/worktrees/<name>
	// We need to read HEAD from the main repo's worktrees directory
	var gitdir string
	if _, err := filepath.Rel(r.root, worktreePath); err == nil {
		// Try to parse gitdir from .git file
		lines := string(content)
		if len(lines) > 8 && lines[:8] == "gitdir: " {
			gitdir = lines[8:]
			gitdir = filepath.Clean(gitdir)

			// Read HEAD from the worktree's gitdir
			headPath := filepath.Join(gitdir, "HEAD")
			headContent, err := os.ReadFile(headPath)
			if err == nil {
				head := string(headContent)
				if len(head) > 16 && head[:16] == "ref: refs/heads/" {
					return head[16 : len(head)-1] // Remove trailing newline
				}
			}
		}
	}

	return ""
}

// load reads the registry file.
func (r *Registry) load() (*registryData, error) {
	content, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &registryData{Worktrees: make(map[string]map[string]*WorktreeInfo)}, nil
		}
		return nil, err
	}

	var data registryData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	if data.Worktrees == nil {
		data.Worktrees = make(map[string]map[string]*WorktreeInfo)
	}

	return &data, nil
}

// save writes the registry file.
func (r *Registry) save(data *registryData) error {
	// Ensure directory exists
	if err := os.MkdirAll(r.root, 0755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, content, 0644)
}

// Path returns the registry file path.
func (r *Registry) Path() string {
	return r.filePath
}
