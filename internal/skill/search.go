// Package skill provides search functionality for jikime-adk skills.
package skill

import (
	"sort"
	"strings"
)

// SearchQuery represents a search request with multiple criteria.
type SearchQuery struct {
	Text      string   // Free-text search (name, description, keywords)
	Tags      []string // Filter by tags
	Phases    []string // Filter by phases
	Agents    []string // Filter by agents
	Languages []string // Filter by languages
	Limit     int      // Maximum results (0 = unlimited)
}

// SearchResult represents a search result with relevance score.
type SearchResult struct {
	Skill *Skill
	Score float64 // Relevance score (higher is better)
}

// Search performs a search on the registry with the given query.
func (r *Registry) Search(query SearchQuery) []*SearchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*SearchResult

	for _, skill := range r.skills {
		score := calculateScore(skill, query)
		if score > 0 {
			results = append(results, &SearchResult{
				Skill: skill,
				Score: score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results
}

// calculateScore computes relevance score for a skill against a query.
func calculateScore(skill *Skill, query SearchQuery) float64 {
	var score float64
	matchCount := 0

	// Text search (name, description, keywords)
	if query.Text != "" {
		text := strings.ToLower(query.Text)

		// Exact name match (highest score)
		if strings.ToLower(skill.Name) == text {
			score += 100
			matchCount++
		} else if strings.Contains(strings.ToLower(skill.Name), text) {
			// Partial name match
			score += 50
			matchCount++
		}

		// Description match
		if strings.Contains(strings.ToLower(skill.Description), text) {
			score += 20
			matchCount++
		}

		// Keyword match
		for _, keyword := range skill.Triggers.Keywords {
			if strings.Contains(strings.ToLower(keyword), text) {
				score += 30
				matchCount++
				break
			}
		}

		// Tag match
		for _, tag := range skill.Tags {
			if strings.Contains(strings.ToLower(tag), text) {
				score += 15
				matchCount++
				break
			}
		}

		// No text match found
		if matchCount == 0 && query.Text != "" {
			// Check if all filters are also empty
			if len(query.Tags) == 0 && len(query.Phases) == 0 &&
				len(query.Agents) == 0 && len(query.Languages) == 0 {
				return 0 // No match at all
			}
		}
	}

	// Tag filter
	if len(query.Tags) > 0 {
		tagMatch := false
		for _, queryTag := range query.Tags {
			if skill.HasTag(queryTag) {
				score += 25
				tagMatch = true
			}
		}
		if !tagMatch && len(query.Tags) > 0 {
			return 0 // Required tag filter not matched
		}
	}

	// Phase filter
	if len(query.Phases) > 0 {
		phaseMatch := false
		for _, phase := range query.Phases {
			if skill.HasPhase(phase) {
				score += 20
				phaseMatch = true
			}
		}
		if !phaseMatch && len(query.Phases) > 0 {
			return 0 // Required phase filter not matched
		}
	}

	// Agent filter
	if len(query.Agents) > 0 {
		agentMatch := false
		for _, agent := range query.Agents {
			if skill.HasAgent(agent) {
				score += 20
				agentMatch = true
			}
		}
		if !agentMatch && len(query.Agents) > 0 {
			return 0 // Required agent filter not matched
		}
	}

	// Language filter
	if len(query.Languages) > 0 {
		langMatch := false
		for _, lang := range query.Languages {
			if skill.HasLanguage(lang) {
				score += 15
				langMatch = true
			}
		}
		if !langMatch && len(query.Languages) > 0 {
			return 0 // Required language filter not matched
		}
	}

	// If no criteria specified, return base score for all skills
	if query.Text == "" && len(query.Tags) == 0 && len(query.Phases) == 0 &&
		len(query.Agents) == 0 && len(query.Languages) == 0 {
		return 1 // Base score for listing all
	}

	return score
}

// SearchByText performs a simple text search.
func (r *Registry) SearchByText(text string, limit int) []*SearchResult {
	return r.Search(SearchQuery{
		Text:  text,
		Limit: limit,
	})
}

// FindByTriggers finds skills that match any of the given trigger conditions.
// This is the primary method for skill discovery during Claude conversations.
func (r *Registry) FindByTriggers(keywords, phases, agents, languages []string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var results []*Skill

	// Find by keywords
	for _, keyword := range keywords {
		lower := strings.ToLower(keyword)
		for _, skill := range r.byKeyword[lower] {
			if !seen[skill.Name] {
				seen[skill.Name] = true
				results = append(results, skill)
			}
		}
	}

	// Find by phases
	for _, phase := range phases {
		for _, skill := range r.byPhase[phase] {
			if !seen[skill.Name] {
				seen[skill.Name] = true
				results = append(results, skill)
			}
		}
	}

	// Find by agents
	for _, agent := range agents {
		for _, skill := range r.byAgent[agent] {
			if !seen[skill.Name] {
				seen[skill.Name] = true
				results = append(results, skill)
			}
		}
	}

	// Find by languages
	for _, lang := range languages {
		lower := strings.ToLower(lang)
		for _, skill := range r.byLanguage[lower] {
			if !seen[skill.Name] {
				seen[skill.Name] = true
				results = append(results, skill)
			}
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// MatchInput matches user input against skill triggers using fuzzy keyword matching.
// Returns skills sorted by match quality.
func (r *Registry) MatchInput(input string) []*SearchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	input = strings.ToLower(input)
	words := strings.Fields(input)

	scores := make(map[string]float64)

	for _, skill := range r.skills {
		score := 0.0

		// Check keywords
		for _, keyword := range skill.Triggers.Keywords {
			lowerKeyword := strings.ToLower(keyword)

			// Exact word match
			for _, word := range words {
				if word == lowerKeyword {
					score += 30
				} else if strings.Contains(word, lowerKeyword) || strings.Contains(lowerKeyword, word) {
					score += 15
				}
			}

			// Phrase match
			if strings.Contains(input, lowerKeyword) {
				score += 20
			}
		}

		if score > 0 {
			scores[skill.Name] = score
		}
	}

	// Convert to results
	var results []*SearchResult
	for name, score := range scores {
		results = append(results, &SearchResult{
			Skill: r.skills[name],
			Score: score,
		})
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// GroupByTag groups all skills by their tags.
func (r *Registry) GroupByTag() map[string][]*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Deep copy to avoid exposing internal state
	result := make(map[string][]*Skill)
	for tag, skills := range r.byTag {
		result[tag] = append([]*Skill{}, skills...)
	}
	return result
}

// GroupByPhase groups all skills by their phases.
func (r *Registry) GroupByPhase() map[string][]*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*Skill)
	for phase, skills := range r.byPhase {
		result[phase] = append([]*Skill{}, skills...)
	}
	return result
}

// Stats returns statistics about the registry.
type Stats struct {
	TotalSkills    int
	TotalTags      int
	TotalKeywords  int
	TotalPhases    int
	TotalAgents    int
	TotalLanguages int
}

// GetStats returns registry statistics.
func (r *Registry) GetStats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return Stats{
		TotalSkills:    len(r.skills),
		TotalTags:      len(r.byTag),
		TotalKeywords:  len(r.byKeyword),
		TotalPhases:    len(r.byPhase),
		TotalAgents:    len(r.byAgent),
		TotalLanguages: len(r.byLanguage),
	}
}
