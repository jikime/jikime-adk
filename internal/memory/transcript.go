package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const maxScannerBuffer = 1 << 20 // 1MB per line

// ParseTranscript reads a Claude Code JSONL transcript file.
// It categorizes records by type: summary, user, and skips others.
func ParseTranscript(path string) (*Transcript, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, maxScannerBuffer), maxScannerBuffer)

	t := &Transcript{}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var record TranscriptRecord
		if err := json.Unmarshal(line, &record); err != nil {
			continue // skip malformed lines
		}

		switch record.Type {
		case "summary":
			t.Summaries = append(t.Summaries, record)
		case "user":
			t.UserMsgs = append(t.UserMsgs, record)
			// Extract session ID from first user record
			if t.SessionID == "" && record.SessionID != "" {
				t.SessionID = record.SessionID
			}
		}
		// Skip "progress", "file-history-snapshot", and other types in Phase 1
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transcript: %w", err)
	}

	return t, nil
}

// GetMessageText extracts the text content from a TranscriptRecord's Message field.
// The message field can be a string or a structured object with role/content.
func GetMessageText(r TranscriptRecord) string {
	if r.Message == nil {
		return ""
	}

	// Case 1: message is a plain string
	if s, ok := r.Message.(string); ok {
		return stripInternalMarkup(s)
	}

	// Case 2: message is a map with "content" field
	if m, ok := r.Message.(map[string]interface{}); ok {
		if content, exists := m["content"]; exists {
			switch v := content.(type) {
			case string:
				return stripInternalMarkup(v)
			case []interface{}:
				// Content can be an array of content blocks
				var parts []string
				for _, block := range v {
					if bm, ok := block.(map[string]interface{}); ok {
						if text, exists := bm["text"]; exists {
							if s, ok := text.(string); ok {
								parts = append(parts, s)
							}
						}
					}
				}
				if len(parts) > 0 {
					return stripInternalMarkup(joinStrings(parts, "\n"))
				}
			}
		}
	}

	return ""
}

// Claude Code internal XML tag patterns to strip from transcript text.
var (
	// Matches full XML elements: <tag ...>content</tag>
	xmlElementRegex = regexp.MustCompile(`<(?:local-command-caveat|local-command-stdout|command-name|command-message|command-args|system-reminder|user-prompt-submit-hook)[^>]*>[\s\S]*?</(?:local-command-caveat|local-command-stdout|command-name|command-message|command-args|system-reminder|user-prompt-submit-hook)>`)
	// Matches self-closing or unclosed XML tags
	xmlTagRegex = regexp.MustCompile(`</?(?:local-command-caveat|local-command-stdout|command-name|command-message|command-args|system-reminder|user-prompt-submit-hook)[^>]*>`)
)

// stripInternalMarkup removes Claude Code internal XML tags from transcript text.
func stripInternalMarkup(text string) string {
	// Remove full XML elements first (tag + content + closing tag)
	text = xmlElementRegex.ReplaceAllString(text, "")
	// Remove any remaining standalone tags
	text = xmlTagRegex.ReplaceAllString(text, "")
	// Clean up leftover whitespace
	text = strings.TrimSpace(text)
	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text
}

// ExtractLastMessage reads a transcript JSONL file in reverse and returns
// the last message text for the given role ("user" or "assistant").
// Mirrors jikime-mem's extractLastMessage() behavior.
func ExtractLastMessage(path string, role string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read transcript: %w", err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", fmt.Errorf("empty transcript: %s", path)
	}

	lines := strings.Split(content, "\n")

	// Reverse scan to find the last message of the given role
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		var record TranscriptRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		if record.Type != role {
			continue
		}

		text := GetMessageText(record)
		if text != "" {
			return text, nil
		}
	}

	return "", fmt.Errorf("no %s message found in transcript", role)
}

// ExtractLastAssistantMessage extracts the last assistant response from a transcript JSONL file.
// Convenience wrapper matching jikime-mem's extractLastAssistantMessage().
func ExtractLastAssistantMessage(path string) (string, error) {
	return ExtractLastMessage(path, "assistant")
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
