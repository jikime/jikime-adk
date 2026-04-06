// Package skill provides Skill Registry for jikime-adk.
package skill

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Registry manages a collection of skills with indexing for fast lookup.
type Registry struct {
	mu sync.RWMutex

	// Primary storage: name -> Skill
	skills map[string]*Skill

	// Indexes for fast lookup
	byTag      map[string][]*Skill // tag -> skills with that tag
	byKeyword  map[string][]*Skill // keyword -> skills triggered by that keyword
	byPhase    map[string][]*Skill // phase -> skills active in that phase
	byAgent    map[string][]*Skill // agent -> skills used by that agent
	byLanguage map[string][]*Skill // language -> skills supporting that language
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		skills:     make(map[string]*Skill),
		byTag:      make(map[string][]*Skill),
		byKeyword:  make(map[string][]*Skill),
		byPhase:    make(map[string][]*Skill),
		byAgent:    make(map[string][]*Skill),
		byLanguage: make(map[string][]*Skill),
	}
}

// Register adds a skill to the registry and updates indexes.
func (r *Registry) Register(skill *Skill) {
	if skill == nil || skill.Name == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Add to primary storage
	r.skills[skill.Name] = skill

	// Update tag index
	for _, tag := range skill.Tags {
		r.byTag[tag] = append(r.byTag[tag], skill)
	}

	// Update trigger indexes
	for _, keyword := range skill.Triggers.Keywords {
		lowerKeyword := strings.ToLower(keyword)
		r.byKeyword[lowerKeyword] = append(r.byKeyword[lowerKeyword], skill)
	}

	for _, phase := range skill.Triggers.Phases {
		r.byPhase[phase] = append(r.byPhase[phase], skill)
	}

	for _, agent := range skill.Triggers.Agents {
		r.byAgent[agent] = append(r.byAgent[agent], skill)
	}

	for _, language := range skill.Triggers.Languages {
		lowerLang := strings.ToLower(language)
		r.byLanguage[lowerLang] = append(r.byLanguage[lowerLang], skill)
	}
}

// RegisterAll adds multiple skills to the registry.
func (r *Registry) RegisterAll(skills []*Skill) {
	for _, skill := range skills {
		r.Register(skill)
	}
}

// Get retrieves a skill by name.
func (r *Registry) Get(name string) *Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.skills[name]
}

// All returns all registered skills.
func (r *Registry) All() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills
}

// AllSorted returns all skills sorted by name.
func (r *Registry) AllSorted() []*Skill {
	skills := r.All()
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills
}

// Count returns the number of registered skills.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}

// GetByTag returns all skills with a specific tag.
func (r *Registry) GetByTag(tag string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byTag[tag]
}

// GetByKeyword returns all skills triggered by a keyword.
func (r *Registry) GetByKeyword(keyword string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byKeyword[strings.ToLower(keyword)]
}

// GetByPhase returns all skills active in a specific phase.
func (r *Registry) GetByPhase(phase string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byPhase[phase]
}

// GetByAgent returns all skills used by a specific agent.
func (r *Registry) GetByAgent(agent string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byAgent[agent]
}

// GetByLanguage returns all skills supporting a specific language.
func (r *Registry) GetByLanguage(language string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byLanguage[strings.ToLower(language)]
}

// GetRelated returns skills related to a given skill (sharing tags or triggers).
func (r *Registry) GetRelated(skillName string, maxResults int) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill := r.skills[skillName]
	if skill == nil {
		return nil
	}

	// Score skills by relationship strength
	scores := make(map[string]int)

	// Score by shared tags
	for _, tag := range skill.Tags {
		for _, related := range r.byTag[tag] {
			if related.Name != skillName {
				scores[related.Name] += 2 // Tags are strong relationship
			}
		}
	}

	// Score by shared phases
	for _, phase := range skill.Triggers.Phases {
		for _, related := range r.byPhase[phase] {
			if related.Name != skillName {
				scores[related.Name]++
			}
		}
	}

	// Score by shared agents
	for _, agent := range skill.Triggers.Agents {
		for _, related := range r.byAgent[agent] {
			if related.Name != skillName {
				scores[related.Name]++
			}
		}
	}

	// Score by shared languages
	for _, lang := range skill.Triggers.Languages {
		for _, related := range r.byLanguage[strings.ToLower(lang)] {
			if related.Name != skillName {
				scores[related.Name]++
			}
		}
	}

	// Sort by score
	type scoredSkill struct {
		name  string
		score int
	}
	var scored []scoredSkill
	for name, score := range scores {
		scored = append(scored, scoredSkill{name, score})
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top results
	var results []*Skill
	for i := 0; i < len(scored) && (maxResults <= 0 || i < maxResults); i++ {
		if s := r.skills[scored[i].name]; s != nil {
			results = append(results, s)
		}
	}

	return results
}

// AllTags returns all unique tags in the registry.
func (r *Registry) AllTags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]string, 0, len(r.byTag))
	for tag := range r.byTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// AllPhases returns all unique phases in the registry.
func (r *Registry) AllPhases() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	phases := make([]string, 0, len(r.byPhase))
	for phase := range r.byPhase {
		phases = append(phases, phase)
	}
	sort.Strings(phases)
	return phases
}

// AllAgents returns all unique agents in the registry.
func (r *Registry) AllAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]string, 0, len(r.byAgent))
	for agent := range r.byAgent {
		agents = append(agents, agent)
	}
	sort.Strings(agents)
	return agents
}

// AllLanguages returns all unique languages in the registry.
func (r *Registry) AllLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	languages := make([]string, 0, len(r.byLanguage))
	for lang := range r.byLanguage {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	return languages
}

// LoadFromProjectRoot loads all skills from a project's skill directories.
// It searches both .claude/skills (for initialized projects) and
// templates/.claude/skills (for development/build environment).
func (r *Registry) LoadFromProjectRoot(projectRoot string) error {
	// Define paths to search for skills
	skillPaths := []string{
		filepath.Join(projectRoot, ".claude", "skills"),
		filepath.Join(projectRoot, "templates", ".claude", "skills"),
	}

	var foundAny bool
	var lastErr error

	for _, skillsDir := range skillPaths {
		// Check if directory exists
		info, err := os.Stat(skillsDir)
		if err != nil || !info.IsDir() {
			continue
		}

		// Load skills from this directory
		skills, err := LoadMetadataFromDirectory(skillsDir, true)
		if err != nil {
			lastErr = err
			continue
		}

		if len(skills) > 0 {
			r.RegisterAll(skills)
			foundAny = true
		}
	}

	if !foundAny && lastErr != nil {
		return lastErr
	}

	return nil
}

// Clear removes all skills and indexes from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.skills = make(map[string]*Skill)
	r.byTag = make(map[string][]*Skill)
	r.byKeyword = make(map[string][]*Skill)
	r.byPhase = make(map[string][]*Skill)
	r.byAgent = make(map[string][]*Skill)
	r.byLanguage = make(map[string][]*Skill)
}

// DefaultRegistry is the global default registry.
var DefaultRegistry = NewRegistry()
