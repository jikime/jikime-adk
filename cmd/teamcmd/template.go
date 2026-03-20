package teamcmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
	embedded "jikime-adk/templates"
)

// templateDirs returns the search paths for team templates.
// Embedded built-in templates are auto-installed to ~/.jikime/templates/
// on first use so users never need to copy files manually.
func templateDirs() []string {
	home, _ := os.UserHomeDir()
	globalDir := filepath.Join(home, ".jikime", "templates")
	ensureEmbeddedTemplates(globalDir)
	return []string{globalDir}
}

// ensureEmbeddedTemplates copies built-in team templates from the embedded FS
// into dir if they are not already present. Existing user files are never overwritten.
func ensureEmbeddedTemplates(dir string) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	const prefix = ".jikime/templates"
	_ = fs.WalkDir(embedded.EmbeddedFS, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		name := filepath.Base(path)
		dest := filepath.Join(dir, name)
		if _, statErr := os.Stat(dest); statErr == nil {
			return nil // already exists — keep user version
		}
		data, readErr := embedded.EmbeddedFS.ReadFile(path)
		if readErr != nil {
			return nil
		}
		_ = os.WriteFile(dest, data, 0o644)
		return nil
	})
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "List and inspect team templates",
	}
	cmd.AddCommand(newTemplateListCmd())
	cmd.AddCommand(newTemplateShowCmd())
	return cmd
}

func newTemplateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available team templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			ts := team.NewTemplateStore(templateDirs()...)
			templates, err := ts.List()
			if err != nil {
				return err
			}
			if len(templates) == 0 {
				fmt.Println("No templates found.")
				fmt.Printf("Add YAML files to ~/.jikime/templates/\n")
				return nil
			}
			fmt.Printf("%-20s  %s\n", "NAME", "DESCRIPTION")
			for _, t := range templates {
				desc := t.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				fmt.Printf("%-20s  %s\n", t.Name, desc)
			}
			return nil
		},
	}
}

func newTemplateShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show template details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ts := team.NewTemplateStore(templateDirs()...)
			def, err := ts.Load(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(def)
			}
			fmt.Printf("Name:    %s\n", def.Name)
			fmt.Printf("Version: %s\n", def.Version)
			fmt.Printf("Desc:    %s\n", def.Description)
			if def.DefaultBudget > 0 {
				fmt.Printf("Budget:  %d tokens\n", def.DefaultBudget)
			}
			fmt.Printf("\nAgents (%d):\n", len(def.Agents))
			for _, a := range def.Agents {
				auto := ""
				if a.AutoSpawn {
					auto = " [auto-spawn]"
				}
				fmt.Printf("  %-12s  role:%-10s%s\n", a.ID, a.Role, auto)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
