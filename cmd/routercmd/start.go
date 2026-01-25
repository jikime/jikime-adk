package routercmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	router "jikime-adk/internal/router"
)

var (
	startPort   int
	startDaemon bool
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the LLM router proxy",
		RunE:  runStart,
	}

	cmd.Flags().IntVarP(&startPort, "port", "p", 0, "Override port (default from config)")
	cmd.Flags().BoolVarP(&startDaemon, "daemon", "d", false, "Run in background")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := router.LoadConfig()
	if err != nil {
		return err
	}

	// Override port if specified
	if startPort > 0 {
		cfg.Router.Port = startPort
	}

	// Check if already running
	if pid := readPID(); pid > 0 {
		if processExists(pid) {
			return fmt.Errorf("router already running (pid: %d). Use 'jikime router stop' first", pid)
		}
		// Stale PID file, remove it
		os.Remove(router.PIDPath())
	}

	if startDaemon {
		return startDaemonProcess(cfg)
	}

	return startForeground(cfg)
}

func startForeground(cfg *router.Config) error {
	printStartInfo(cfg)

	srv, err := router.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	// Write PID
	writePID(os.Getpid())
	defer os.Remove(router.PIDPath())

	return srv.Start()
}

func startDaemonProcess(cfg *router.Config) error {
	// Re-launch self without --daemon flag
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	args := []string{"router", "start"}
	if startPort > 0 {
		args = append(args, "--port", strconv.Itoa(startPort))
	}

	proc := exec.Command(exe, args...)
	proc.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// Redirect output to log file
	logPath := router.PIDPath() + ".log"
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	proc.Stdout = logFile
	proc.Stderr = logFile

	if err := proc.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start daemon: %w", err)
	}

	logFile.Close()
	printStartInfo(cfg)
	color.Green("  Router started in background (pid: %d)", proc.Process.Pid)
	fmt.Printf("  Log: %s\n", logPath)

	return nil
}

func printStartInfo(cfg *router.Config) {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println()
	fmt.Printf("  %s Router\n", green("LLM"))
	fmt.Printf("  Address:   %s\n", cyan(fmt.Sprintf("http://%s:%d", cfg.Router.Host, cfg.Router.Port)))

	// List available providers
	providers := cfg.GetProviderNames()
	if len(providers) > 0 {
		fmt.Printf("  Providers: %s\n", cyan(fmt.Sprintf("%v", providers)))
	}
	fmt.Println()
}

// --- PID file helpers ---

func writePID(pid int) {
	dir := router.PIDPath()
	os.MkdirAll(dir[:len(dir)-len("/router.pid")], 0o755)
	os.WriteFile(router.PIDPath(), []byte(strconv.Itoa(pid)), 0o644)
}

func readPID() int {
	data, err := os.ReadFile(router.PIDPath())
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(string(data))
	return pid
}

func processExists(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
