package team

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CostStore persists and aggregates token cost events per agent.
// Events are stored as individual JSON files under:
//
//	~/.jikime/teams/<team>/costs/<agentID>-<timestamp>-<id>.json
type CostStore struct {
	mu      sync.Mutex
	costDir string
	budget  int // 0 = no limit
}

// NewCostStore creates a CostStore rooted at costDir.
func NewCostStore(costDir string, budget int) (*CostStore, error) {
	if err := os.MkdirAll(costDir, 0o755); err != nil {
		return nil, fmt.Errorf("team/cost: mkdir %s: %w", costDir, err)
	}
	return &CostStore{costDir: costDir, budget: budget}, nil
}

// Record appends a new cost event to disk.
func (c *CostStore) Record(agentID, taskID, toolName, model string, input, output int) (*CostEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	total := input + output
	ev := &CostEvent{
		ID:           uuid.New().String(),
		AgentID:      agentID,
		TaskID:       taskID,
		ToolName:     toolName,
		Model:        model,
		InputTokens:  input,
		OutputTokens: output,
		TotalTokens:  total,
		OccurredAt:   time.Now(),
	}

	data, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("team/cost: marshal: %w", err)
	}

	ts := ev.OccurredAt.UTC().Format("20060102T150405")
	name := fmt.Sprintf("%s-%s-%s.json", agentID, ts, ev.ID[:8])
	tmp := filepath.Join(c.costDir, name+".tmp")
	dst := filepath.Join(c.costDir, name)

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return nil, fmt.Errorf("team/cost: write: %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return nil, fmt.Errorf("team/cost: rename: %w", err)
	}
	return ev, nil
}

// Summary returns aggregated totals, optionally filtered to one agent.
// Pass agentID="" for a team-wide summary.
func (c *CostStore) Summary(agentID string) (*CostSummary, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	events, err := c.listEvents(agentID)
	if err != nil {
		return nil, err
	}

	s := &CostSummary{
		AgentID:    agentID,
		Budget:     c.budget,
		EventCount: len(events),
	}
	for _, ev := range events {
		s.TotalInputTokens += ev.InputTokens
		s.TotalOutputTokens += ev.OutputTokens
		s.TotalTokens += ev.TotalTokens
	}
	if c.budget > 0 {
		s.BudgetUsedPercent = float64(s.TotalTokens) / float64(c.budget) * 100
	}
	return s, nil
}

// BudgetExceeded returns true if the team-wide total tokens exceed the budget.
func (c *CostStore) BudgetExceeded() (bool, error) {
	if c.budget <= 0 {
		return false, nil
	}
	s, err := c.Summary("")
	if err != nil {
		return false, err
	}
	return s.TotalTokens >= c.budget, nil
}

// SetBudget updates the budget limit at runtime.
func (c *CostStore) SetBudget(tokens int) {
	c.mu.Lock()
	c.budget = tokens
	c.mu.Unlock()
}

// Events returns all cost events, optionally filtered to one agent.
func (c *CostStore) Events(agentID string) ([]*CostEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.listEvents(agentID)
}

// listEvents reads all event files. Caller must hold c.mu.
func (c *CostStore) listEvents(agentID string) ([]*CostEvent, error) {
	entries, err := os.ReadDir(c.costDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/cost: readdir: %w", err)
	}

	var events []*CostEvent
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		// Fast path: filter by agentID prefix before parsing.
		if agentID != "" && len(e.Name()) > len(agentID) {
			if e.Name()[:len(agentID)] != agentID {
				continue
			}
		}
		data, err := os.ReadFile(filepath.Join(c.costDir, e.Name()))
		if err != nil {
			continue
		}
		var ev CostEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			continue
		}
		if agentID != "" && ev.AgentID != agentID {
			continue
		}
		events = append(events, &ev)
	}
	return events, nil
}
