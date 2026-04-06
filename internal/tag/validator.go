// Package tag provides TAG validation for TAG System v2.0.
package tag

import (
	"fmt"
	"strings"
)

// ValidationError represents a TAG validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult contains the result of TAG validation.
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidateSpecIDFormat validates SPEC-ID format: SPEC-{DOMAIN}-{NUMBER}.
// Examples:
//   - ValidateSpecIDFormat("SPEC-AUTH-001") => true
//   - ValidateSpecIDFormat("spec-auth-001") => false (lowercase)
//   - ValidateSpecIDFormat("SPEC-AUTH-1") => false (missing digits)
func ValidateSpecIDFormat(specID string) bool {
	return SpecIDPattern.MatchString(specID)
}

// ValidateVerb validates TAG verb.
// Valid verbs: impl, verify, depends, related
func ValidateVerb(verb string) bool {
	return ValidVerbs[verb]
}

// GetDefaultVerb returns the default TAG verb.
func GetDefaultVerb() string {
	return DefaultVerb
}

// GetValidVerbs returns all valid verbs.
func GetValidVerbs() []string {
	verbs := make([]string, 0, len(ValidVerbs))
	for v := range ValidVerbs {
		verbs = append(verbs, v)
	}
	return verbs
}

// ValidateTAG validates a complete TAG.
// Returns ValidationResult with validity status and any errors.
func ValidateTAG(tag *TAG) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []ValidationError{}}

	// Validate SPEC-ID format
	if !ValidateSpecIDFormat(tag.SpecID) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field: "spec_id",
			Message: fmt.Sprintf(
				"Invalid SPEC-ID format: '%s'. Expected format: SPEC-{DOMAIN}-{NUMBER} (e.g., SPEC-AUTH-001)",
				tag.SpecID,
			),
		})
	}

	// Validate verb
	if !ValidateVerb(tag.Verb) {
		result.Valid = false
		validVerbs := strings.Join(GetValidVerbs(), ", ")
		result.Errors = append(result.Errors, ValidationError{
			Field:   "verb",
			Message: fmt.Sprintf("Invalid verb: '%s'. Valid verbs: %s", tag.Verb, validVerbs),
		})
	}

	return result
}

// ParseTAGString parses TAG from comment string.
// Format: # @SPEC SPEC-ID [verb]
// Returns nil if not a valid TAG comment.
func ParseTAGString(comment, filePath string, line int) *TAG {
	// Remove leading/trailing whitespace
	comment = strings.TrimSpace(comment)

	// Check if comment starts with #
	if !strings.HasPrefix(comment, "#") {
		return nil
	}

	// Remove # and leading whitespace
	comment = strings.TrimSpace(comment[1:])

	// Check for @SPEC prefix
	if !strings.HasPrefix(comment, "@SPEC") {
		return nil
	}

	// Remove @SPEC prefix
	comment = strings.TrimSpace(comment[5:]) // len("@SPEC") == 5

	// Split into parts
	parts := strings.Fields(comment)
	if len(parts) == 0 {
		return nil
	}

	// First part is SPEC-ID
	specID := parts[0]

	// Validate SPEC-ID format
	if !ValidateSpecIDFormat(specID) {
		return nil
	}

	// Second part (if present) is verb
	verb := DefaultVerb
	if len(parts) > 1 && ValidateVerb(parts[1]) {
		verb = parts[1]
	}

	return NewTAG(specID, verb, filePath, line)
}
