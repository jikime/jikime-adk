package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

// TeamPlanGateCmd fires on UserPromptSubmit when JIKIME_TEAM_NAME is set.
// If the prompt contains a plan submission marker, it waits for leader approval.
var TeamPlanGateCmd = &cobra.Command{
	Use:   "team-plan-gate",
	Short: "Gate execution until plan is approved by team leader",
	Long: `Team-aware UserPromptSubmit hook that detects plan submissions.
When the agent submits a prompt containing [PLAN_SUBMIT], it writes the plan
to the PlanStore and blocks until the leader approves or rejects it.
Only activates when JIKIME_TEAM_NAME and JIKIME_PLAN_GATE=1 are set.`,
	RunE: runTeamPlanGate,
}

func runTeamPlanGate(cmd *cobra.Command, args []string) error {
	teamName := os.Getenv("JIKIME_TEAM_NAME")
	if teamName == "" {
		return writeResponse(HookResponse{Continue: true})
	}
	if os.Getenv("JIKIME_PLAN_GATE") != "1" {
		// Plan gate not enabled for this agent.
		return writeResponse(HookResponse{Continue: true})
	}

	agentID := os.Getenv("JIKIME_AGENT_ID")
	dataDir := os.Getenv("JIKIME_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".jikime")
	}

	var input userPromptInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Only gate prompts that start with the plan-submit marker.
	if len(input.Prompt) < 14 || input.Prompt[:14] != "[PLAN_SUBMIT] " {
		return writeResponse(HookResponse{Continue: true})
	}

	planBody := input.Prompt[14:] // strip marker

	td := filepath.Join(dataDir, "teams", teamName)
	plansDir := filepath.Join(dataDir, "plans")

	planStore, err := team.NewPlanStore(plansDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[jikime/team] plan store: %v\n", err)
		return writeResponse(HookResponse{Continue: true})
	}

	plan, err := planStore.Submit(teamName, agentID, "Agent Plan", planBody, nil)
	if err != nil {
		return writeResponse(HookResponse{
			Continue:      false,
			SystemMessage: fmt.Sprintf("Failed to submit plan: %v", err),
		})
	}

	// Notify leader via inbox.
	ti := team.NewTeamInbox(td)
	_ = ti.Send(&team.Message{
		TeamName: teamName,
		Kind:     team.MessageKindDirect,
		From:     agentID,
		To:       "leader",
		Subject:  "plan_review_required",
		Body:     fmt.Sprintf("Plan %s submitted by %s — review with: jikime team plan approve %s %s", plan.ID[:8], agentID, teamName, plan.ID[:8]),
		SentAt:   time.Now(),
	})

	fmt.Fprintf(os.Stderr, "[jikime/team] plan %s submitted, waiting for approval…\n", plan.ID[:8])

	// Poll for approval (max 10 minutes, check every 5 seconds).
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		p, err := planStore.Get(plan.ID)
		if err != nil || p == nil {
			break
		}
		switch p.Status {
		case team.PlanStatusApproved:
			fmt.Fprintf(os.Stderr, "[jikime/team] plan %s approved\n", plan.ID[:8])
			return writeResponse(HookResponse{
				Continue:      true,
				SystemMessage: fmt.Sprintf("✅ Plan approved by %s. Proceeding.", p.ReviewedBy),
			})
		case team.PlanStatusRejected:
			msg := fmt.Sprintf("❌ Plan rejected by %s: %s", p.ReviewedBy, p.RejectionReason)
			fmt.Fprintf(os.Stderr, "[jikime/team] %s\n", msg)
			return writeResponse(HookResponse{
				Continue:      false,
				SystemMessage: msg,
			})
		}
		time.Sleep(5 * time.Second)
	}

	return writeResponse(HookResponse{
		Continue:      false,
		SystemMessage: "⏰ Plan approval timed out (10 minutes). Stopping agent.",
	})
}
