package teamcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Submit and review agent plans",
	}
	cmd.AddCommand(newPlanSubmitCmd())
	cmd.AddCommand(newPlanApproveCmd())
	cmd.AddCommand(newPlanRejectCmd())
	cmd.AddCommand(newPlanListCmd())
	return cmd
}

func newPlanSubmitCmd() *cobra.Command {
	var (
		title  string
		body   string
		file   string
		agent  string
	)

	cmd := &cobra.Command{
		Use:   "submit <team-name>",
		Short: "Submit a plan for leader approval",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]

			if agent == "" {
				agent = os.Getenv("JIKIME_AGENT_ID")
			}
			if agent == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}

			planBody := body
			if file != "" {
				data, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("read plan file: %w", err)
				}
				planBody = string(data)
			}

			ps, err := team.NewPlanStore(filepath.Join(dataDir(), "plans"))
			if err != nil {
				return err
			}
			plan, err := ps.Submit(teamName, agent, title, planBody, nil)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Plan submitted: %s\n", plan.ID)
			fmt.Printf("   title:  %s\n", plan.Title)
			fmt.Printf("   status: %s\n", plan.Status)
			return nil
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "Plan", "Plan title")
	cmd.Flags().StringVarP(&body, "body", "b", "", "Plan body text")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Read plan body from file")
	cmd.Flags().StringVarP(&agent, "agent", "a", "", "Agent ID submitting the plan")
	return cmd
}

func newPlanApproveCmd() *cobra.Command {
	var reviewer string

	cmd := &cobra.Command{
		Use:   "approve <plan-id>",
		Short: "Approve a pending plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if reviewer == "" {
				reviewer = os.Getenv("JIKIME_AGENT_ID")
			}
			if reviewer == "" {
				reviewer = "leader"
			}
			ps, err := team.NewPlanStore(filepath.Join(dataDir(), "plans"))
			if err != nil {
				return err
			}
			plan, err := ps.Approve(args[0], reviewer)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Plan %s approved by %s\n", plan.ID[:8], reviewer)
			return nil
		},
	}
	cmd.Flags().StringVarP(&reviewer, "reviewer", "r", "", "Reviewer agent ID (default: JIKIME_AGENT_ID)")
	return cmd
}

func newPlanRejectCmd() *cobra.Command {
	var (
		reviewer string
		reason   string
	)

	cmd := &cobra.Command{
		Use:   "reject <plan-id>",
		Short: "Reject a pending plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if reviewer == "" {
				reviewer = os.Getenv("JIKIME_AGENT_ID")
			}
			if reviewer == "" {
				reviewer = "leader"
			}
			ps, err := team.NewPlanStore(filepath.Join(dataDir(), "plans"))
			if err != nil {
				return err
			}
			plan, err := ps.Reject(args[0], reviewer, reason)
			if err != nil {
				return err
			}
			fmt.Printf("❌ Plan %s rejected by %s\n", plan.ID[:8], reviewer)
			return nil
		},
	}
	cmd.Flags().StringVarP(&reviewer, "reviewer", "r", "", "Reviewer agent ID")
	cmd.Flags().StringVar(&reason, "reason", "", "Rejection reason")
	return cmd
}

func newPlanListCmd() *cobra.Command {
	var teamFilter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			ps, err := team.NewPlanStore(filepath.Join(dataDir(), "plans"))
			if err != nil {
				return err
			}
			plans, err := ps.List(teamFilter, "")
			if err != nil {
				return err
			}
			if len(plans) == 0 {
				fmt.Println("No plans.")
				return nil
			}
			for _, p := range plans {
				fmt.Printf("  %s  [%s]  %s  by:%s\n",
					p.ID[:8], p.Status, p.Title, p.SubmittedBy)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&teamFilter, "team", "t", "", "Filter by team name")
	return cmd
}
