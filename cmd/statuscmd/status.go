// Package statuscmd provides the status command for jikime-adk.
package statuscmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Config represents the jikime configuration structure
type Config struct {
	User struct {
		Name string `yaml:"name"`
	} `yaml:"user"`
	Language struct {
		ConversationLanguage     string `yaml:"conversation_language"`
		ConversationLanguageName string `yaml:"conversation_language_name"`
		AgentPromptLanguage      string `yaml:"agent_prompt_language"`
		GitCommitMessages        string `yaml:"git_commit_messages"`
		CodeComments             string `yaml:"code_comments"`
		Documentation            string `yaml:"documentation"`
		ErrorMessages            string `yaml:"error_messages"`
	} `yaml:"language"`
	Jikime struct {
		Version string `yaml:"version"`
	} `yaml:"jikime"`
}

// NewStatus creates the status command.
func NewStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status and configuration",
		Long:  "Display current project status including configuration, SPEC count, and Git information.",
		RunE:  runStatus,
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if this is a jikime project
	jikimeDir := filepath.Join(cwd, ".jikime")
	claudeDir := filepath.Join(cwd, ".claude")
	if _, err := os.Stat(jikimeDir); os.IsNotExist(err) {
		if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
			return fmt.Errorf("not a Jikime project. Run 'jikime-adk init' first")
		}
	}

	// Load configuration
	config, err := loadConfig(cwd)
	if err != nil {
		// Continue without config
		config = &Config{}
	}

	// Print header
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	cyan.Println("╔════════════════════════════════════════╗")
	cyan.Println("║         Jikime-ADK Status              ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Project info
	projectName := filepath.Base(cwd)
	fmt.Printf("%-20s %s\n", "Project:", green.Sprint(projectName))
	fmt.Printf("%-20s %s\n", "Path:", dim.Sprint(cwd))
	fmt.Println()

	// Configuration section
	yellow.Println("Configuration")
	yellow.Println("─────────────")
	if config.User.Name != "" {
		fmt.Printf("%-20s %s\n", "User:", config.User.Name)
	}
	if config.Language.ConversationLanguage != "" {
		fmt.Printf("%-20s %s (%s)\n", "Language:",
			config.Language.ConversationLanguage,
			config.Language.ConversationLanguageName)
	}
	if config.Jikime.Version != "" {
		fmt.Printf("%-20s %s\n", "Jikime Version:", config.Jikime.Version)
	}
	fmt.Println()

	// SPEC count
	specsDir := filepath.Join(cwd, ".jikime", "specs")
	specCount := countSpecs(specsDir)
	yellow.Println("SPECs")
	yellow.Println("─────")
	fmt.Printf("%-20s %d\n", "Active SPECs:", specCount)
	fmt.Println()

	// Git information
	yellow.Println("Git")
	yellow.Println("───")
	printGitInfo(cwd)

	return nil
}

func loadConfig(cwd string) (*Config, error) {
	config := &Config{}

	// Try loading from sections first
	sectionsDir := filepath.Join(cwd, ".jikime", "config", "sections")

	// Load user.yaml
	userPath := filepath.Join(sectionsDir, "user.yaml")
	if data, err := os.ReadFile(userPath); err == nil {
		var userConfig struct {
			User struct {
				Name string `yaml:"name"`
			} `yaml:"user"`
		}
		if err := yaml.Unmarshal(data, &userConfig); err == nil {
			config.User = userConfig.User
		}
	}

	// Load language.yaml
	langPath := filepath.Join(sectionsDir, "language.yaml")
	if data, err := os.ReadFile(langPath); err == nil {
		var langConfig struct {
			Language struct {
				ConversationLanguage     string `yaml:"conversation_language"`
				ConversationLanguageName string `yaml:"conversation_language_name"`
				AgentPromptLanguage      string `yaml:"agent_prompt_language"`
				GitCommitMessages        string `yaml:"git_commit_messages"`
				CodeComments             string `yaml:"code_comments"`
				Documentation            string `yaml:"documentation"`
				ErrorMessages            string `yaml:"error_messages"`
			} `yaml:"language"`
		}
		if err := yaml.Unmarshal(data, &langConfig); err == nil {
			config.Language = langConfig.Language
		}
	}

	// Load system.yaml for version
	systemPath := filepath.Join(sectionsDir, "system.yaml")
	if data, err := os.ReadFile(systemPath); err == nil {
		var systemConfig struct {
			Jikime struct {
				Version string `yaml:"version"`
			} `yaml:"jikime"`
		}
		if err := yaml.Unmarshal(data, &systemConfig); err == nil {
			config.Jikime = systemConfig.Jikime
		}
	}

	// Fallback to main config.yaml if sections not found
	configPath := filepath.Join(cwd, ".jikime", "config", "config.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func countSpecs(specsDir string) int {
	count := 0
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "SPEC-") {
			specFile := filepath.Join(specsDir, entry.Name(), "spec.md")
			if _, err := os.Stat(specFile); err == nil {
				count++
			}
		}
	}

	return count
}

func printGitInfo(cwd string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	dim := color.New(color.Faint)

	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		dim.Println("Git not available")
		return
	}

	// Check if this is a git repository
	gitCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	gitCmd.Dir = cwd
	if err := gitCmd.Run(); err != nil {
		dim.Println("Not a git repository")
		return
	}

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = cwd
	branchOutput, err := branchCmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(branchOutput))
		fmt.Printf("%-20s %s\n", "Branch:", green.Sprint(branch))
	}

	// Get status (modified/staged files count)
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = cwd
	statusOutput, err := statusCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
		modified := 0
		staged := 0
		untracked := 0

		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			indexStatus := line[0]
			workTreeStatus := line[1]

			if indexStatus == '?' {
				untracked++
			} else {
				if indexStatus != ' ' && indexStatus != '?' {
					staged++
				}
				if workTreeStatus != ' ' && workTreeStatus != '?' {
					modified++
				}
			}
		}

		if staged > 0 || modified > 0 || untracked > 0 {
			statusParts := []string{}
			if staged > 0 {
				statusParts = append(statusParts, green.Sprintf("%d staged", staged))
			}
			if modified > 0 {
				statusParts = append(statusParts, red.Sprintf("%d modified", modified))
			}
			if untracked > 0 {
				statusParts = append(statusParts, dim.Sprintf("%d untracked", untracked))
			}
			fmt.Printf("%-20s %s\n", "Changes:", strings.Join(statusParts, ", "))
		} else {
			fmt.Printf("%-20s %s\n", "Changes:", green.Sprint("Clean"))
		}
	}

	// Get last commit info
	logCmd := exec.Command("git", "log", "-1", "--format=%h %s", "--date=relative")
	logCmd.Dir = cwd
	logOutput, err := logCmd.Output()
	if err == nil {
		lastCommit := strings.TrimSpace(string(logOutput))
		if len(lastCommit) > 60 {
			lastCommit = lastCommit[:57] + "..."
		}
		fmt.Printf("%-20s %s\n", "Last commit:", dim.Sprint(lastCommit))
	}

	// Check remote status
	remoteCmd := exec.Command("git", "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	remoteCmd.Dir = cwd
	remoteOutput, err := remoteCmd.Output()
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(remoteOutput)))
		if len(parts) == 2 {
			behind := parts[0]
			ahead := parts[1]
			if behind != "0" || ahead != "0" {
				statusText := ""
				if ahead != "0" {
					statusText += green.Sprintf("↑%s ahead", ahead)
				}
				if behind != "0" {
					if statusText != "" {
						statusText += ", "
					}
					statusText += red.Sprintf("↓%s behind", behind)
				}
				fmt.Printf("%-20s %s\n", "Remote:", statusText)
			} else {
				fmt.Printf("%-20s %s\n", "Remote:", green.Sprint("Up to date"))
			}
		}
	}
}
