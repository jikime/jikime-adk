package routercmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	router "jikime-adk/internal/router"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show router proxy status",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	pid := readPID()

	fmt.Println()
	if pid > 0 && processExists(pid) {
		color.Green("  Router: running (pid: %d)", pid)
	} else {
		color.Red("  Router: stopped")
		if pid > 0 {
			os.Remove(router.PIDPath())
			fmt.Printf("  (stale PID file cleaned)\n")
		}
	}

	// Show config info
	cfg, err := router.LoadConfig()
	if err != nil {
		fmt.Printf("  Config: not found (%v)\n", err)
		fmt.Println()
		return nil
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("  Provider: %s\n", cyan(cfg.Router.Provider))

	if p, ok := cfg.Providers[cfg.Router.Provider]; ok {
		fmt.Printf("  Model:    %s\n", cyan(p.Model))
		if p.BaseURL != "" {
			fmt.Printf("  Base URL: %s\n", cyan(p.BaseURL))
		}
	}

	fmt.Printf("  Address:  %s\n", cyan(fmt.Sprintf("http://%s:%d", cfg.Router.Host, cfg.Router.Port)))

	// Show Claude Code integration status
	if hasManagedEnv() {
		color.Green("  Claude:   configured (.claude/settings.local.json)")
	} else {
		fmt.Printf("  Claude:   not configured\n")
	}
	fmt.Println()

	return nil
}
