package hooks

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the complete configuration from section files
type Config struct {
	User        UserConfig        `yaml:"user"`
	Language    LanguageConfig    `yaml:"language"`
	Project     ProjectConfig     `yaml:"project"`
	GitStrategy GitStrategyConfig `yaml:"git_strategy"`
	Quality     QualityConfig     `yaml:"quality"`
}

// UserConfig holds user configuration
type UserConfig struct {
	Name string `yaml:"name"`
}

// LanguageConfig holds language configuration
type LanguageConfig struct {
	ConversationLanguage     string `yaml:"conversation_language"`
	ConversationLanguageName string `yaml:"conversation_language_name"`
	GitCommitMessages        string `yaml:"git_commit_messages"`
	CodeComments             string `yaml:"code_comments"`
	Documentation            string `yaml:"documentation"`
}

// ProjectConfig holds project configuration
type ProjectConfig struct {
	Name string `yaml:"name"`
}

// GitStrategyConfig holds git strategy configuration
type GitStrategyConfig struct {
	Mode     string           `yaml:"mode"`
	Manual   GitModeConfig    `yaml:"manual"`
	Personal GitModeConfig    `yaml:"personal"`
	Team     GitModeConfig    `yaml:"team"`
}

// GitModeConfig holds mode-specific git configuration
type GitModeConfig struct {
	BranchCreation BranchCreationConfig `yaml:"branch_creation"`
}

// BranchCreationConfig holds branch creation settings
type BranchCreationConfig struct {
	AutoEnabled bool `yaml:"auto_enabled"`
}

// QualityConfig holds quality configuration
type QualityConfig struct {
	DDD struct {
		MinCoverage int `yaml:"min_coverage"`
	} `yaml:"ddd"`
	Linting struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"linting"`
	Formatting struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"formatting"`
}

// LoadConfig loads configuration from section files
func LoadConfig() (*Config, error) {
	sectionsDir, err := FindSectionsDir()
	if err != nil {
		return nil, err
	}

	config := &Config{}
	sectionFiles := []string{
		"user.yaml",
		"language.yaml",
		"project.yaml",
		"git-strategy.yaml",
		"quality.yaml",
	}

	for _, filename := range sectionFiles {
		path := filepath.Join(sectionsDir, filename)
		if !FileExists(path) {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			continue
		}
	}

	return config, nil
}

// GetLanguageCode returns the conversation language code (defaults to "en")
func (c *Config) GetLanguageCode() string {
	if c.Language.ConversationLanguage != "" {
		return c.Language.ConversationLanguage
	}
	return "en"
}

// GetLanguageName returns the conversation language name (defaults to "English")
func (c *Config) GetLanguageName() string {
	if c.Language.ConversationLanguageName != "" {
		return c.Language.ConversationLanguageName
	}
	return "English"
}

// GetUserName returns the user name
func (c *Config) GetUserName() string {
	return c.User.Name
}

// GetGitMode returns the git strategy mode (defaults to "manual")
func (c *Config) GetGitMode() string {
	if c.GitStrategy.Mode != "" {
		return c.GitStrategy.Mode
	}
	return "manual"
}

// IsAutoBranchEnabled checks if auto branch is enabled for current mode
func (c *Config) IsAutoBranchEnabled() bool {
	switch c.GetGitMode() {
	case "manual":
		return c.GitStrategy.Manual.BranchCreation.AutoEnabled
	case "personal":
		return c.GitStrategy.Personal.BranchCreation.AutoEnabled
	case "team":
		return c.GitStrategy.Team.BranchCreation.AutoEnabled
	default:
		return false
	}
}

// IsLintingEnabled checks if linting is enabled
func (c *Config) IsLintingEnabled() bool {
	return c.Quality.Linting.Enabled
}

// IsFormattingEnabled checks if formatting is enabled
func (c *Config) IsFormattingEnabled() bool {
	return c.Quality.Formatting.Enabled
}
