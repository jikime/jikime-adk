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
	addr := fmt.Sprintf("http://%s:%d", cfg.Router.Host, cfg.Router.Port)
	fmt.Printf("  Address:   %s\n", cyan(addr))

	// List available providers with their models
	providers := cfg.GetProviderNames()
	fmt.Printf("  Providers: %s\n", cyan(fmt.Sprintf("%v", providers)))
	for _, name := range providers {
		if p, ok := cfg.Providers[name]; ok {
			fmt.Printf("    - %s: %s\n", cyan(name), p.Model)
		}
	}

	// Show Claude Code integration status
	if hasManagedEnv() {
		color.Green("  Claude:   configured (.claude/settings.local.json)")
	} else {
		fmt.Printf("  Claude:   not configured\n")
	}
	fmt.Println()

	return nil
}
