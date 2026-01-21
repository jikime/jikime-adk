package hooks

import (
	"fmt"
	"strings"
)

// DDDCommitPhase represents DDD commit phase information
type DDDCommitPhase struct {
	PhaseName   string // "ANALYZE", "PRESERVE", or "IMPROVE"
	CommitType  string // "chore", "test", "refactor"
	Description string
	TestStatus  string // "analyzing", "preserving", or "improving"
}

// DDDPhaseInfo contains information about a DDD phase
type DDDPhaseInfo struct {
	CommitType  string
	Description string
	TestStatus  string
}

// DDDPhases defines the three phases of DDD development
var DDDPhases = map[string]DDDPhaseInfo{
	"ANALYZE": {
		CommitType:  "chore",
		Description: "Analyze existing code and behavior",
		TestStatus:  "analyzing",
	},
	"PRESERVE": {
		CommitType:  "test",
		Description: "Create characterization tests",
		TestStatus:  "preserving",
	},
	"IMPROVE": {
		CommitType:  "refactor",
		Description: "Refactor with behavior preservation",
		TestStatus:  "improving",
	},
}

// FormatDDDCommit formats a DDD phase commit message
func FormatDDDCommit(commitType, scope, subject, phase string) string {
	baseMsg := fmt.Sprintf("%s(%s): %s", commitType, scope, subject)

	phaseIndicators := map[string]string{
		"ANALYZE":  "(ANALYZE phase)",
		"PRESERVE": "(PRESERVE phase)",
		"IMPROVE":  "(IMPROVE phase)",
	}

	indicator, ok := phaseIndicators[phase]
	if !ok {
		indicator = fmt.Sprintf("(%s)", phase)
	}

	return fmt.Sprintf("%s %s", baseMsg, indicator)
}

// GetWorkflowCommands returns workflow commands for a given strategy
func GetWorkflowCommands(strategy, specID string) []string {
	var commands []string

	switch strategy {
	case "feature_branch":
		commands = []string{
			fmt.Sprintf("git switch -c feature/%s", specID),
			"# ANALYZE phase: git commit -m 'chore(...): analyze existing code'",
			"# PRESERVE phase: git commit -m 'test(...): add characterization tests'",
			"# IMPROVE phase: git commit -m 'refactor(...): improve with behavior preservation'",
			"gh pr create --base develop --generate-description",
			fmt.Sprintf("gh pr merge %s --auto --squash", specID),
		}
	case "direct_commit":
		commands = []string{
			"git switch develop",
			"# ANALYZE phase: git commit -m 'chore(...): analyze existing code'",
			"# PRESERVE phase: git commit -m 'test(...): add characterization tests'",
			"# IMPROVE phase: git commit -m 'refactor(...): improve with behavior preservation'",
			"git push origin develop",
		}
	}

	return commands
}

// GetDDDPhase returns phase information for a given phase name
func GetDDDPhase(phaseName string) (DDDPhaseInfo, bool) {
	info, ok := DDDPhases[phaseName]
	return info, ok
}

// CreateDDDCommitPhase creates a DDDCommitPhase from phase name
func CreateDDDCommitPhase(phaseName string) DDDCommitPhase {
	info, ok := DDDPhases[phaseName]
	if !ok {
		return DDDCommitPhase{
			PhaseName:   phaseName,
			CommitType:  "chore",
			Description: "Unknown phase",
			TestStatus:  "unknown",
		}
	}

	return DDDCommitPhase{
		PhaseName:   phaseName,
		CommitType:  info.CommitType,
		Description: info.Description,
		TestStatus:  info.TestStatus,
	}
}

// GitInfo holds git repository information
type GitInfo struct {
	Branch     string
	Changes    int
	Staged     int
	Modified   int
	Untracked  int
	LastCommit string
	IsRepo     bool
}

// GetGitInfo collects git repository information
func GetGitInfo() GitInfo {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return GitInfo{IsRepo: false}
	}
	
	return GetGitInfoForDir(projectRoot)
}

// GetGitInfoForDir collects git info for a specific directory
func GetGitInfoForDir(dir string) GitInfo {
	info := GitInfo{IsRepo: false}
	
	// Check if git repo
	_, err := RunCommandInDir(dir, "git", "rev-parse", "--git-dir")
	if err != nil {
		return info
	}
	info.IsRepo = true
	
	// Get current branch
	branch, err := RunCommandInDir(dir, "git", "branch", "--show-current")
	if err == nil && branch != "" {
		info.Branch = branch
	} else {
		// Try to get HEAD ref (for detached HEAD)
		head, err := RunCommandInDir(dir, "git", "rev-parse", "--short", "HEAD")
		if err == nil {
			info.Branch = "HEAD@" + head
		} else {
			info.Branch = "unknown"
		}
	}
	
	// Get status counts
	status, err := RunCommandInDir(dir, "git", "status", "--porcelain")
	if err == nil && status != "" {
		lines := strings.Split(status, "\n")
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			x, y := line[0], line[1]
			
			// Staged changes (index)
			if x == 'A' || x == 'M' || x == 'D' || x == 'R' || x == 'C' {
				info.Staged++
			}
			
			// Modified in working tree
			if y == 'M' || y == 'D' {
				info.Modified++
			}
			
			// Untracked files
			if x == '?' && y == '?' {
				info.Untracked++
			}
		}
		info.Changes = info.Staged + info.Modified + info.Untracked
	}
	
	// Get last commit
	lastCommit, err := RunCommandInDir(dir, "git", "log", "-1", "--pretty=format:%h - %s (%ar)")
	if err == nil && lastCommit != "" {
		info.LastCommit = lastCommit
	} else {
		info.LastCommit = "No commits yet"
	}
	
	return info
}

// GetChangesDisplay returns a formatted string of changes
func (g GitInfo) GetChangesDisplay() string {
	if !g.IsRepo {
		return "Not a git repo"
	}
	if g.Changes == 0 {
		return "No changes"
	}
	return fmt.Sprintf("%d file(s) modified", g.Changes)
}

// GetStatusDisplay returns a formatted status string
func (g GitInfo) GetStatusDisplay() string {
	if !g.IsRepo {
		return ""
	}
	return fmt.Sprintf("+%d M%d ?%d", g.Staged, g.Modified, g.Untracked)
}
