// Package skill provides Skill System for jikime-adk.
// Implements tag-based skill discovery for JikiME-ADK.
package skill

// Version of the Skill system.
const Version = "1.0.0"

// Skill represents a SKILL.md file's metadata.
type Skill struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Triggers    Triggers `yaml:"triggers"`

	// Optional fields
	Type           string   `yaml:"type,omitempty"`           // skill type (e.g., "version", "domain")
	Framework      string   `yaml:"framework,omitempty"`      // framework name (e.g., "nextjs")
	Version        string   `yaml:"version,omitempty"`        // version string
	UserInvocable  bool     `yaml:"user-invocable,omitempty"` // whether user can invoke directly
	Context        string   `yaml:"context,omitempty"`        // context type (e.g., "fork")
	Agent          string   `yaml:"agent,omitempty"`          // agent name
	AllowedTools   []string `yaml:"allowed-tools,omitempty"`  // allowed tools list

	// Internal fields (not from YAML)
	FilePath string `yaml:"-"` // Path to SKILL.md file
	Body     string `yaml:"-"` // Markdown body content (Level 2)
}

// Triggers defines conditions for loading a skill.
type Triggers struct {
	Keywords  []string `yaml:"keywords"`  // Keywords to detect in user input
	Phases    []string `yaml:"phases"`    // Development phases (plan, run, sync)
	Agents    []string `yaml:"agents"`    // Agents that use this skill
	Languages []string `yaml:"languages"` // Programming languages
}

// ValidPhases defines allowed development phases.
var ValidPhases = map[string]bool{
	"plan": true,
	"run":  true,
	"sync": true,
}

// NewSkill creates a new Skill with the given name.
func NewSkill(name string) *Skill {
	return &Skill{
		Name: name,
		Triggers: Triggers{
			Keywords:  []string{},
			Phases:    []string{},
			Agents:    []string{},
			Languages: []string{},
		},
	}
}

// IsValid checks if the Skill has required fields.
func (s *Skill) IsValid() bool {
	return s.Name != "" && s.Description != ""
}

// HasTag checks if the skill has a specific tag.
func (s *Skill) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// HasKeyword checks if the skill triggers on a specific keyword.
func (s *Skill) HasKeyword(keyword string) bool {
	for _, k := range s.Triggers.Keywords {
		if k == keyword {
			return true
		}
	}
	return false
}

// HasPhase checks if the skill is active in a specific phase.
func (s *Skill) HasPhase(phase string) bool {
	for _, p := range s.Triggers.Phases {
		if p == phase {
			return true
		}
	}
	return false
}

// HasAgent checks if the skill is used by a specific agent.
func (s *Skill) HasAgent(agent string) bool {
	for _, a := range s.Triggers.Agents {
		if a == agent {
			return true
		}
	}
	return false
}

// HasLanguage checks if the skill supports a specific language.
func (s *Skill) HasLanguage(language string) bool {
	for _, l := range s.Triggers.Languages {
		if l == language {
			return true
		}
	}
	return false
}

// MatchesTriggers checks if the skill matches any of the given trigger conditions.
func (s *Skill) MatchesTriggers(keywords, phases, agents, languages []string) bool {
	// Check keywords
	for _, k := range keywords {
		if s.HasKeyword(k) {
			return true
		}
	}

	// Check phases
	for _, p := range phases {
		if s.HasPhase(p) {
			return true
		}
	}

	// Check agents
	for _, a := range agents {
		if s.HasAgent(a) {
			return true
		}
	}

	// Check languages
	for _, l := range languages {
		if s.HasLanguage(l) {
			return true
		}
	}

	return false
}

// ToMap converts Skill to a map for JSON serialization.
func (s *Skill) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":        s.Name,
		"description": s.Description,
		"tags":        s.Tags,
		"triggers": map[string]interface{}{
			"keywords":  s.Triggers.Keywords,
			"phases":    s.Triggers.Phases,
			"agents":    s.Triggers.Agents,
			"languages": s.Triggers.Languages,
		},
		"file_path": s.FilePath,
	}
}
