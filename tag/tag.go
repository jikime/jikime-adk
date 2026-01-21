// Package tag provides TAG System v2.0 for jikime-adk.
// Implements traceability between SPEC documents and code through @SPEC TAG annotations.
package tag

import (
	"regexp"
)

// Version of the TAG system.
const Version = "2.0.0"

// SPEC-ID pattern: SPEC-{DOMAIN}-{NUMBER}
// DOMAIN: Uppercase letters and digits, at least 1 char
// NUMBER: 3 digits
var SpecIDPattern = regexp.MustCompile(`^SPEC-[A-Z0-9]+-\d{3}$`)

// ValidVerbs defines the allowed TAG relationship verbs.
var ValidVerbs = map[string]bool{
	"impl":    true, // Implementation
	"verify":  true, // Verification/test
	"depends": true, // Dependency
	"related": true, // Related reference
}

// DefaultVerb is the default verb when not specified.
const DefaultVerb = "impl"

// TAG represents a single TAG annotation.
// TAG format: # @SPEC SPEC-ID [verb]
type TAG struct {
	SpecID   string `json:"spec_id"`   // SPEC identifier (e.g., "SPEC-AUTH-001")
	Verb     string `json:"verb"`      // TAG relationship verb (impl, verify, depends, related)
	FilePath string `json:"file_path"` // Path to the file containing the TAG
	Line     int    `json:"line"`      // Line number where TAG appears
}

// NewTAG creates a new TAG with validation and normalization.
func NewTAG(specID, verb, filePath string, line int) *TAG {
	// Normalize verb to lowercase
	if verb == "" {
		verb = DefaultVerb
	}

	return &TAG{
		SpecID:   specID,
		Verb:     verb,
		FilePath: filePath,
		Line:     line,
	}
}

// IsValid checks if the TAG has valid format.
func (t *TAG) IsValid() bool {
	return ValidateSpecIDFormat(t.SpecID) && ValidateVerb(t.Verb)
}

// Equals checks if two TAGs are equal (same spec_id, verb, file_path, line).
func (t *TAG) Equals(other *TAG) bool {
	if other == nil {
		return false
	}
	return t.SpecID == other.SpecID &&
		t.Verb == other.Verb &&
		t.FilePath == other.FilePath &&
		t.Line == other.Line
}

// ToMap converts TAG to a map for JSON serialization.
func (t *TAG) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"spec_id":   t.SpecID,
		"verb":      t.Verb,
		"file_path": t.FilePath,
		"line":      t.Line,
	}
}

// TAGFromMap creates a TAG from a map.
func TAGFromMap(m map[string]interface{}) *TAG {
	specID, _ := m["spec_id"].(string)
	verb, _ := m["verb"].(string)
	filePath, _ := m["file_path"].(string)
	line := 0
	if l, ok := m["line"].(float64); ok {
		line = int(l)
	} else if l, ok := m["line"].(int); ok {
		line = l
	}

	return NewTAG(specID, verb, filePath, line)
}
