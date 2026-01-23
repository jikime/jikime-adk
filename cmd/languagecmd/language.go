// Package languagecmd provides language management commands for jikime-adk.
package languagecmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// LanguageInfo represents information about a supported language
type LanguageInfo struct {
	Name       string `json:"name" yaml:"name"`
	NativeName string `json:"native_name" yaml:"native_name"`
	Family     string `json:"family" yaml:"family"`
}

// SupportedLanguages is the map of all supported languages
var SupportedLanguages = map[string]LanguageInfo{
	"en": {Name: "English", NativeName: "English", Family: "Germanic"},
	"ko": {Name: "Korean", NativeName: "한국어", Family: "Koreanic"},
	"ja": {Name: "Japanese", NativeName: "日本語", Family: "Japonic"},
	"zh": {Name: "Chinese", NativeName: "中文", Family: "Sinitic"},
	"es": {Name: "Spanish", NativeName: "Español", Family: "Romance"},
	"fr": {Name: "French", NativeName: "Français", Family: "Romance"},
	"de": {Name: "German", NativeName: "Deutsch", Family: "Germanic"},
	"pt": {Name: "Portuguese", NativeName: "Português", Family: "Romance"},
	"it": {Name: "Italian", NativeName: "Italiano", Family: "Romance"},
	"ru": {Name: "Russian", NativeName: "Русский", Family: "Slavic"},
}

// NewLanguage creates the language command group.
func NewLanguage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "language",
		Short: "Language management and multilingual support",
		Long:  "Manage language configuration and multilingual support for Jikime-ADK.",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newValidateCmd())

	return cmd
}

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all supported languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json-output", false, "Output as JSON")

	return cmd
}

func newInfoCmd() *cobra.Command {
	var detail bool

	cmd := &cobra.Command{
		Use:   "info [language_code]",
		Short: "Show information about a specific language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(args[0], detail)
		},
	}

	cmd.Flags().BoolVar(&detail, "detail", false, "Show detailed information")

	return cmd
}

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [language_code]",
		Short: "Set the conversation language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(args[0])
		},
	}

	return cmd
}

func newValidateCmd() *cobra.Command {
	var validateLanguages bool

	cmd := &cobra.Command{
		Use:   "validate [config_file]",
		Short: "Validate language configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := ""
			if len(args) > 0 {
				configFile = args[0]
			}
			return runValidate(configFile, validateLanguages)
		},
	}

	cmd.Flags().BoolVar(&validateLanguages, "validate-languages", false, "Validate language codes in config")

	return cmd
}

func runList(jsonOutput bool) error {
	if jsonOutput {
		data, err := json.MarshalIndent(SupportedLanguages, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	blue := color.New(color.FgBlue)

	cyan.Println("Supported Languages")
	cyan.Println("═══════════════════")
	fmt.Println()
	fmt.Printf("%-8s %-15s %-15s %-10s\n", "Code", "English Name", "Native Name", "Family")
	fmt.Println("────────────────────────────────────────────────────")

	for code, info := range SupportedLanguages {
		fmt.Printf("%-8s %-15s %-15s %-10s\n",
			green.Sprint(code),
			info.Name,
			yellow.Sprint(info.NativeName),
			blue.Sprint(info.Family))
	}

	return nil
}

func runInfo(langCode string, detail bool) error {
	code := strings.ToLower(langCode)
	info, exists := SupportedLanguages[code]

	if !exists {
		red := color.New(color.FgRed)
		red.Printf("Language code '%s' not found.\n", langCode)

		// Show available codes
		codes := make([]string, 0, len(SupportedLanguages))
		for c := range SupportedLanguages {
			codes = append(codes, c)
		}
		fmt.Printf("Available codes: %s\n", strings.Join(codes, ", "))
		return nil
	}

	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	cyan.Println("Language Information")
	cyan.Println("════════════════════")
	fmt.Printf("%-15s %s\n", "Code:", green.Sprint(code))
	fmt.Printf("%-15s %s\n", "English Name:", info.Name)
	fmt.Printf("%-15s %s\n", "Native Name:", info.NativeName)
	fmt.Printf("%-15s %s\n", "Family:", info.Family)

	if detail {
		fmt.Println()
		fmt.Printf("%-15s %s\n", "Optimal Model:", getOptimalModel(code))
	}

	return nil
}

func runSet(langCode string) error {
	code := strings.ToLower(langCode)
	info, exists := SupportedLanguages[code]

	if !exists {
		red := color.New(color.FgRed)
		red.Printf("Language code '%s' not found.\n", langCode)
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Update language.yaml in config
	langPath := filepath.Join(cwd, ".jikime", "config", "language.yaml")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(langPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	langConfig := map[string]interface{}{
		"language": map[string]interface{}{
			"conversation_language":      code,
			"conversation_language_name": info.NativeName,
			"agent_prompt_language":      "en",
			"git_commit_messages":        "en",
			"code_comments":              "en",
			"documentation":              "en",
			"error_messages":             "en",
		},
	}

	data, err := yaml.Marshal(langConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := "# Language Settings (CLAUDE.md Reference)\n# This file is auto-loaded by CLAUDE.md for language configuration\n\n"
	content := header + string(data)

	if err := os.WriteFile(langPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	green := color.New(color.FgGreen)
	green.Printf("Language set to: %s (%s)\n", code, info.NativeName)
	fmt.Printf("Configuration saved to: %s\n", langPath)

	return nil
}

func runValidate(configFile string, validateLanguages bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Default config file path
	if configFile == "" {
		configFile = filepath.Join(cwd, ".jikime", "config", "language.yaml")
	}

	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Printf("Validating config: %s\n", configFile)
	fmt.Println()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Check language section
	langSection, ok := config["language"]
	if !ok {
		yellow.Println("  No 'language' section found in config")
		return nil
	}

	langMap, ok := langSection.(map[string]interface{})
	if !ok {
		red.Println("  'language' section must be an object")
		return nil
	}

	green.Println("  Language section structure is valid")

	// Check conversation_language
	convLang, ok := langMap["conversation_language"].(string)
	if ok && convLang != "" {
		if _, exists := SupportedLanguages[convLang]; exists {
			green.Printf("  conversation_language '%s' is supported\n", convLang)
		} else {
			red.Printf("  conversation_language '%s' is not supported\n", convLang)
		}
	} else {
		yellow.Println("  No conversation_language specified")
	}

	// Check conversation_language_name
	convLangName, _ := langMap["conversation_language_name"].(string)
	if convLangName != "" && convLang != "" {
		expectedName := SupportedLanguages[convLang].NativeName
		if convLangName == expectedName {
			green.Println("  conversation_language_name matches")
		} else {
			yellow.Printf("  conversation_language_name '%s' doesn't match expected '%s'\n",
				convLangName, expectedName)
		}
	}

	// Validate all language codes if requested
	if validateLanguages {
		fmt.Println()
		fmt.Println("Scanning for language codes...")
		configStr := string(data)
		found := []string{}
		for code := range SupportedLanguages {
			if strings.Contains(configStr, code) {
				found = append(found, code)
			}
		}
		if len(found) > 0 {
			fmt.Printf("Found language codes in config: %s\n", strings.Join(found, ", "))
		}
	}

	return nil
}

func getOptimalModel(langCode string) string {
	// All languages work well with Claude
	switch langCode {
	case "en":
		return "claude-sonnet-4-20250514"
	case "ko", "ja", "zh":
		return "claude-sonnet-4-20250514 (excellent CJK support)"
	default:
		return "claude-sonnet-4-20250514"
	}
}

// GetNativeName returns the native name for a language code
func GetNativeName(code string) string {
	if info, exists := SupportedLanguages[code]; exists {
		return info.NativeName
	}
	return code
}

// GetAllSupportedCodes returns all supported language codes
func GetAllSupportedCodes() []string {
	codes := make([]string, 0, len(SupportedLanguages))
	for code := range SupportedLanguages {
		codes = append(codes, code)
	}
	return codes
}
