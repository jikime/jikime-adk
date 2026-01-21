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
	"jikime-adk-v2/cmd/banner"
	"jikime-adk-v2/project"
	"jikime-adk-v2/version"
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

func getGitModeOption(options []selectOption, value string) selectOption {
	for _, opt := range options {
		if opt.Value == value {
			return opt
		}
	}
	return selectOption{Value: value, Display: value, Description: ""}
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
			fmt.Println()

			// Starting installation
			fmt.Println(sectionStyle.Render("    ░▒▓ ") + highlightStyle.Render("INSTALLATION SEQUENCE") + sectionStyle.Render(" ▓▒░"))
			fmt.Println()

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
		fmt.Printf("\r              %s %s", bar, percent)
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
	printCyberSection("🌐", localize(localizer, "language_section", nil))
	fmt.Println(dimStyle.Render("    " + localize(localizer, "language_prompt", nil)))
	answers.Locale = promptSelectLanguage(localize(localizer, "language_prompt", nil), languageOptions, answers.Locale)

	localizer = getTranslation(answers.Locale)
	fmt.Println(successStyle.Render("    ✓ ") + whiteStyle.Render(localize(localizer, "language_selected", map[string]interface{}{"Language": languageDisplay(answers.Locale)})))

	// User Section
	printCyberSection("👤", localize(localizer, "user_section", nil))
	answers.UserName = promptInput("    "+localize(localizer, "user_prompt", nil), answers.UserName)
	if answers.UserName != "" {
		fmt.Println(successStyle.Render("    ✓ ") + whiteStyle.Render(localize(localizer, "user_welcome", map[string]interface{}{"Name": answers.UserName})))
	}

	// Project name
	answers.ProjectName = promptInput("    "+localize(localizer, "project_prompt", nil), answers.ProjectName)

	// Git Setup section
	printCyberSection("🔀", localize(localizer, "git_setup_section", nil))
	fmt.Println(dimStyle.Render("    " + localize(localizer, "git_mode_select_help", nil)))
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
	answers.GitMode = promptSelectWithDescription(gitModeOptions, answers.GitMode)

	if answers.GitMode == "personal" || answers.GitMode == "team" {
		answers.GitHubUser = promptInputRequired("    "+localize(localizer, "github_prompt", nil), answers.GitHubUser)
	}

	answers.GitCommitLang = promptSelectLanguage(localize(localizer, "commit_lang_prompt", nil), languageOptions, answers.GitCommitLang)
	answers.CodeCommentLang = promptSelectLanguage(localize(localizer, "code_comment_prompt", nil), languageOptions, answers.CodeCommentLang)
	answers.DocLang = promptSelectLanguage(localize(localizer, "documentation_prompt", nil), languageOptions, answers.DocLang)

	// Git Summary
	fmt.Println()
	fmt.Println(sectionStyle.Render("    ─── ") + highlightStyle.Render("GIT CONFIGURATION") + sectionStyle.Render(" ───"))
	gitModeOption := getGitModeOption(gitModeOptions, answers.GitMode)
	fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Mode: ") + whiteStyle.Render(gitModeOption.Display))

	// Output Language Summary
	fmt.Println()
	fmt.Println(sectionStyle.Render("    ─── ") + highlightStyle.Render("OUTPUT LANGUAGES") + sectionStyle.Render(" ───"))
	fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Commits: ") + whiteStyle.Render(languageDisplay(answers.GitCommitLang)))
	fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Comments: ") + whiteStyle.Render(languageDisplay(answers.CodeCommentLang)))
	fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Docs: ") + whiteStyle.Render(languageDisplay(answers.DocLang)))

	// TAG System Setup
	printCyberSection("🏷️", localize(localizer, "tag_setup_section", nil))
	fmt.Println(dimStyle.Render("    " + localize(localizer, "tag_setup_desc", nil)))

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
	tagEnabledResult := promptSelectWithDescription(tagEnabledOptions, tagEnabledValue)
	answers.TagEnabled = tagEnabledResult == "yes"

	if answers.TagEnabled {
		fmt.Println(dimStyle.Render("    " + localize(localizer, "tag_mode_select_help", nil)))
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
		answers.TagMode = promptSelectWithDescription(tagModeOptions, answers.TagMode)
	} else {
		answers.TagMode = "off"
	}

	// TAG Summary
	fmt.Println()
	fmt.Println(sectionStyle.Render("    ─── ") + highlightStyle.Render("TAG SYSTEM") + sectionStyle.Render(" ───"))
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
		fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Enabled: ") + whiteStyle.Render("Yes"))
		fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Mode: ") + whiteStyle.Render(tagModeDisplay))
	} else {
		fmt.Println(successStyle.Render("    ✓ ") + dimStyle.Render("Enabled: ") + whiteStyle.Render("No"))
	}

	return answers, nil
}

func printCyberSection(icon, title string) {
	fmt.Println()
	header := sectionStyle.Render("    ░▒▓ ") +
		lipgloss.NewStyle().Render(icon+" ") +
		highlightStyle.Render(strings.ToUpper(title)) +
		sectionStyle.Render(" ▓▒░")
	fmt.Println(header)
}

func promptInput(label, current string) string {
	prompt := promptui.Prompt{
		Label:     label,
		Default:   current,
		AllowEdit: true,
		Templates: &promptui.PromptTemplates{
			Prompt:  "{{ . | cyan }}",
			Valid:   "{{ . | cyan }}",
			Invalid: "{{ . | red }}",
			Success: "{{ . | green }}",
		},
	}
	if result, err := prompt.Run(); err == nil {
		return strings.TrimSpace(result)
	} else if handleInterrupt(err) {
		os.Exit(1)
	}
	return current
}

func promptInputRequired(label, current string) string {
	prompt := promptui.Prompt{
		Label:     label,
		Default:   current,
		AllowEdit: true,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("this field is required")
			}
			return nil
		},
	}
	for {
		if result, err := prompt.Run(); err == nil {
			trimmed := strings.TrimSpace(result)
			if trimmed != "" {
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
		Icon    string
	}

	var items []option
	for _, choice := range choices {
		icon := "  "
		if choice.value == current {
			icon = "▶"
		}
		items = append(items, option{Display: choice.display, Value: choice.value, Icon: icon})
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . | cyan }}",
		Active:   "\033[38;5;201m❯\033[0m {{ .Icon }} \033[38;5;226m{{ .Display }}\033[0m",
		Inactive: "  {{ .Display | faint }}",
		Selected: "\033[38;5;46m✓\033[0m {{ .Display | green }}",
	}

	prompt := promptui.Select{
		Label:     fmt.Sprintf("    %s:", label),
		Items:     items,
		Templates: templates,
		Size:      4,
	}

	if i, _, err := prompt.Run(); err == nil {
		return items[i].Value
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

func promptSelectWithDescription(options []selectOption, current string) string {
	type displayOption struct {
		Value       string
		Display     string
		Description string
		Icon        string
	}

	var items []displayOption
	for _, opt := range options {
		icon := "  "
		if opt.Value == current {
			icon = "▶"
		}
		items = append(items, displayOption{
			Value:       opt.Value,
			Display:     opt.Display,
			Description: opt.Description,
			Icon:        icon,
		})
	}

	templates := &promptui.SelectTemplates{
		Active:   "\033[38;5;201m❯\033[0m {{ .Icon }} \033[38;5;226m{{ .Display }}\033[0m \033[38;5;242m{{ .Description }}\033[0m",
		Inactive: "  {{ .Display }} \033[38;5;242m{{ .Description }}\033[0m",
		Selected: "\033[38;5;46m✓\033[0m {{ .Display | green }}",
	}

	prompt := promptui.Select{
		Label:     "",
		Items:     items,
		Templates: templates,
		Size:      len(items),
	}

	if i, _, err := prompt.Run(); err == nil {
		return items[i].Value
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

func formatPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		rel := strings.TrimPrefix(path, home)
		if rel == "" {
			return "~"
		}
		if rel[0] != '/' {
			return "~/" + rel
		}
		return "~" + rel
	}
	return path
}

func displayFinalResult(result project.InitializeResult, answers project.SetupAnswers, duration time.Duration, localizer *i18n.Localizer) {
	// Completion banner
	banner.PrintInitComplete()

	// Summary box
	fmt.Println(sectionStyle.Render("    ╔══════════════════════════════════════════════════════════╗"))
	fmt.Println(sectionStyle.Render("    ║") + highlightStyle.Render("                    PROJECT SUMMARY                       ") + sectionStyle.Render("║"))
	fmt.Println(sectionStyle.Render("    ╚══════════════════════════════════════════════════════════╝"))
	fmt.Println()

	// Summary details
	printSummaryRow("📁", localize(localizer, "location", nil), result.ProjectPath)
	printSummaryRow("🌐", localize(localizer, "language", nil), languageDisplay(answers.Locale))

	// Git mode display
	gitModeDisplay := answers.GitMode
	if answers.GitMode == "manual" {
		gitModeDisplay = fmt.Sprintf("%s (%s)", answers.GitMode, localize(localizer, "local_only", nil))
	} else if answers.GitMode == "personal" {
		gitModeDisplay = fmt.Sprintf("%s (%s)", answers.GitMode, localize(localizer, "github_personal", nil))
	} else if answers.GitMode == "team" {
		gitModeDisplay = fmt.Sprintf("%s (%s)", answers.GitMode, localize(localizer, "github_team", nil))
	}
	printSummaryRow("🔀", localize(localizer, "git", nil), gitModeDisplay)

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
	printSummaryRow("🏷️", localize(localizer, "tag_enabled_label", nil), tagDisplay)

	printSummaryRow("📄", localize(localizer, "files", nil), fmt.Sprintf("%d %s", len(result.CreatedFiles), localize(localizer, "created", nil)))
	printSummaryRow("⏱️", localize(localizer, "duration", nil), duration.Round(time.Millisecond).String())

	if result.Reinitialized {
		printSummaryRow("💾", localize(localizer, "backup", nil), ".jikime/backups/")
	}

	// Next steps
	fmt.Println()
	fmt.Println(sectionStyle.Render("    ╔══════════════════════════════════════════════════════════╗"))
	fmt.Println(sectionStyle.Render("    ║") + highlightStyle.Render("                      NEXT STEPS                         ") + sectionStyle.Render("║"))
	fmt.Println(sectionStyle.Render("    ╚══════════════════════════════════════════════════════════╝"))
	fmt.Println()

	stepStyle := lipgloss.NewStyle().Foreground(magenta).Bold(true)
	fmt.Printf("    %s %s\n", stepStyle.Render("[1]"), whiteStyle.Render(localize(localizer, "next_step_1", nil)))
	fmt.Println(dimStyle.Render("        " + localize(localizer, "next_step_1_desc", nil)))
	fmt.Printf("    %s %s\n", stepStyle.Render("[2]"), whiteStyle.Render(localize(localizer, "next_step_2", nil)))
	fmt.Println()

	// Final message
	fmt.Println(systemStyle.Render("    [SYS]") + whiteStyle.Render(" System ready. Happy coding! 🚀"))
	fmt.Println()
}

func printSummaryRow(icon, label, value string) {
	labelStyle := lipgloss.NewStyle().
		Foreground(dimCyan).
		Width(18)

	fmt.Printf("    %s %s %s\n",
		icon,
		labelStyle.Render(label+":"),
		whiteStyle.Render(value),
	)
}
