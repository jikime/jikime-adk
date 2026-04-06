package teamcmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newInboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "Send and receive agent messages",
	}
	cmd.AddCommand(newInboxSendCmd())
	cmd.AddCommand(newInboxBroadcastCmd())
	cmd.AddCommand(newInboxReceiveCmd())
	cmd.AddCommand(newInboxPeekCmd())
	cmd.AddCommand(newInboxWatchCmd())
	cmd.AddCommand(newInboxLogCmd())
	return cmd
}

func newInboxSendCmd() *cobra.Command {
	var (
		from    string
		subject string
	)
	cmd := &cobra.Command{
		Use:   "send <team-name> <to-agent> <message>",
		Short: "Send a direct message to an agent",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" {
				from = os.Getenv("JIKIME_AGENT_ID")
			}
			if from == "" {
				from = "cli"
			}
			ti := team.NewTeamInbox(teamDir(args[0]))
			msg := &team.Message{
				TeamName: args[0],
				Kind:     team.MessageKindDirect,
				From:     from,
				To:       args[1],
				Subject:  subject,
				Body:     args[2],
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Message sent to %s\n", args[1])
			return nil
		},
	}
	cmd.Flags().StringVarP(&from, "from", "f", "", "Sender agent ID (default: JIKIME_AGENT_ID)")
	cmd.Flags().StringVarP(&subject, "subject", "s", "", "Message subject")
	return cmd
}

func newInboxBroadcastCmd() *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "broadcast <team-name> <message>",
		Short: "Broadcast a message to all team agents",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" {
				from = os.Getenv("JIKIME_AGENT_ID")
			}
			if from == "" {
				from = "cli"
			}
			td := teamDir(args[0])
			reg, err := team.NewRegistry(td + "/registry")
			if err != nil {
				return err
			}
			agents, _ := reg.List()
			agentIDs := make([]string, 0, len(agents))
			for _, a := range agents {
				agentIDs = append(agentIDs, a.ID)
			}
			ti := team.NewTeamInbox(td)
			msg := &team.Message{
				TeamName: args[0],
				Kind:     team.MessageKindBroadcast,
				From:     from,
				Body:     args[1],
				SentAt:   time.Now(),
			}
			if err := ti.Broadcast(msg, agentIDs); err != nil {
				return err
			}
			fmt.Printf("✅ Broadcast sent to %d agents\n", len(agentIDs)-1)
			return nil
		},
	}
	cmd.Flags().StringVarP(&from, "from", "f", "", "Sender agent ID")
	return cmd
}

func newInboxReceiveCmd() *cobra.Command {
	var (
		agentID string
		limit   int
	)
	cmd := &cobra.Command{
		Use:   "receive <team-name>",
		Short: "Receive and consume messages from inbox",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			ib, err := team.NewInbox(team.InboxDir(teamDir(args[0]), agentID))
			if err != nil {
				return err
			}
			msgs, err := ib.Receive(limit)
			if err != nil {
				return err
			}
			if len(msgs) == 0 {
				fmt.Println("No messages.")
				return nil
			}
			for _, m := range msgs {
				fmt.Printf("[%s] from:%s  %s\n  %s\n",
					m.SentAt.Format("15:04:05"), m.From, m.Subject, m.Body)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Max messages to receive")
	return cmd
}

func newInboxPeekCmd() *cobra.Command {
	var agentID string
	cmd := &cobra.Command{
		Use:   "peek <team-name>",
		Short: "Peek at inbox messages without consuming them",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			ib, err := team.NewInbox(team.InboxDir(teamDir(args[0]), agentID))
			if err != nil {
				return err
			}
			msgs, err := ib.Peek(0)
			if err != nil {
				return err
			}
			fmt.Printf("Inbox for %s (%d messages):\n", agentID, len(msgs))
			for _, m := range msgs {
				fmt.Printf("  [%s] from:%-12s  %s\n", m.SentAt.Format("15:04:05"), m.From, m.Body)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	return cmd
}

func newInboxWatchCmd() *cobra.Command {
	var (
		agentID string
		execCmd string
	)
	cmd := &cobra.Command{
		Use:   "watch <team-name>",
		Short: "Watch inbox for new messages in real-time (Ctrl+C to stop)",
		Long: `Watch inbox for new messages. With --exec, runs a shell command
for each incoming message. Message data is passed as environment variables:
  JIKIME_MSG_FROM, JIKIME_MSG_TO, JIKIME_MSG_SUBJECT, JIKIME_MSG_BODY,
  JIKIME_MSG_KIND, JIKIME_MSG_ID, JIKIME_MSG_TIME`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			ib, err := team.NewInbox(team.InboxDir(teamDir(args[0]), agentID))
			if err != nil {
				return err
			}
			fmt.Printf("Watching inbox for %s… (Ctrl+C to stop)\n", agentID)
			done := make(chan struct{})
			return ib.Watch(done, func(m *team.Message) {
				fmt.Printf("[%s] from:%-12s  %s\n", m.SentAt.Format("15:04:05"), m.From, m.Body)
				if execCmd != "" {
					// Security: use env vars for message data instead of shell interpolation.
					// WARNING: Do NOT use $JIKIME_MSG_BODY directly in shell expansions
					// (e.g. echo $JIKIME_MSG_BODY). Use quoted form: echo "$JIKIME_MSG_BODY"
					// to prevent shell injection from malicious message content.
					c := exec.Command("sh", "-c", execCmd)
					c.Env = append(os.Environ(),
						"JIKIME_MSG_FROM="+m.From,
						"JIKIME_MSG_TO="+m.To,
						"JIKIME_MSG_SUBJECT="+m.Subject,
						"JIKIME_MSG_BODY="+m.Body,
						"JIKIME_MSG_KIND="+string(m.Kind),
						"JIKIME_MSG_ID="+m.ID,
						"JIKIME_MSG_TIME="+m.SentAt.Format(time.RFC3339),
					)
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					_ = c.Run()
				}
			})
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	cmd.Flags().StringVarP(&execCmd, "exec", "e", "", "Shell command for each message (use quoted $JIKIME_MSG_BODY to avoid injection)")
	return cmd
}

func newInboxLogCmd() *cobra.Command {
	var (
		limit   int
		from    string
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "log <team-name>",
		Short: "View team message history (non-destructive event log)",
		Long: `Show all messages that have been sent in the team, oldest first.
Unlike 'receive', this does not consume messages — it reads the event log.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ti := team.NewTeamInbox(teamDir(args[0]))
			msgs, err := ti.EventLog(limit, from)
			if err != nil {
				return err
			}
			if len(msgs) == 0 {
				fmt.Println("No messages in event log.")
				return nil
			}

			if jsonOut {
				return printJSONList(msgs)
			}

			fmt.Printf("Message history: %d message(s)\n\n", len(msgs))
			for _, m := range msgs {
				to := m.To
				if to == "" {
					to = "all"
				}
				kind := string(m.Kind)
				fmt.Printf("  [%s]  %-12s → %-12s  (%s)  %s\n",
					m.SentAt.Format("01-02 15:04:05"),
					m.From, to, kind,
					truncate(m.Body, 80))
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Max messages to show (0 = all)")
	cmd.Flags().StringVarP(&from, "from", "f", "", "Filter by sender agent ID")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
