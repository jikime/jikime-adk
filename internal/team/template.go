package team

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TemplateStore loads TemplateDef objects from a directory of YAML files.
type TemplateStore struct {
	dirs []string // searched in order; later dirs override earlier ones
}

// NewTemplateStore returns a TemplateStore that searches the given directories.
// Built-in templates should come first so user templates can override them.
func NewTemplateStore(dirs ...string) *TemplateStore {
	return &TemplateStore{dirs: dirs}
}

// Load returns the TemplateDef for the given name, searching all template
// directories. Returns an error if no template with that name is found.
func (ts *TemplateStore) Load(name string) (*TemplateDef, error) {
	// Search in reverse so later (user) dirs take precedence.
	for i := len(ts.dirs) - 1; i >= 0; i-- {
		path := filepath.Join(ts.dirs[i], name+".yaml")
		def, err := loadYAML(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("team/template: load %s: %w", path, err)
		}
		if err := validate(def); err != nil {
			return nil, fmt.Errorf("team/template: validate %s: %w", name, err)
		}
		return def, nil
	}
	return nil, fmt.Errorf("team/template: template %q not found in %v", name, ts.dirs)
}

// List returns a summary of all available templates (name + description).
// Templates with the same name in later directories shadow earlier ones.
func (ts *TemplateStore) List() ([]TemplateDef, error) {
	seen := map[string]bool{}
	var results []TemplateDef

	// Reverse so user templates take precedence.
	for i := len(ts.dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(ts.dirs[i])
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("team/template: readdir %s: %w", ts.dirs[i], err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			name := e.Name()[:len(e.Name())-5]
			if seen[name] {
				continue
			}
			def, err := loadYAML(filepath.Join(ts.dirs[i], e.Name()))
			if err != nil {
				continue
			}
			seen[name] = true
			results = append(results, *def)
		}
	}
	return results, nil
}

// Render expands placeholders in agent task prompts and returns a new TemplateDef.
// Placeholders: {{goal}}, {{team_name}}, {{agent_id}}, {{leader_id}}.
func Render(def *TemplateDef, goal, teamName string) *TemplateDef {
	leaderID := ""
	for _, a := range def.Agents {
		if a.Role == "leader" {
			leaderID = a.ID
			break
		}
	}

	cp := *def
	agents := make([]TemplateAgentDef, len(def.Agents))
	for i, a := range def.Agents {
		a.SystemPromptFile = expandVars(a.SystemPromptFile, goal, teamName, a.ID, leaderID)
		a.Task = expandVars(a.Task, goal, teamName, a.ID, leaderID)
		agents[i] = a
	}
	cp.Agents = agents
	return &cp
}

// --- internal helpers ---

func loadYAML(path string) (*TemplateDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var def TemplateDef
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	return &def, nil
}

func validate(def *TemplateDef) error {
	if def.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if len(def.Agents) == 0 {
		return fmt.Errorf("template must define at least one agent")
	}
	for i, a := range def.Agents {
		if a.ID == "" {
			return fmt.Errorf("agent[%d] missing id", i)
		}
		if a.Role == "" {
			return fmt.Errorf("agent[%d] missing role", i)
		}
	}
	return nil
}

// expandVars replaces all supported placeholders in s.
func expandVars(s, goal, teamName, agentID, leaderID string) string {
	r := strings.NewReplacer(
		"{{goal}}", goal,
		"{{team_name}}", teamName,
		"{{agent_id}}", agentID,
		"{{leader_id}}", leaderID,
	)
	return r.Replace(s)
}
