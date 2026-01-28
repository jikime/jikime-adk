package initcmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
	"jikime-adk/cmd/banner"
	"jikime-adk/cmd/ui"
	"jikime-adk/project"
	"jikime-adk/version"
)

// Cyberpunk colors
var (
	cyan      = lipgloss.Color("#00FFFF")
	magenta   = lipgloss.Color("#FF00FF")
	neonGreen = lipgloss.Color("#39FF14")
	neonPink  = lipgloss.Color("#FF6EC7")
	dimCyan   = lipgloss.Color("#008B8B")
	white     = lipgloss.Color("#FFFFFF")
	dim       = lipgloss.Color("#666666")
	success   = lipgloss.Color("#00FF00")
	warning   = lipgloss.Color("#FFD700")
)

// Styles
var (
	sectionStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(success)

	dimStyle = lipgloss.NewStyle().
			Foreground(dim)

	highlightStyle = lipgloss.NewStyle().
			Foreground(neonPink).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(magenta).
			Bold(true)

	systemStyle = lipgloss.NewStyle().
			Foreground(neonGreen).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warning)

	whiteStyle = lipgloss.NewStyle().
			Foreground(white)
)

var languageOptions = []languageOption{
	{value: "ko", display: "Korean (한국어)"},
	{value: "en", display: "English"},
	{value: "ja", display: "Japanese (日本語)"},
	{value: "zh", display: "Chinese (中文)"},
}

func languageDisplay(value string) string {
	for _, opt := range languageOptions {
		if opt.value == value {
			return opt.display
		}
	}
	return value
}

func NewInit() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init [path] [project-name]",
		Short: "Initialize a new jikime-adk project",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			projectNameArg := ""
			if len(args) > 0 {
				path = args[0]
			}
			if len(args) > 1 {
				projectNameArg = args[1]
			}

			projectPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			// Print cyberpunk intro
			banner.PrintIntro(version.String())

			answers, err := collectSetupAnswers(opts, projectPath, projectNameArg)
			if err != nil {
				return err
			}

			localizer := getTranslation(answers.Locale)

			// Setup complete message
			fmt.Println()
			fmt.Println(successStyle.Render("    ✓ ") + whiteStyle.Render(localize(localizer, "setup_completed", nil)))

			// Starting installation
			ui.PrintBorderedSection("INSTALLATION SEQUENCE")
			ui.PrintBorderedEmpty()

			initializer := project.NewInitializer(projectPath)
			start := time.Now()

			// Phase 1: Preparation
			printPhaseWithProgress(1, localize(localizer, "phase1", nil))

			// Phase 2: Creating directories
			printPhaseWithProgress(2, localize(localizer, "phase2", nil))

			// Phase 3: Installing resources
			banner.PrintPhase(3, localize(localizer, "phase3", nil))
			result, err := initializer.Initialize(answers, opts.force)
			if err != nil {
				return err
			}
			banner.PrintPhaseOK()

			// Phase 4: Generating configurations
			printPhaseWithProgress(4, localize(localizer, "phase4", nil))

			// Phase 5: Validation
			printPhaseWithProgress(5, localize(localizer, "phase5", nil))

			duration := time.Since(start)

			// Display final result
			displayFinalResult(result, answers, duration, localizer)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.nonInteractive, "non-interactive", "y", false, "Use defaults without prompts")
	cmd.Flags().StringVar(&opts.mode, "mode", "personal", "Project mode")
	cmd.Flags().StringVar(&opts.locale, "locale", "en", "Conversation locale (ko/en/ja/zh)")
	cmd.Flags().StringVar(&opts.language, "language", "", "Programming language")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Force reinitialize without confirmation")

	return cmd
}

func printPhaseWithProgress(num int, label string) {
	banner.PrintPhase(num, label)
	// Simulate progress with cyber animation
	printCyberProgress(40, 150*time.Millisecond)
	banner.PrintPhaseOK()
}

func printCyberProgress(width int, duration time.Duration) {
	borderPrefix := lipgloss.NewStyle().Foreground(dimCyan).Render("  │")
	chars := []string{"░", "▒", "▓", "█"}
	steps := width

	for i := 0; i <= steps; i++ {
		bar := ""
		for j := 0; j < width; j++ {
			if j < i {
				bar += lipgloss.NewStyle().Foreground(neonGreen).Render("█")
			} else if j == i {
				bar += lipgloss.NewStyle().Foreground(cyan).Render(chars[i%len(chars)])
			} else {
				bar += lipgloss.NewStyle().Foreground(dim).Render("░")
			}
		}

		percent := lipgloss.NewStyle().Foreground(neonGreen).Bold(true).Render(fmt.Sprintf("%3d%%", (i*100)/steps))
		fmt.Printf("\r%s          %s %s", borderPrefix, bar, percent)
		time.Sleep(duration / time.Duration(steps))
	}
	fmt.Println()
}

type languageOption struct {
	value   string
	display string
}

func collectSetupAnswers(opts *initOptions, projectPath, defaultProjectName string) (project.SetupAnswers, error) {
	answers := project.SetupAnswers{
		ProjectName:     defaultProjectName,
		Locale:          opts.locale,
		UserName:        "",
		Honorific:       "",
		TonePreset:      "friendly",
		GitMode:         "manual",
		GitHubUser:      "",
		GitCommitLang:   "en",
		CodeCommentLang: "en",
		DocLang:         "en",
		TagEnabled:      true,
		TagMode:         "warn",
	}

	if answers.ProjectName == "" {
		answers.ProjectName = filepath.Base(projectPath)
	}

	if opts.nonInteractive {
		return answers, nil
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)
	go func() {
		<-interrupt
		fmt.Println()
		fmt.Println(warningStyle.Render("    ⚠ ABORT: User cancelled initialization"))
		os.Exit(1)
	}()

	localizer := getTranslation(answers.Locale)

	// Language Selection
	ui.PrintBorderedSection(localize(localizer, "language_section", nil))
	answers.Locale = promptSelectLanguage(localize(localizer, "language_prompt", nil), languageOptions, answers.Locale)

	localizer = getTranslation(answers.Locale)
	ui.PrintBorderedResult(localize(localizer, "language_label", nil), languageDisplay(answers.Locale))

	// User Section
	ui.PrintBorderedSection(localize(localizer, "user_section", nil))
	answers.UserName = promptInputWithLabel(localize(localizer, "user_prompt", nil), localize(localizer, "user_name_label", nil), answers.UserName)
	answers.Honorific = promptInputWithLabel(localize(localizer, "honorific_prompt", nil), localize(localizer, "honorific_label", nil), answers.Honorific)

	// Tone preset selection
	ui.PrintBorderedLine(dimStyle.Render(localize(localizer, "tone_preset_help", nil)))
	tonePresetOptions := []selectOption{
		{
			Value:       "friendly",
			Display:     localize(localizer, "tone_friendly_display", nil),
			Description: localize(localizer, "tone_friendly_desc", nil),
		},
		{
			Value:       "professional",
			Display:     localize(localizer, "tone_professional_display", nil),
			Description: localize(localizer, "tone_professional_desc", nil),
		},
		{
			Value:       "casual",
			Display:     localize(localizer, "tone_casual_display", nil),
			Description: localize(localizer, "tone_casual_desc", nil),
		},
		{
			Value:       "mentor",
			Display:     localize(localizer, "tone_mentor_display", nil),
			Description: localize(localizer, "tone_mentor_desc", nil),
		},
	}
	answers.TonePreset = promptSelectWithDescriptionWithLabel(tonePresetOptions, localize(localizer, "tone_preset_label", nil), answers.TonePreset)

	// Project name
	answers.ProjectName = promptInputWithLabel(localize(localizer, "project_prompt", nil), localize(localizer, "project_name_label", nil), answers.ProjectName)

	// Git Setup section
	ui.PrintBorderedSection(localize(localizer, "git_setup_section", nil))
	ui.PrintBorderedLine(dimStyle.Render(localize(localizer, "git_mode_select_help", nil)))
	gitModeOptions := []selectOption{
		{
			Value:       "manual",
			Display:     localize(localizer, "git_mode_manual_display", nil),
			Description: localize(localizer, "git_mode_manual_desc", nil),
		},
		{
			Value:       "personal",
			Display:     localize(localizer, "git_mode_personal_display", nil),
			Description: localize(localizer, "git_mode_personal_desc", nil),
		},
		{
			Value:       "team",
			Display:     localize(localizer, "git_mode_team_display", nil),
			Description: localize(localizer, "git_mode_team_desc", nil),
		},
	}
	answers.GitMode = promptSelectWithDescriptionWithLabel(gitModeOptions, localize(localizer, "select_git_mode_summary", nil), answers.GitMode)

	if answers.GitMode == "personal" || answers.GitMode == "team" {
		answers.GitHubUser = promptInputRequiredWithLabel(localize(localizer, "github_prompt", nil), localize(localizer, "github_label", nil), answers.GitHubUser)
	}

	// Output Language Settings section
	ui.PrintBorderedSection(localize(localizer, "output_language_settings", nil))
	answers.GitCommitLang = promptSelectLanguageWithLabel(
		localize(localizer, "commit_lang_prompt", nil),
		localize(localizer, "commit_lang_label", nil),
		languageOptions, answers.GitCommitLang)
	answers.CodeCommentLang = promptSelectLanguageWithLabel(
		localize(localizer, "code_comment_prompt", nil),
		localize(localizer, "code_comment_label", nil),
		languageOptions, answers.CodeCommentLang)
	answers.DocLang = promptSelectLanguageWithLabel(
		localize(localizer, "documentation_prompt", nil),
		localize(localizer, "documentation_label", nil),
		languageOptions, answers.DocLang)

	// TAG System Setup
	ui.PrintBorderedSection(localize(localizer, "tag_setup_section", nil))
	ui.PrintBorderedLine(dimStyle.Render(localize(localizer, "tag_setup_desc", nil)))
	tagEnabledOptions := []selectOption{
		{
			Value:       "yes",
			Display:     localize(localizer, "tag_enabled_yes", nil),
			Description: localize(localizer, "tag_enabled_yes_desc", nil),
		},
		{
			Value:       "no",
			Display:     localize(localizer, "tag_enabled_no", nil),
			Description: localize(localizer, "tag_enabled_no_desc", nil),
		},
	}
	tagEnabledValue := "yes"
	if !answers.TagEnabled {
		tagEnabledValue = "no"
	}
	tagEnabledResult := promptSelectWithDescriptionWithLabel(tagEnabledOptions, localize(localizer, "tag_enabled_label", nil), tagEnabledValue)
	answers.TagEnabled = tagEnabledResult == "yes"

	if answers.TagEnabled {
		ui.PrintBorderedLine(dimStyle.Render(localize(localizer, "tag_mode_select_help", nil)))
		tagModeOptions := []selectOption{
			{
				Value:       "warn",
				Display:     localize(localizer, "tag_mode_warn_display", nil),
				Description: localize(localizer, "tag_mode_warn_desc", nil),
			},
			{
				Value:       "enforce",
				Display:     localize(localizer, "tag_mode_enforce_display", nil),
				Description: localize(localizer, "tag_mode_enforce_desc", nil),
			},
			{
				Value:       "off",
				Display:     localize(localizer, "tag_mode_off_display", nil),
				Description: localize(localizer, "tag_mode_off_desc", nil),
			},
		}
		answers.TagMode = promptSelectWithDescriptionWithLabel(tagModeOptions, localize(localizer, "tag_mode_label", nil), answers.TagMode)
	} else {
		answers.TagMode = "off"
	}

	// TAG Summary Box
	var tagLines []string
	if answers.TagEnabled {
		tagModeDisplay := answers.TagMode
		switch answers.TagMode {
		case "warn":
			tagModeDisplay = localize(localizer, "tag_mode_warn_display", nil)
		case "enforce":
			tagModeDisplay = localize(localizer, "tag_mode_enforce_display", nil)
		case "off":
			tagModeDisplay = localize(localizer, "tag_mode_off_display", nil)
		}
		tagLines = []string{
			successStyle.Render("✓ ") + dimStyle.Render("Enabled: ") + whiteStyle.Render("Yes"),
			successStyle.Render("✓ ") + dimStyle.Render("Mode: ") + whiteStyle.Render(tagModeDisplay),
		}
	} else {
		tagLines = []string{
			successStyle.Render("✓ ") + dimStyle.Render("Enabled: ") + whiteStyle.Render("No"),
		}
	}
	ui.PrintBorderedBox("TAG SYSTEM", tagLines, 56)

	return answers, nil
}


// promptInputWithLabel prompts for input and shows a success message with the setting name
func promptInputWithLabel(prompt, settingLabel, current string) string {
	borderPrefix := "\033[38;5;30m  │ \033[0m"
	p := promptui.Prompt{
		Label:     prompt,
		Default:   current,
		AllowEdit: true,
		Templates: &promptui.PromptTemplates{
			Prompt:  borderPrefix + "{{ . | cyan }}: ",
			Valid:   borderPrefix + "{{ . | cyan }}: ",
			Invalid: borderPrefix + "{{ . | red }}: ",
			Success: borderPrefix,
		},
	}
	if result, err := p.Run(); err == nil {
		trimmed := strings.TrimSpace(result)
		// Clear the previous line (promptui's input echo) and print success message
		fmt.Print("\033[A\033[K")
		if trimmed != "" {
			ui.PrintBorderedResult(settingLabel, trimmed)
		} else {
			fmt.Println()
		}
		return trimmed
	} else if handleInterrupt(err) {
		os.Exit(1)
	}
	return current
}

// promptInputRequiredWithLabel prompts for required input and shows a success message
func promptInputRequiredWithLabel(prompt, settingLabel, current string) string {
	borderPrefix := "\033[38;5;30m  │ \033[0m"
	p := promptui.Prompt{
		Label:     prompt,
		Default:   current,
		AllowEdit: true,
		Templates: &promptui.PromptTemplates{
			Prompt:  borderPrefix + "{{ . | cyan }}: ",
			Valid:   borderPrefix + "{{ . | cyan }}: ",
			Invalid: borderPrefix + "{{ . | red }}: ",
			Success: borderPrefix,
		},
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("this field is required")
			}
			return nil
		},
	}
	for {
		if result, err := p.Run(); err == nil {
			trimmed := strings.TrimSpace(result)
			if trimmed != "" {
				// Clear the previous line (promptui's input echo) and print success message
				fmt.Print("\033[A\033[K")
				ui.PrintBorderedResult(settingLabel, trimmed)
				return trimmed
			}
		} else if handleInterrupt(err) {
			os.Exit(1)
		}
	}
}

func promptSelectLanguage(label string, choices []languageOption, current string) string {
	type option struct {
		Display string
		Value   string
	}

	var items []option
	for _, choice := range choices {
		items = append(items, option{Display: choice.display, Value: choice.value})
	}

	borderPrefix := "\033[38;5;30m  │ \033[0m"
	templates := &promptui.SelectTemplates{
		Label:    borderPrefix + "{{ . | cyan }}",
		Active:   "\033[38;5;30m│\033[0m \033[38;5;201m❯\033[0m \033[38;5;226m{{ .Display }}\033[0m",
		Inactive: "\033[38;5;30m│\033[0m    {{ .Display | faint }}",
		Selected: borderPrefix,
	}

	prompt := promptui.Select{
		Label:     label + ":",
		Items:     items,
		Templates: templates,
		Size:      4,
		HideHelp:  true,
	}

	if i, _, err := prompt.Run(); err == nil {
		// Clear the Selected line and print bordered result
		fmt.Print("\033[A\033[K")
		return items[i].Value
	} else if handleInterrupt(err) {
		os.Exit(1)
	}
	return current
}

// promptSelectLanguageWithLabel shows a contextual success message with the setting name
func promptSelectLanguageWithLabel(prompt, settingLabel string, choices []languageOption, current string) string {
	type option struct {
		Display string
		Value   string
	}

	var items []option
	for _, choice := range choices {
		items = append(items, option{Display: choice.display, Value: choice.value})
	}

	borderPrefix := "\033[38;5;30m  │ \033[0m"
	templates := &promptui.SelectTemplates{
		Label:    borderPrefix + "{{ . | cyan }}",
		Active:   "\033[38;5;30m│\033[0m \033[38;5;201m❯\033[0m \033[38;5;226m{{ .Display }}\033[0m",
		Inactive: "\033[38;5;30m│\033[0m    {{ .Display | faint }}",
		Selected: borderPrefix,
	}

	sel := promptui.Select{
		Label:     prompt + ":",
		Items:     items,
		Templates: templates,
		Size:      4,
		HideHelp:  true,
	}

	if i, _, err := sel.Run(); err == nil {
		selected := items[i]
		// Clear the previous line (promptui's selected echo) and print success message
		fmt.Print("\033[A\033[K")
		ui.PrintBorderedResult(settingLabel, selected.Display)
		return selected.Value
	} else if handleInterrupt(err) {
		os.Exit(1)
	}
	return current
}

type selectOption struct {
	Value       string
	Display     string
	Description string
}

// promptSelectWithDescriptionWithLabel shows a contextual success message with the setting name
func promptSelectWithDescriptionWithLabel(options []selectOption, settingLabel, current string) string {
	type displayOption struct {
		Value       string
		Display     string
		Description string
	}

	var items []displayOption
	for _, opt := range options {
		items = append(items, displayOption{
			Value:       opt.Value,
			Display:     opt.Display,
			Description: opt.Description,
		})
	}

	borderPrefix := "\033[38;5;30m  │ \033[0m"
	templates := &promptui.SelectTemplates{
		Label:    borderPrefix,
		Active:   "\033[38;5;30m│\033[0m \033[38;5;201m❯\033[0m \033[38;5;226m{{ .Display }}\033[0m \033[38;5;242m{{ .Description }}\033[0m",
		Inactive: "\033[38;5;30m│\033[0m    {{ .Display }} \033[38;5;242m{{ .Description }}\033[0m",
		Selected: borderPrefix,
	}

	prompt := promptui.Select{
		Label:     " ",
		Items:     items,
		Templates: templates,
		Size:      len(items),
		HideHelp:  true,
	}

	if i, _, err := prompt.Run(); err == nil {
		selected := items[i]
		// Clear the previous line (promptui's selected echo) and print success message
		fmt.Print("\033[A\033[K")
		ui.PrintBorderedResult(settingLabel, selected.Display)
		return selected.Value
	} else if handleInterrupt(err) {
		os.Exit(1)
	}
	return current
}

type initOptions struct {
	nonInteractive bool
	mode           string
	locale         string
	language       string
	force          bool
}

func handleInterrupt(err error) bool {
	if err == promptui.ErrInterrupt {
		fmt.Println()
		fmt.Println(warningStyle.Render("    ⚠ ABORT: User cancelled initialization"))
		return true
	}
	return false
}

func localize(localizer *i18n.Localizer, messageID string, data map[string]interface{}) string {
	if localizer == nil {
		return messageID
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return result
}

func displayFinalResult(result project.InitializeResult, answers project.SetupAnswers, duration time.Duration, localizer *i18n.Localizer) {
	// Completion banner
	banner.PrintInitComplete()

	// Summary details
	labelStyle := lipgloss.NewStyle().Foreground(dimCyan).Width(18)

	formatRow := func(label, value string) string {
		return labelStyle.Render(label+":") + " " + whiteStyle.Render(value)
	}

	// Git mode display
	gitModeDisplay := answers.GitMode
	if answers.GitMode == "manual" {
		gitModeDisplay = fmt.Sprintf("%s (%s)", answers.GitMode, localize(localizer, "local_only", nil))
	} else if answers.GitMode == "personal" {
		gitModeDisplay = fmt.Sprintf("%s (%s)", answers.GitMode, localize(localizer, "github_personal", nil))
	} else if answers.GitMode == "team" {
		gitModeDisplay = fmt.Sprintf("%s (%s)", answers.GitMode, localize(localizer, "github_team", nil))
	}

	// TAG System display
	tagDisplay := localize(localizer, "no", nil)
	if answers.TagEnabled {
		tagModeDisplay := answers.TagMode
		switch answers.TagMode {
		case "warn":
			tagModeDisplay = localize(localizer, "tag_mode_warn_display", nil)
		case "enforce":
			tagModeDisplay = localize(localizer, "tag_mode_enforce_display", nil)
		case "off":
			tagModeDisplay = localize(localizer, "tag_mode_off_display", nil)
		}
		tagDisplay = fmt.Sprintf("%s (%s)", localize(localizer, "yes", nil), tagModeDisplay)
	}

	summaryLines := []string{
		formatRow(localize(localizer, "location", nil), result.ProjectPath),
		formatRow(localize(localizer, "language", nil), languageDisplay(answers.Locale)),
		formatRow(localize(localizer, "git", nil), gitModeDisplay),
		formatRow(localize(localizer, "tag_enabled_label", nil), tagDisplay),
		formatRow(localize(localizer, "files", nil), fmt.Sprintf("%d %s", len(result.CreatedFiles), localize(localizer, "created", nil))),
		formatRow(localize(localizer, "duration", nil), duration.Round(time.Millisecond).String()),
	}

	if result.Reinitialized {
		summaryLines = append(summaryLines, formatRow(localize(localizer, "backup", nil), ".jikime/backups/"))
	}

	ui.PrintBorderedBox("PROJECT SUMMARY", summaryLines, 70)

	// Next steps
	stepStyle := lipgloss.NewStyle().Foreground(magenta).Bold(true)
	nextStepLines := []string{
		stepStyle.Render("[1]") + " " + whiteStyle.Render(localize(localizer, "next_step_1", nil)),
		"    " + dimStyle.Render(localize(localizer, "next_step_1_desc", nil)),
		stepStyle.Render("[2]") + " " + whiteStyle.Render(localize(localizer, "next_step_2", nil)),
	}

	ui.PrintBorderedBox("NEXT STEPS", nextStepLines, 70)
	fmt.Println()

	// Final message
	fmt.Println(systemStyle.Render("    [SYS]") + whiteStyle.Render(" System ready. Happy coding!"))
	fmt.Println()
}

