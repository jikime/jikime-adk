package memory

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Keyword patterns for extracting important content from user messages.
var (
	decisionPatterns = []string{
		"decided", "chose", "decision", "agreed", "settled on",
		"will use", "going with", "선택", "결정", "채택",
	}
	errorFixPatterns = []string{
		"fixed", "resolved", "the issue was", "root cause",
		"solution was", "the problem was", "bug was",
		"해결", "수정", "원인",
	}
	architecturePatterns = []string{
		"architecture", "pattern", "approach", "design", "structure",
		"아키텍처", "패턴", "설계", "구조",
	}
	preferencePatterns = []string{
		"prefer", "always use", "never use", "convention",
		"standard is", "rule is", "our approach",
		"선호", "규칙", "컨벤션",
	}
	learningPatterns = []string{
		"learned", "discovered", "realized", "turns out",
		"important to note", "key insight",
		"배움", "발견", "깨달",
	}

	filePathRegex = regexp.MustCompile(`(?:^|\s|[("'])([a-zA-Z0-9_\-./]+\.[a-zA-Z]{1,10})(?:\s|[)"']|$|:)`)
)

// Extract analyzes a transcript and returns memories worth saving.
func Extract(t *Transcript, opts ExtractOptions) []Memory {
	var memories []Memory
	seen := make(map[string]bool) // content hash dedup

	// Priority 1: summary records → direct conversion (already Claude-summarized)
	memories = append(memories, extractFromSummaries(t.Summaries, opts, seen)...)

	// Priority 2: user messages → keyword pattern matching
	memories = append(memories, extractFromUserMessages(t.UserMsgs, opts, seen)...)

	return memories
}

func extractFromSummaries(summaries []TranscriptRecord, opts ExtractOptions, seen map[string]bool) []Memory {
	var memories []Memory
	for _, s := range summaries {
		if s.Summary == "" {
			continue
		}

		hash := ContentHash(s.Summary)
		if seen[hash] {
			continue
		}
		seen[hash] = true

		metadata, _ := json.Marshal(map[string]string{
			"source":  "transcript_summary",
			"trigger": opts.Trigger,
		})

		memories = append(memories, Memory{
			ID:          generateID(),
			SessionID:   opts.SessionID,
			ProjectDir:  opts.ProjectDir,
			Type:        TypeSessionSummary,
			Content:     s.Summary,
			ContentHash: hash,
			Metadata:    string(metadata),
			CreatedAt:   time.Now(),
		})
	}
	return memories
}

func extractFromUserMessages(msgs []TranscriptRecord, opts ExtractOptions, seen map[string]bool) []Memory {
	var memories []Memory

	for _, msg := range msgs {
		text := GetMessageText(msg)
		if text == "" || len(text) < 20 {
			continue
		}

		important, memType := classifyContent(text)
		if !important {
			continue
		}

		// Truncate long content to first 2000 chars
		content := text
		if len(content) > 2000 {
			content = content[:2000] + "..."
		}

		hash := ContentHash(content)
		if seen[hash] {
			continue
		}
		seen[hash] = true

		metadata, _ := json.Marshal(map[string]string{
			"source":  "transcript_user",
			"trigger": opts.Trigger,
			"files":   strings.Join(extractFilePaths(text), ","),
		})

		memories = append(memories, Memory{
			ID:          generateID(),
			SessionID:   opts.SessionID,
			ProjectDir:  opts.ProjectDir,
			Type:        memType,
			Content:     content,
			ContentHash: hash,
			Metadata:    string(metadata),
			CreatedAt:   time.Now(),
		})
	}

	return memories
}

// classifyContent checks if text matches extraction heuristics and returns the memory type.
func classifyContent(text string) (bool, string) {
	lower := strings.ToLower(text)

	if matchesAny(lower, decisionPatterns) {
		return true, TypeDecision
	}
	if matchesAny(lower, errorFixPatterns) {
		return true, TypeErrorFix
	}
	if matchesAny(lower, architecturePatterns) {
		return true, TypeDecision
	}
	if matchesAny(lower, preferencePatterns) {
		return true, TypeLearning
	}
	if matchesAny(lower, learningPatterns) {
		return true, TypeLearning
	}

	return false, ""
}

func matchesAny(text string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// extractFilePaths finds file paths mentioned in text.
func extractFilePaths(text string) []string {
	matches := filePathRegex.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var paths []string

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		p := m[1]
		// Filter out common false positives
		if isLikelyFilePath(p) && !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	return paths
}

func isLikelyFilePath(p string) bool {
	// Must contain a dot and a directory separator or start with valid prefix
	if !strings.Contains(p, ".") {
		return false
	}
	// Filter out domain names, version numbers, etc.
	if strings.Contains(p, "www.") || strings.Contains(p, "http") {
		return false
	}
	// Common file extensions
	exts := []string{
		".go", ".ts", ".js", ".tsx", ".jsx", ".py", ".rs", ".java",
		".yaml", ".yml", ".json", ".toml", ".md", ".sql", ".sh",
		".css", ".scss", ".html", ".vue", ".svelte",
	}
	for _, ext := range exts {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}
