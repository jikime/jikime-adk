package project

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AnnouncementData represents the structure of announcement JSON files
type AnnouncementData struct {
	CompanyAnnouncements []string `json:"companyAnnouncements"`
}

// LoadAnnouncements loads announcements for the specified language
func LoadAnnouncements(templateRoot, language string) ([]string, error) {
	// Try requested language first
	announcementsPath := filepath.Join(templateRoot, ".jikime", "announcements", language+".json")
	announcements, err := loadAnnouncementsFromFile(announcementsPath)
	if err == nil && len(announcements) > 0 {
		return announcements, nil
	}

	// Fallback to English
	if language != "en" {
		englishPath := filepath.Join(templateRoot, ".jikime", "announcements", "en.json")
		announcements, err = loadAnnouncementsFromFile(englishPath)
		if err == nil && len(announcements) > 0 {
			return announcements, nil
		}
	}

	// Final fallback: default announcements
	return getDefaultAnnouncements(), nil
}

// loadAnnouncementsFromFile reads announcements from a JSON file
func loadAnnouncementsFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var announcementData AnnouncementData
	if err := json.Unmarshal(data, &announcementData); err != nil {
		return nil, err
	}

	return announcementData.CompanyAnnouncements, nil
}

// getDefaultAnnouncements returns default English announcements as fallback
func getDefaultAnnouncements() []string {
	return []string{
		"🗿 JikiME-ADK: SPEC-First DDD with Skills and Context7 integration",
		"⚡ /jikime:alfred: One-stop Plan→Run→Sync automation with intelligent routing",
		"🤖 Expert Agents: backend, frontend, security, devops, performance, debug, testing, refactoring",
		"🤖 Manager Agents: git, spec, ddd, docs, quality, project, strategy, claude-code",
		"🤖 Builder Agents: agent, command, skill, plugin - create custom extensions",
		"📋 Workflow: /jikime:1-plan (SPEC) → /jikime:2-run (DDD) → /jikime:3-sync (Docs)",
		"✅ Quality: TRUST 5 + ≥85% coverage + Ralph Engine (LSP + AST-grep)",
		"🔄 Git Strategy: 3-Mode (Manual/Personal/Team) with Smart Merge config updates",
		"📚 Tip: jikime update --templates-only syncs latest skills and agents to your project",
	}
}
