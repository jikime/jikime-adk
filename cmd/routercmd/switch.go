package routercmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	router "jikime-adk/internal/router"
)

// Claude Code env keys managed by switch
var managedEnvKeys = []string{
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_API_KEY",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
}

func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <provider>",
		Short: "Switch LLM provider for Claude Code",
		Long: `Switch between LLM providers. Updates the project's .claude/settings.local.json
to route requests through the appropriate backend.

Available providers:
  claude   - Use native Claude (default, removes proxy settings)
  openai   - Use OpenAI via router proxy
  gemini   - Use Gemini via router proxy
  glm      - Use GLM directly (Anthropic-compatible endpoint)
  ollama   - Use Ollama via router proxy`,
		Args:             cobra.ExactArgs(1),
		ValidArgs:        []string{"claude", "openai", "gemini", "glm", "ollama"},
		RunE:             runSwitch,
	}
}

func runSwitch(cmd *cobra.Command, args []string) error {
	input := strings.ToLower(args[0])

	fmt.Println()

	// Handle "claude" case: restore native mode
	if input == "claude" {
		return switchToClaude()
	}

	// Parse provider/model format
	provider, modelOverride := parseProviderInput(input)

	// Load router config
	cfg, err := router.LoadConfig()
	if err != nil {
		return err
	}

	// Check if provider exists in config
	provCfg, ok := cfg.Providers[provider]
	if !ok {
		return fmt.Errorf("provider '%s' not found in config (%s)", provider, router.ConfigPath())
	}

	// Apply model override if specified
	if modelOverride != "" {
		provCfg.Model = modelOverride
	}

	// Validate API key is available (skip for ollama)
	// LoadConfig resolves env vars via resolveAPIKeys(), so empty or unexpanded means missing
	if provider != "ollama" && (provCfg.APIKey == "" || strings.Contains(provCfg.APIKey, "${")) {
		envVar := providerEnvVar(provider)
		fmt.Println()
		color.Red("  API key not set for '%s'.", provider)
		fmt.Println()
		fmt.Println("  Add to ~/.zshrc (or ~/.bashrc):")
		fmt.Printf("     export %s=\"your-api-key\"\n", envVar)
		fmt.Println()
		fmt.Println("  Then reload and retry:")
		fmt.Println("     source ~/.zshrc")
		fmt.Printf("     jikime router switch %s\n", provider)
		fmt.Println()
		return fmt.Errorf("API key not set for '%s'", provider)
	}

	if provCfg.IsAnthropicCompatible() {
		return switchAnthropicCompatible(provider, &provCfg)
	}

	return switchViaProxy(provider, cfg, &provCfg)
}

// providerEnvVar returns the expected environment variable name for a provider's API key.
func providerEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "glm":
		return "GLM_API_KEY"
	default:
		return strings.ToUpper(provider) + "_API_KEY"
	}
}

// parseProviderInput splits "provider/model" into parts.
func parseProviderInput(input string) (provider, model string) {
	parts := strings.SplitN(input, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// switchToClaude removes managed env vars and stops the router if running.
func switchToClaude() error {
	// Check if already using Claude
	if !hasManagedEnv() {
		color.Yellow("  Already using Claude backend.")
		fmt.Println()
		return nil
	}

	// Remove managed env keys from .claude/settings.local.json
	if err := removeManagedEnv(); err != nil {
		return fmt.Errorf("removing env from settings: %w", err)
	}

	// Stop router if running
	if pid := readPID(); pid > 0 && processExists(pid) {
		runStop(nil, nil)
	}

	// Clear router state
	router.ClearState()

	color.Green("  Switched to Claude (native)")
	fmt.Println("  Removed proxy settings from Claude Code.")
	printRestartNotice()
	return nil
}

// switchAnthropicCompatible sets env vars directly (no proxy needed).
func switchAnthropicCompatible(provider string, provCfg *router.ProviderConfig) error {
	envVars := map[string]string{
		"ANTHROPIC_BASE_URL":             provCfg.AnthropicURL,
		"ANTHROPIC_API_KEY":              provCfg.APIKey,
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  provCfg.Model,
		"ANTHROPIC_DEFAULT_SONNET_MODEL": provCfg.Model,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   provCfg.Model,
	}

	if err := setManagedEnv(envVars); err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}

	// Save router state for statusline
	router.SaveState(&router.RouterState{
		Provider: provider,
		Model:    provCfg.Model,
		Mode:     "direct",
		Active:   true,
	})

	cyan := color.New(color.FgCyan).SprintFunc()
	color.Green("  Switched to %s (direct, no proxy needed)", provider)
	fmt.Printf("  Model:    %s\n", cyan(provCfg.Model))
	fmt.Printf("  Endpoint: %s\n", cyan(provCfg.AnthropicURL))
	printRestartNotice()
	return nil
}

// switchViaProxy ensures router is running and configures settings.
func switchViaProxy(provider string, cfg *router.Config, provCfg *router.ProviderConfig) error {
	// Update active provider in config
	cfg.Router.Provider = provider
	if err := router.SaveConfig(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Restart router to pick up new provider config
	if pid := readPID(); pid > 0 && processExists(pid) {
		runStop(nil, nil)
	}
	startDaemon = true
	if err := runStart(nil, nil); err != nil {
		return fmt.Errorf("starting router: %w", err)
	}

	addr := fmt.Sprintf("http://%s:%d", cfg.Router.Host, cfg.Router.Port)

	envVars := map[string]string{
		"ANTHROPIC_BASE_URL":             addr,
		"ANTHROPIC_API_KEY":              "router-proxy",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  provCfg.Model,
		"ANTHROPIC_DEFAULT_SONNET_MODEL": provCfg.Model,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   provCfg.Model,
	}

	if err := setManagedEnv(envVars); err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}

	// Save router state for statusline
	router.SaveState(&router.RouterState{
		Provider: provider,
		Model:    provCfg.Model,
		Mode:     "proxy",
		Active:   true,
	})

	cyan := color.New(color.FgCyan).SprintFunc()
	color.Green("  Switched to %s (via proxy)", provider)
	fmt.Printf("  Model:   %s\n", cyan(provCfg.Model))
	fmt.Printf("  Router:  %s\n", cyan(addr))
	printRestartNotice()
	return nil
}

func printRestartNotice() {
	gray := color.New(color.FgHiBlack).SprintFunc()
	fmt.Printf("  Settings: %s\n", gray(router.ClaudeSettingsPath()))
	fmt.Println()
	color.Yellow("  Restart Claude Code to apply changes.")
	fmt.Println()
}

// --- .claude/settings.local.json helpers ---

type claudeSettings struct {
	data map[string]any
}

func loadClaudeSettings() (*claudeSettings, error) {
	path := router.ClaudeSettingsPath()
	if path == "" {
		return nil, fmt.Errorf("not inside a project (no .git or .claude directory found)")
	}
	s := &claudeSettings{data: make(map[string]any)}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(raw, &s.data); err != nil {
		// If corrupt, start fresh
		return s, nil
	}

	return s, nil
}

func (s *claudeSettings) save() error {
	path := router.ClaudeSettingsPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func (s *claudeSettings) getEnv() map[string]any {
	env, ok := s.data["env"]
	if !ok {
		return nil
	}
	envMap, ok := env.(map[string]any)
	if !ok {
		return nil
	}
	return envMap
}

func (s *claudeSettings) setEnv(vars map[string]string) {
	env := s.getEnv()
	if env == nil {
		env = make(map[string]any)
		s.data["env"] = env
	}
	for k, v := range vars {
		env[k] = v
	}
}

func (s *claudeSettings) removeEnvKeys(keys []string) {
	env := s.getEnv()
	if env == nil {
		return
	}
	for _, k := range keys {
		delete(env, k)
	}
	// Remove env section if empty
	if len(env) == 0 {
		delete(s.data, "env")
	}
}

func hasManagedEnv() bool {
	path := router.ClaudeSettingsPath()
	if path == "" {
		return false
	}
	s, err := loadClaudeSettings()
	if err != nil {
		return false
	}
	env := s.getEnv()
	if env == nil {
		return false
	}
	_, ok := env["ANTHROPIC_BASE_URL"]
	return ok
}

func setManagedEnv(vars map[string]string) error {
	s, err := loadClaudeSettings()
	if err != nil {
		return err
	}
	s.setEnv(vars)
	return s.save()
}

func removeManagedEnv() error {
	s, err := loadClaudeSettings()
	if err != nil {
		return err
	}
	s.removeEnvKeys(managedEnvKeys)
	return s.save()
}
