package team

import (
	"fmt"
	"strings"
)

// PromptConfig holds all parameters needed to build an agent's initial prompt.
type PromptConfig struct {
	// TeamName is the team this agent belongs to.
	TeamName string

	// AgentID is the unique identifier for this agent (e.g. "worker-1").
	AgentID string

	// Role is the agent's functional role: "leader", "worker", "reviewer".
	Role string

	// LeaderID is the ID of the leader agent (empty if this agent IS the leader).
	LeaderID string

	// Workers is the list of worker agent IDs (used in leader prompt).
	Workers []string

	// Goal is the high-level objective for the team.
	Goal string

	// TaskBody is the custom task prompt from the template's task field.
	// If empty, a default prompt is generated based on Role.
	TaskBody string

	// WorktreePath is the git worktree path for this agent (optional).
	WorktreePath string
}

// BuildAgentPrompt constructs the full initial prompt for an agent.
// It combines identity, optional workspace info, task, and coordination protocol.
func BuildAgentPrompt(cfg PromptConfig) string {
	var b strings.Builder

	// --- Identity ---
	b.WriteString("## Identity\n\n")
	b.WriteString(fmt.Sprintf("- Agent: %s\n", cfg.AgentID))
	b.WriteString(fmt.Sprintf("- Role: %s\n", cfg.Role))
	b.WriteString(fmt.Sprintf("- Team: %s\n", cfg.TeamName))
	if cfg.LeaderID != "" {
		b.WriteString(fmt.Sprintf("- Leader: %s\n", cfg.LeaderID))
	}

	// --- Workspace (optional) ---
	if cfg.WorktreePath != "" {
		b.WriteString("\n## Workspace\n\n")
		b.WriteString(fmt.Sprintf("- Directory: %s\n", cfg.WorktreePath))
		b.WriteString("- Isolated git worktree — your changes do not affect the main branch.\n")
		b.WriteString("- Commit your work before marking a task complete.\n")
	}

	// --- Task ---
	b.WriteString("\n## Task\n\n")
	if cfg.TaskBody != "" {
		b.WriteString(cfg.TaskBody)
		b.WriteString("\n")
	} else {
		b.WriteString(defaultTaskBody(cfg))
	}

	// --- Coordination Protocol ---
	b.WriteString("\n## Coordination Protocol\n\n")
	b.WriteString(coordinationProtocol(cfg))

	return b.String()
}

// defaultTaskBody returns a role-appropriate task description when the template
// does not define a custom task prompt.
func defaultTaskBody(cfg PromptConfig) string {
	goal := cfg.Goal
	if goal == "" {
		goal = "(no goal specified)"
	}
	team := cfg.TeamName
	agent := cfg.AgentID

	switch cfg.Role {
	case "leader":
		workers := strings.Join(cfg.Workers, ", ")
		if workers == "" {
			workers = "(none)"
		}
		return fmt.Sprintf(`Goal: %s

Your responsibilities as team leader:
1. Analyze the goal and break it into concrete, parallelizable tasks.
2. Create a task for each unit of work:
   jikime team tasks create %s "Task title" --desc "What to do" --dod "Done when..."
3. Workers will claim tasks automatically. Monitor progress:
   jikime team status %s
4. Respond to worker messages in your inbox:
   jikime team inbox receive %s
5. When ALL tasks are done, YOU must perform final integration:
   a. Review all files created/modified by workers.
   b. Fix any integration issues (missing imports, broken references, type errors).
   c. Run the build or test command to verify everything compiles and works.
   d. Fix any build errors you find.
   e. Commit the final integrated result with a clear summary commit message.
6. After successful integration, shut down the team:
   jikime team inbox broadcast %s "Integration complete. Shutting down."
   jikime team lifecycle shutdown %s

Available workers: %s
`, goal, team, team, team, team, team, workers)

	case "worker":
		leaderID := cfg.LeaderID
		if leaderID == "" {
			leaderID = "leader"
		}
		return fmt.Sprintf(`Goal: %s

Your responsibilities as worker:
Loop until no pending tasks remain:
1. Check for pending tasks:
   jikime team tasks list %s --status pending
2. Claim a task:
   jikime team tasks claim %s <task-id> --agent %s
3. Read the task details and implement the work.
4. Mark the task complete with a brief result summary:
   jikime team tasks complete %s <task-id> --result "What was done"
5. Notify the leader:
   jikime team inbox send %s %s "Completed <task-id>: <one-line summary>"
6. Repeat from step 1.

When no tasks remain, notify the leader you are idle:
   jikime team inbox send %s %s "No more pending tasks. Idle."
`, goal, team, team, agent, team, team, leaderID, team, leaderID)

	case "reviewer":
		leaderID := cfg.LeaderID
		if leaderID == "" {
			leaderID = "leader"
		}
		return fmt.Sprintf(`Goal: %s

Your responsibilities as reviewer:
1. Check for tasks pending review (status: done):
   jikime team tasks list %s --status done
2. Review each completed task for quality and correctness.
3. For approved work, notify the leader:
   jikime team inbox send %s %s "Approved task <id>: <notes>"
4. For work needing revision, update the task status back to pending and notify:
   jikime team tasks update %s <task-id> --status pending
   jikime team inbox send %s %s "Revision needed for <id>: <what to fix>"
5. Repeat until all tasks are reviewed.
`, goal, team, team, leaderID, team, team, leaderID)

	default:
		return fmt.Sprintf("Goal: %s\n\nCarry out your assigned work for team %s.\n", goal, team)
	}
}

// coordinationProtocol returns the CLI command reference for an agent.
func coordinationProtocol(cfg PromptConfig) string {
	team := cfg.TeamName
	agent := cfg.AgentID
	leader := cfg.LeaderID
	if leader == "" {
		leader = "leader"
	}

	var lines []string
	lines = append(lines, "Use these commands to coordinate with the team:\n")

	switch cfg.Role {
	case "leader":
		lines = append(lines,
			fmt.Sprintf("  jikime team tasks create %s \"title\" --desc \"...\" --dod \"...\"   # create task", team),
			fmt.Sprintf("  jikime team tasks list %s                                          # view all tasks", team),
			fmt.Sprintf("  jikime team status %s                                              # team overview", team),
			fmt.Sprintf("  jikime team inbox receive %s                                       # read messages", team),
			fmt.Sprintf("  jikime team inbox broadcast %s \"message\"                          # message all agents", team),
			fmt.Sprintf("  jikime team lifecycle shutdown %s                                  # shut down team", team),
		)
	case "worker":
		lines = append(lines,
			fmt.Sprintf("  jikime team tasks list %s --status pending                         # find available tasks", team),
			fmt.Sprintf("  jikime team tasks claim %s <id> --agent %s              # claim a task", team, agent),
			fmt.Sprintf("  jikime team tasks complete %s <id> --result \"summary\"             # mark task done", team),
			fmt.Sprintf("  jikime team inbox send %s %s \"message\"                # message leader", team, leader),
			fmt.Sprintf("  jikime team inbox receive %s                                       # check your inbox", team),
		)
	case "reviewer":
		lines = append(lines,
			fmt.Sprintf("  jikime team tasks list %s --status done                            # find tasks to review", team),
			fmt.Sprintf("  jikime team tasks update %s <id> --status pending                 # send back for revision", team),
			fmt.Sprintf("  jikime team inbox send %s %s \"message\"                # message leader", team, leader),
			fmt.Sprintf("  jikime team inbox receive %s                                       # check your inbox", team),
		)
	default:
		lines = append(lines,
			fmt.Sprintf("  jikime team tasks list %s                                          # view tasks", team),
			fmt.Sprintf("  jikime team inbox receive %s                                       # check inbox", team),
		)
	}

	return strings.Join(lines, "\n") + "\n"
}
