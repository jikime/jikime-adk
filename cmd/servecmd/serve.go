// Package servecmd implements the `jikime serve` CLI command.
// Runs the Symphony-inspired autonomous agent orchestration service.
package servecmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"jikime-adk/internal/serve"
	"jikime-adk/internal/serve/agent"
	"jikime-adk/internal/serve/orchestrator"
	"jikime-adk/internal/serve/tracker"
	"jikime-adk/internal/serve/workflow"
	"jikime-adk/internal/serve/workspace"
)

// NewServe creates the `jikime serve` cobra command.
func NewServe() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve [WORKFLOW.md]",
		Short: "Run autonomous agent orchestration service",
		Long: `Starts a long-running service that polls an issue tracker (GitHub Issues),
creates isolated git worktrees per issue, and runs Claude Code agents autonomously.

Inspired by OpenAI Symphony (SPEC: github.com/openai/symphony).

Examples:
  jikime serve                    # uses ./WORKFLOW.md
  jikime serve path/to/WORKFLOW.md
  jikime serve --port 8080        # enables HTTP status API`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowPath := "WORKFLOW.md"
			if len(args) > 0 {
				workflowPath = args[0]
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(workflowPath)
			if err != nil {
				return fmt.Errorf("resolve workflow path: %w", err)
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("WORKFLOW.md not found: %s\n\nCreate a WORKFLOW.md file first. See: jikime serve --help", absPath)
			}

			// Override port from WORKFLOW.md server.port if not set via flag
			if port == 0 {
				// Will be read from config after loading
			}

			return run(absPath, port)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "HTTP API server port (0 = disabled)")
	return cmd
}

func run(workflowPath string, cliPort int) error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	printBanner()
	logger.Info("loading workflow", "path", workflowPath)

	// Load WORKFLOW.md with hot-reload
	var orch *orchestrator.Orchestrator
	loader, err := workflow.NewLoader(workflowPath, func(def *serve.WorkflowDefinition) {
		logger.Info("WORKFLOW.md reloaded")
		if orch != nil {
			orch.ApplyConfig(workflow.NewConfig(def))
		}
	})
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}
	defer loader.Close()

	cfg := workflow.NewConfig(loader.Current())

	// Validate config before starting
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("workflow validation failed: %w\n\nCheck your WORKFLOW.md configuration.", err)
	}

	// Determine HTTP port
	port := cliPort
	if port == 0 {
		port = cfg.ServerPort()
	}

	// Create tracker client
	t, err := tracker.NewClient(
		cfg.TrackerKind(),
		cfg.TrackerEndpoint(),
		cfg.TrackerAPIKey(),
		cfg.TrackerProjectSlug(),
		cfg.TrackerActiveStates(),
		cfg.TrackerTerminalStates(),
	)
	if err != nil {
		return fmt.Errorf("tracker init: %w", err)
	}

	// Create workspace manager
	wsManager := workspace.NewManager(
		cfg.WorkspaceRoot(),
		workspace.WithHookAfterCreate(cfg.HookAfterCreate()),
		workspace.WithHookBeforeRun(cfg.HookBeforeRun()),
		workspace.WithHookAfterRun(cfg.HookAfterRun()),
		workspace.WithHookBeforeRemove(cfg.HookBeforeRemove()),
		workspace.WithHookTimeoutMS(cfg.HookTimeoutMS()),
		workspace.WithLogger(logger),
	)

	// Create agent runner
	agentRunner := agent.NewRunner(
		agent.WithClaudeCommand(cfg.ClaudeCommand()),
		agent.WithTurnTimeoutMS(cfg.TurnTimeoutMS()),
		agent.WithStallTimeoutMS(cfg.StallTimeoutMS()),
		agent.WithLogger(logger),
		agent.WithEventCallback(func(e serve.AgentEvent) {
			logger.Info("agent event",
				"type", string(e.Type),
				"issue_id", e.IssueID,
				"message", e.Message,
			)
		}),
	)

	// Create orchestrator
	orch = orchestrator.New(cfg, t, wsManager, agentRunner, logger)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("shutdown signal received", "signal", sig)
		cancel()
	}()

	// Print config summary
	printConfig(cfg, workflowPath, port)

	// Startup cleanup (remove terminal-state workspaces)
	logger.Info("running startup cleanup...")
	orch.StartupCleanup(ctx)

	// Start HTTP API if port configured
	if port > 0 {
		go startHTTPServer(port, orch, logger)
	}

	// Start orchestration loop (blocks until ctx cancelled)
	orch.Run(ctx)

	logger.Info("jikime serve stopped")
	return nil
}

// --- HTTP API Server ---

func startHTTPServer(port int, orch *orchestrator.Orchestrator, logger *slog.Logger) {
	mux := http.NewServeMux()

	// GET / — human-readable dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		snap := orch.Snapshot()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "jikime serve — %s\n", snap.GeneratedAt.Format(time.RFC3339))
		fmt.Fprintf(w, "\nRunning (%d):\n", len(snap.Running))
		for _, r := range snap.Running {
			fmt.Fprintf(w, "  %-20s  turns=%-3d  %s\n",
				r.IssueIdentifier, r.TurnCount, r.LastMessage)
		}
		fmt.Fprintf(w, "\nRetrying (%d):\n", len(snap.Retrying))
		for _, r := range snap.Retrying {
			fmt.Fprintf(w, "  %-20s  attempt=%d  due=%s  error=%s\n",
				r.Identifier, r.Attempt, r.DueAt.Format("15:04:05"), r.Error)
		}
		fmt.Fprintf(w, "\nTokens: input=%d output=%d total=%d runtime=%.1fs\n",
			snap.Totals.InputTokens, snap.Totals.OutputTokens,
			snap.Totals.TotalTokens, snap.Totals.SecondsRunning,
		)
	})

	// GET /api/v1/state
	mux.HandleFunc("/api/v1/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		snap := orch.Snapshot()
		writeJSON(w, http.StatusOK, map[string]any{
			"generated_at": snap.GeneratedAt,
			"counts": map[string]int{
				"running":  len(snap.Running),
				"retrying": len(snap.Retrying),
			},
			"running":      snap.Running,
			"retrying":     snap.Retrying,
			"jikime_totals": snap.Totals,
		})
	})

	// POST /api/v1/refresh — trigger immediate poll
	mux.HandleFunc("/api/v1/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{
			"queued":       true,
			"requested_at": time.Now(),
			"operations":  []string{"poll", "reconcile"},
		})
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Info("HTTP API started", "addr", addr)
	srv := &http.Server{Addr: addr, Handler: mux}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", "error", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// --- Banner & Config Display ---

func printBanner() {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║   jikime serve — Agent Orchestrator  ║")
	fmt.Println("  ║   Powered by Claude Code + Symphony  ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
}

func printConfig(cfg *workflow.Config, workflowPath string, port int) {
	fmt.Printf("  Workflow:    %s\n", workflowPath)
	fmt.Printf("  Tracker:     %s / %s\n", cfg.TrackerKind(), cfg.TrackerProjectSlug())
	fmt.Printf("  Workspace:   %s\n", cfg.WorkspaceRoot())
	fmt.Printf("  Concurrency: %d agents\n", cfg.MaxConcurrentAgents())
	fmt.Printf("  Poll:        every %dms\n", cfg.PollIntervalMS())
	if port > 0 {
		fmt.Printf("  HTTP API:    http://127.0.0.1:%d\n", port)
	}
	fmt.Println()
}
