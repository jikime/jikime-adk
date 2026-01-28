package memory

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Technical patterns: filenames, function-like tokens, package paths
	techPattern = regexp.MustCompile(`(?i)[\w]+\.(?:go|ts|js|py|rs|java|tsx|jsx|css|html|sql|yaml|yml|json|toml|md)|[\w]+(?:Handler|Service|Controller|Provider|Store|Manager|Config|Error|Test)|[\w]+/[\w]+`)
	// Stop words (English)
	engStopWords = map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "shall": true,
		"can": true, "to": true, "of": true, "in": true, "for": true, "on": true,
		"with": true, "at": true, "by": true, "from": true, "as": true, "into": true,
		"about": true, "that": true, "this": true, "it": true, "its": true,
		"and": true, "or": true, "but": true, "not": true, "no": true,
		"if": true, "then": true, "else": true, "when": true, "up": true,
		"so": true, "than": true, "too": true, "very": true, "just": true,
		"i": true, "me": true, "my": true, "we": true, "our": true, "you": true, "your": true,
		"he": true, "she": true, "they": true, "them": true, "their": true,
		"what": true, "which": true, "who": true, "how": true, "where": true, "why": true,
		"all": true, "each": true, "some": true, "any": true, "most": true,
		"please": true, "help": true, "want": true, "need": true, "like": true,
	}
	// Korean stop words (particles/endings)
	korStopWords = map[string]bool{
		"을": true, "를": true, "이": true, "가": true, "은": true, "는": true,
		"에": true, "의": true, "로": true, "으로": true, "와": true, "과": true,
		"도": true, "만": true, "에서": true, "까지": true, "부터": true,
		"해줘": true, "해주세요": true, "좀": true, "것": true, "수": true,
	}
)

// ExtractKeywords extracts searchable keywords from user prompt.
// Returns up to 5 keywords, prioritizing technical terms.
func ExtractKeywords(prompt string) []string {
	if strings.TrimSpace(prompt) == "" {
		return nil
	}

	// 1. Extract technical patterns first (filenames, function names, paths)
	techMatches := techPattern.FindAllString(prompt, -1)

	// 2. Tokenize remaining text
	tokens := tokenize(prompt)

	// 3. Filter stop words and short tokens
	var filtered []string
	seen := make(map[string]bool)

	// Add tech terms first (higher priority)
	for _, t := range techMatches {
		lower := strings.ToLower(t)
		if !seen[lower] {
			seen[lower] = true
			filtered = append(filtered, t)
		}
	}

	// Add remaining meaningful tokens
	for _, t := range tokens {
		lower := strings.ToLower(t)
		if seen[lower] {
			continue
		}
		if isStopWord(lower) {
			continue
		}
		if len([]rune(t)) < 2 {
			continue
		}
		seen[lower] = true
		filtered = append(filtered, t)
	}

	// Limit to 5 keywords
	if len(filtered) > 5 {
		filtered = filtered[:5]
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func tokenize(s string) []string {
	var tokens []string
	var current []rune

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				tokens = append(tokens, string(current))
				current = current[:0]
			}
		}
	}
	if len(current) > 0 {
		tokens = append(tokens, string(current))
	}
	return tokens
}

func isStopWord(w string) bool {
	return engStopWords[w] || korStopWords[w]
}
