package routercmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	router "jikime-adk/internal/router"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running router proxy",
		RunE:  runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	pid := readPID()
	if pid == 0 {
		return fmt.Errorf("no router running (PID file not found)")
	}

	if !processExists(pid) {
		os.Remove(router.PIDPath())
		return fmt.Errorf("router process (pid: %d) not found, cleaned up stale PID file", pid)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop router: %w", err)
	}

	os.Remove(router.PIDPath())
	color.Green("  Router stopped (pid: %d)", pid)
	return nil
}
