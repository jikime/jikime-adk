package team

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Registry tracks spawned agents and checks their liveness.
// Persisted to ~/.jikime/teams/<team>/registry/<agentID>.json.
type Registry struct {
	mu  sync.RWMutex
	dir string
}

// NewRegistry returns a Registry rooted at dir, creating it if needed.
func NewRegistry(dir string) (*Registry, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("team/registry: mkdir %s: %w", dir, err)
	}
	return &Registry{dir: dir}, nil
}

// Register adds or updates an agent record.
func (r *Registry) Register(info *AgentInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if info.JoinedAt.IsZero() {
		info.JoinedAt = time.Now()
	}
	info.LastHeartbeat = time.Now()
	return r.save(info)
}

// Heartbeat updates the LastHeartbeat timestamp for an agent.
func (r *Registry) Heartbeat(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, err := r.load(agentID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("team/registry: agent %s not found", agentID)
	}
	info.LastHeartbeat = time.Now()
	return r.save(info)
}

// SetStatus updates the Status field of an agent.
func (r *Registry) SetStatus(agentID string, status AgentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, err := r.load(agentID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("team/registry: agent %s not found", agentID)
	}
	info.Status = status
	info.LastHeartbeat = time.Now()
	return r.save(info)
}

// SetCurrentTask updates the task currently being worked on by agentID.
func (r *Registry) SetCurrentTask(agentID, taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, err := r.load(agentID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("team/registry: agent %s not found", agentID)
	}
	info.CurrentTaskID = taskID
	info.LastHeartbeat = time.Now()
	return r.save(info)
}

// Get returns the AgentInfo for agentID, or nil if not found.
func (r *Registry) Get(agentID string) (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.load(agentID)
}

// List returns all registered agents.
func (r *Registry) List() ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/registry: readdir: %w", err)
	}

	var agents []*AgentInfo
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		info, err := r.load(id)
		if err != nil || info == nil {
			continue
		}
		agents = append(agents, info)
	}
	return agents, nil
}

// IsAlive checks whether the agent process is still running.
// It returns (true, nil) if alive, (false, nil) if dead, and (false, err) on error.
func (r *Registry) IsAlive(agentID string) (bool, error) {
	r.mu.RLock()
	info, err := r.load(agentID)
	r.mu.RUnlock()
	if err != nil {
		return false, err
	}
	if info == nil {
		return false, nil
	}

	// 1. Try tmux session check.
	if info.TmuxSession != "" {
		alive := tmuxAlive(info.TmuxSession)
		return alive, nil
	}

	// 2. Fallback to PID check.
	if info.PID > 0 {
		return pidAlive(info.PID), nil
	}

	// 3. Heartbeat staleness check (>30s without heartbeat → offline).
	if time.Since(info.LastHeartbeat) > 30*time.Second {
		return false, nil
	}
	return true, nil
}

// DeadAgents returns the IDs of all agents whose process is no longer alive.
func (r *Registry) DeadAgents() ([]string, error) {
	agents, err := r.List()
	if err != nil {
		return nil, err
	}
	var dead []string
	for _, a := range agents {
		if a.Status == AgentStatusOffline {
			dead = append(dead, a.ID)
			continue
		}
		alive, err := r.IsAlive(a.ID)
		if err != nil {
			continue
		}
		if !alive {
			dead = append(dead, a.ID)
		}
	}
	return dead, nil
}

// MarkDead sets an agent's status to offline.
func (r *Registry) MarkDead(agentID string) error {
	return r.SetStatus(agentID, AgentStatusOffline)
}

// Remove deletes an agent record from the registry.
func (r *Registry) Remove(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	err := os.Remove(r.path(agentID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// --- internal helpers ---

func (r *Registry) load(agentID string) (*AgentInfo, error) {
	data, err := os.ReadFile(r.path(agentID))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/registry: read %s: %w", agentID, err)
	}
	var info AgentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("team/registry: unmarshal %s: %w", agentID, err)
	}
	return &info, nil
}

func (r *Registry) save(info *AgentInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("team/registry: marshal %s: %w", info.ID, err)
	}
	tmp := r.path(info.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("team/registry: write %s: %w", info.ID, err)
	}
	if err := os.Rename(tmp, r.path(info.ID)); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("team/registry: rename %s: %w", info.ID, err)
	}
	return nil
}

func (r *Registry) path(agentID string) string {
	return filepath.Join(r.dir, agentID+".json")
}

// tmuxAlive returns true if the tmux session (or session:window) is running
// AND a non-shell process (i.e. claude) is still active in the pane.
//
// ClawTeam-compatible liveness check:
//   - pane_dead=1 → pane is dead
//   - pane_current_command is bash/zsh/sh/fish → claude exited, only shell remains
func tmuxAlive(target string) bool {
	// "jikime-myteam:worker-1" → check pane state.
	if strings.Contains(target, ":") {
		out, err := exec.Command("tmux", "list-panes", "-t", target,
			"-F", "#{pane_dead} #{pane_current_command}").Output()
		if err != nil {
			return false
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			if parts[0] == "1" {
				return false // pane_dead=1
			}
			switch parts[1] {
			case "bash", "zsh", "sh", "fish", "dash":
				return false // claude exited; only shell remains
			}
		}
		return true
	}
	// Session-only target: just check if session exists.
	return exec.Command("tmux", "has-session", "-t", target).Run() == nil
}

// pidAlive returns true if the process with the given PID is still running.
func pidAlive(pid int) bool {
	// On Unix, kill(pid, 0) succeeds if the process exists.
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// os.FindProcess always succeeds on Unix; send signal 0 to probe.
	data, err := os.ReadFile(fmt.Sprintf("/proc/%s/status", strconv.Itoa(pid)))
	if err == nil && len(data) > 0 {
		return true
	}
	// Fallback: use kill -0 via shell.
	err = exec.Command("kill", "-0", strconv.Itoa(proc.Pid)).Run()
	return err == nil
}
