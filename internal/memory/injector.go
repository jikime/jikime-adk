package memory

import (
	"fmt"
	"strings"
)

const maxContextBytes = 16 * 1024 // ~4K tokens

// BuildStartupContext builds context for source="startup".
// Includes: last session summary, project knowledge, recent memories.
func BuildStartupContext(lastSession *SessionRecord, knowledge []ProjectKnowledge, memories []Memory) string {
	var b strings.Builder
	b.WriteString("## Session Memory\n\n")

	// Last session
	if lastSession != nil && lastSession.Summary != "" {
		b.WriteString("### Last Session\n")
		b.WriteString(fmt.Sprintf("- Session: %s\n", lastSession.SessionID))
		if !lastSession.EndedAt.IsZero() {
			b.WriteString(fmt.Sprintf("- Date: %s\n", lastSession.EndedAt.Format("2006-01-02 15:04")))
		}
		b.WriteString(fmt.Sprintf("- Summary: %s\n", truncate(lastSession.Summary, 500)))

		if len(lastSession.Topics) > 0 {
			b.WriteString(fmt.Sprintf("- Topics: %s\n", strings.Join(lastSession.Topics, ", ")))
		}
		if len(lastSession.FilesModified) > 0 {
			files := lastSession.FilesModified
			if len(files) > 10 {
				files = files[:10]
			}
			b.WriteString(fmt.Sprintf("- Files Modified: %s\n", strings.Join(files, ", ")))
		}
		b.WriteString("\n")
	}

	// Project knowledge
	writeKnowledge(&b, knowledge)

	// Recent memories
	writeRecentMemories(&b, memories)

	return enforceLimit(b.String())
}

// BuildCompactContext builds context for source="compact".
// Reloads current session memories + project knowledge after compaction.
func BuildCompactContext(memories []Memory, knowledge []ProjectKnowledge) string {
	var b strings.Builder
	b.WriteString("## Session Memory (Post-Compaction)\n\n")
	b.WriteString("Context was compacted. Key information from this session:\n\n")

	// Current session memories
	if len(memories) > 0 {
		b.WriteString("### Current Session Context\n")
		for i, m := range memories {
			if i >= 15 {
				b.WriteString(fmt.Sprintf("... and %d more items\n", len(memories)-15))
				break
			}
			b.WriteString(fmt.Sprintf("- [%s] %s\n", m.Type, truncate(m.Content, 300)))
		}
		b.WriteString("\n")
	}

	// Project knowledge
	writeKnowledge(&b, knowledge)

	return enforceLimit(b.String())
}

// BuildResumeContext builds context for source="resume".
func BuildResumeContext(session *SessionRecord) string {
	if session == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Session Memory (Resumed)\n\n")
	b.WriteString(fmt.Sprintf("Resuming session %s.\n\n", session.SessionID))

	if session.Summary != "" {
		b.WriteString("### Session Summary\n")
		b.WriteString(truncate(session.Summary, 1000))
		b.WriteString("\n\n")
	}

	if len(session.Topics) > 0 {
		b.WriteString("### Topics\n")
		for _, t := range session.Topics {
			b.WriteString(fmt.Sprintf("- %s\n", t))
		}
		b.WriteString("\n")
	}

	if len(session.FilesModified) > 0 {
		b.WriteString("### Files Modified\n")
		files := session.FilesModified
		if len(files) > 15 {
			files = files[:15]
		}
		for _, f := range files {
			b.WriteString(fmt.Sprintf("- %s\n", f))
		}
		b.WriteString("\n")
	}

	return enforceLimit(b.String())
}

// BuildClearContext builds context for source="clear".
// Only project knowledge (no session-specific memories).
func BuildClearContext(knowledge []ProjectKnowledge) string {
	if len(knowledge) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Project Knowledge\n\n")
	writeKnowledge(&b, knowledge)

	return enforceLimit(b.String())
}

// Summarize creates a SessionSummary from a parsed transcript.
func Summarize(t *Transcript) SessionSummary {
	summary := SessionSummary{}

	// Collect summaries
	var summaryParts []string
	for _, s := range t.Summaries {
		if s.Summary != "" {
			summaryParts = append(summaryParts, s.Summary)
		}
	}

	if len(summaryParts) > 0 {
		summary.Text = strings.Join(summaryParts, " | ")
	} else {
		// Fall back to user messages
		var msgParts []string
		for i, msg := range t.UserMsgs {
			if i >= 5 {
				break
			}
			text := GetMessageText(msg)
			if text != "" {
				msgParts = append(msgParts, truncate(text, 200))
			}
		}
		if len(msgParts) > 0 {
			summary.Text = strings.Join(msgParts, " | ")
		}
	}

	// Extract topics (from summary text)
	summary.Topics = extractTopics(summary.Text)

	// Extract file paths from all messages
	fileSet := make(map[string]bool)
	for _, msg := range t.UserMsgs {
		text := GetMessageText(msg)
		for _, f := range extractFilePaths(text) {
			fileSet[f] = true
		}
	}
	for f := range fileSet {
		summary.Files = append(summary.Files, f)
	}

	return summary
}

func writeKnowledge(b *strings.Builder, knowledge []ProjectKnowledge) {
	if len(knowledge) == 0 {
		return
	}

	b.WriteString("### Project Knowledge\n")
	for i, k := range knowledge {
		if i >= 10 {
			b.WriteString(fmt.Sprintf("... and %d more items\n", len(knowledge)-10))
			break
		}
		label := k.KnowledgeType
		if label == "" {
			label = "info"
		}
		b.WriteString(fmt.Sprintf("- [%s] %s\n", label, truncate(k.Content, 300)))
	}
	b.WriteString("\n")
}

func writeRecentMemories(b *strings.Builder, memories []Memory) {
	if len(memories) == 0 {
		return
	}

	// Group by type for cleaner output
	decisions := filterByType(memories, TypeDecision)
	learnings := filterByType(memories, TypeLearning)
	fixes := filterByType(memories, TypeErrorFix)
	summaries := filterByType(memories, TypeSessionSummary)

	if len(decisions) > 0 {
		b.WriteString("### Key Decisions\n")
		for i, m := range decisions {
			if i >= 5 {
				break
			}
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncate(m.Content, 200)))
		}
		b.WriteString("\n")
	}

	if len(learnings) > 0 {
		b.WriteString("### Learnings\n")
		for i, m := range learnings {
			if i >= 5 {
				break
			}
			b.WriteString(fmt.Sprintf("- %s\n", truncate(m.Content, 200)))
		}
		b.WriteString("\n")
	}

	if len(fixes) > 0 {
		b.WriteString("### Error Fixes\n")
		for i, m := range fixes {
			if i >= 3 {
				break
			}
			b.WriteString(fmt.Sprintf("- %s\n", truncate(m.Content, 200)))
		}
		b.WriteString("\n")
	}

	if len(summaries) > 0 {
		b.WriteString("### Recent Activity\n")
		for i, m := range summaries {
			if i >= 5 {
				break
			}
			b.WriteString(fmt.Sprintf("- %s\n", truncate(m.Content, 200)))
		}
		b.WriteString("\n")
	}
}

func filterByType(memories []Memory, memType string) []Memory {
	var filtered []Memory
	for _, m := range memories {
		if m.Type == memType {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func enforceLimit(s string) string {
	if len(s) <= maxContextBytes {
		return s
	}
	// Truncate at the limit, finding the last newline before the cutoff
	cut := s[:maxContextBytes]
	lastNewline := strings.LastIndex(cut, "\n")
	if lastNewline > maxContextBytes/2 {
		return cut[:lastNewline] + "\n\n... (context truncated)\n"
	}
	return cut + "\n\n... (context truncated)\n"
}

// extractTopics extracts topic keywords from summary text.
func extractTopics(text string) []string {
	if text == "" {
		return nil
	}

	// Simple approach: split by common delimiters and take meaningful phrases
	var topics []string
	seen := make(map[string]bool)

	// Split by pipes, commas, periods
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '|' || r == '.' || r == ','
	})

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) < 5 || len(p) > 100 {
			continue
		}
		lower := strings.ToLower(p)
		if !seen[lower] {
			seen[lower] = true
			topics = append(topics, p)
		}
		if len(topics) >= 10 {
			break
		}
	}

	return topics
}
