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

// PlanStore manages plan submissions and approvals.
// Plans are stored at ~/.jikime/plans/<id>.json.
type PlanStore struct {
	mu      sync.Mutex
	planDir string
}

// NewPlanStore returns a PlanStore rooted at planDir.
func NewPlanStore(planDir string) (*PlanStore, error) {
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		return nil, fmt.Errorf("team/plan: mkdir %s: %w", planDir, err)
	}
	return &PlanStore{planDir: planDir}, nil
}

// Submit creates a new plan in pending state and returns it.
func (p *PlanStore) Submit(teamName, submittedBy, title, body string, tasks []Task) (*Plan, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	plan := &Plan{
		ID:          uuid.New().String(),
		TeamName:    teamName,
		SubmittedBy: submittedBy,
		Title:       title,
		Body:        body,
		Tasks:       tasks,
		Status:      PlanStatusPending,
		SubmittedAt: time.Now(),
	}
	return plan, p.save(plan)
}

// Get returns the plan with the given ID, or (nil, nil) if not found.
func (p *PlanStore) Get(planID string) (*Plan, error) {
	data, err := os.ReadFile(p.path(planID))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/plan: read %s: %w", planID, err)
	}
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("team/plan: unmarshal %s: %w", planID, err)
	}
	return &plan, nil
}

// List returns all plans, optionally filtered by teamName or submittedBy.
func (p *PlanStore) List(teamName, submittedBy string) ([]*Plan, error) {
	entries, err := os.ReadDir(p.planDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/plan: readdir: %w", err)
	}
	var plans []*Plan
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		plan, err := p.Get(id)
		if err != nil || plan == nil {
			continue
		}
		if teamName != "" && plan.TeamName != teamName {
			continue
		}
		if submittedBy != "" && plan.SubmittedBy != submittedBy {
			continue
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// Approve transitions a plan to approved status.
func (p *PlanStore) Approve(planID, reviewedBy string) (*Plan, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	plan, err := p.Get(planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("team/plan: plan %s not found", planID)
	}
	if plan.Status != PlanStatusPending {
		return nil, fmt.Errorf("team/plan: plan %s is not pending (status: %s)", planID, plan.Status)
	}

	now := time.Now()
	plan.Status = PlanStatusApproved
	plan.ReviewedBy = reviewedBy
	plan.ReviewedAt = &now
	return plan, p.save(plan)
}

// Reject transitions a plan to rejected status with an optional reason.
func (p *PlanStore) Reject(planID, reviewedBy, reason string) (*Plan, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	plan, err := p.Get(planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("team/plan: plan %s not found", planID)
	}
	if plan.Status != PlanStatusPending {
		return nil, fmt.Errorf("team/plan: plan %s is not pending (status: %s)", planID, plan.Status)
	}

	now := time.Now()
	plan.Status = PlanStatusRejected
	plan.ReviewedBy = reviewedBy
	plan.RejectionReason = reason
	plan.ReviewedAt = &now
	return plan, p.save(plan)
}

// Delete removes a plan file.
func (p *PlanStore) Delete(planID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := os.Remove(p.path(planID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (p *PlanStore) save(plan *Plan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("team/plan: marshal %s: %w", plan.ID, err)
	}
	tmp := p.path(plan.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("team/plan: write %s: %w", plan.ID, err)
	}
	if err := os.Rename(tmp, p.path(plan.ID)); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("team/plan: rename %s: %w", plan.ID, err)
	}
	return nil
}

func (p *PlanStore) path(planID string) string {
	return filepath.Join(p.planDir, planID+".json")
}
