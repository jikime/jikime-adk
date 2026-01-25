package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the router configuration.
type Config struct {
	Router    RouterConfig              `yaml:"router"`
	Providers map[string]ProviderConfig `yaml:"providers"`
	Scenarios *ScenarioConfig           `yaml:"scenarios,omitempty"`
}

// RouterConfig contains router server settings.
type RouterConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// ProviderConfig contains provider-specific settings.
type ProviderConfig struct {
	APIKey       string `yaml:"-"`
	Model        string `yaml:"model"`
	BaseURL      string `yaml:"base_url,omitempty"`
	Region       string `yaml:"region,omitempty"`       // for GLM: international, china
	AnthropicURL string `yaml:"anthropic_url,omitempty"` // if set, provider is Anthropic-compatible (no proxy needed)
}

// IsAnthropicCompatible returns true if the provider has an Anthropic-compatible endpoint.
func (p *ProviderConfig) IsAnthropicCompatible() bool {
	return p.AnthropicURL != ""
}

// ScenarioConfig contains scenario-based routing settings.
type ScenarioConfig struct {
	Default              string `yaml:"default,omitempty"`
	Background           string `yaml:"background,omitempty"`
	Think                string `yaml:"think,omitempty"`
	LongContext          string `yaml:"long_context,omitempty"`
	LongContextThreshold int    `yaml:"long_context_threshold,omitempty"`
}

// DefaultConfig returns a config with default values.
// API keys are resolved from environment variables by resolveAPIKeys().
func DefaultConfig() *Config {
	return &Config{
		Router: RouterConfig{
			Port: 8787,
			Host: "127.0.0.1",
		},
		Providers: map[string]ProviderConfig{
			"openai": {
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-5.1",
			},
			"gemini": {
				BaseURL: "https://generativelanguage.googleapis.com",
				Model:   "gemini-2.5-flash",
			},
			"glm": {
				BaseURL:      "https://api.z.ai/api/paas/v4",
				Model:        "glm-4.7",
				Region:       "international",
				AnthropicURL: "https://api.z.ai/api/anthropic",
			},
			"ollama": {
				BaseURL: "http://localhost:11434",
				Model:   "llama3.1",
			},
		},
	}
}

// ConfigPath returns the path to the router config file.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jikime", "router.yaml")
}

// PIDPath returns the path to the router PID file.
func PIDPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jikime", "router.pid")
}

// ClaudeSettingsPath returns the path to the project-level .claude/settings.local.json.
// It walks up from the current directory to find the project root (containing .git or .claude).
// Returns empty string if no project root is found.
func ClaudeSettingsPath() string {
	if root := findProjectRoot(); root != "" {
		return filepath.Join(root, ".claude", "settings.local.json")
	}
	return ""
}

// findProjectRoot walks up from cwd looking for .git or .claude directory.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check for .git directory
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir
		}
		// Check for .claude directory
		if info, err := os.Stat(filepath.Join(dir, ".claude")); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}

	return ""
}

// RouterState represents the current router state for statusline integration.
type RouterState struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Mode     string `json:"mode"` // "proxy" or "direct"
	Active   bool   `json:"active"`
}

// StatePath returns the path to the router state file.
func StatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jikime", "router-state.json")
}

// SaveState writes the current router state.
func SaveState(state *RouterState) error {
	path := StatePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadState reads the current router state.
func LoadState() *RouterState {
	data, err := os.ReadFile(StatePath())
	if err != nil {
		return nil
	}
	state := &RouterState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil
	}
	return state
}

// ClearState removes the router state file.
func ClearState() {
	os.Remove(StatePath())
}

// LoadConfig loads the router configuration from the config file.
// If the file does not exist, it creates one with default values.
func LoadConfig() (*Config, error) {
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Auto-create with defaults for existing users who haven't run 'jikime init'
			cfg := DefaultConfig()
			if writeErr := SaveConfig(cfg); writeErr != nil {
				return nil, fmt.Errorf("creating default config: %w", writeErr)
			}
			fmt.Printf("  Created default router config: %s\n", path)
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults for missing values
	if cfg.Router.Port == 0 {
		cfg.Router.Port = 8787
	}
	if cfg.Router.Host == "" {
		cfg.Router.Host = "127.0.0.1"
	}

	// Resolve API keys from environment variables if not set in config
	resolveAPIKeys(cfg)

	return cfg, nil
}

// resolveAPIKeys fills in missing API keys from known environment variables.
func resolveAPIKeys(cfg *Config) {
	envVars := map[string]string{
		"openai": "OPENAI_API_KEY",
		"gemini": "GEMINI_API_KEY",
		"glm":    "GLM_API_KEY",
	}

	for name, envVar := range envVars {
		prov, ok := cfg.Providers[name]
		if !ok {
			continue
		}
		if prov.APIKey == "" || strings.Contains(prov.APIKey, "${") {
			if val := os.Getenv(envVar); val != "" {
				prov.APIKey = val
				cfg.Providers[name] = prov
			}
		}
	}
}

// SaveConfig writes the configuration to the config file.
func SaveConfig(cfg *Config) error {
	path := ConfigPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// GetProvider returns the provider config for the specified provider name.
func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	prov, ok := c.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found in config", name)
	}
	return &prov, nil
}

// GetProviderNames returns a list of all configured provider names.
func (c *Config) GetProviderNames() []string {
	names := make([]string, 0, len(c.Providers))
	for name := range c.Providers {
		names = append(names, name)
	}
	return names
}

// parseProviderModel splits "provider/model" into parts.
func parseProviderModel(s string) (string, string) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// expandEnvVars expands ${VAR} and $VAR references in the string.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		return match // Keep original if env var not set
	})
}
