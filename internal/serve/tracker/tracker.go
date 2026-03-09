// Package tracker defines the issue tracker client interface and adapters.
// Implements Symphony SPEC Section 11.
package tracker

import (
	"fmt"

	"jikime-adk/internal/serve"
)

// Client is the issue tracker abstraction.
// Implementations: GitHub, Linear (optional).
type Client interface {
	// FetchCandidateIssues returns open issues in active states for dispatch.
	FetchCandidateIssues() ([]serve.Issue, error)

	// FetchIssueStatesByIDs returns current state for specific issue IDs (reconciliation).
	FetchIssueStatesByIDs(ids []string) ([]serve.Issue, error)

	// FetchIssuesByStates returns issues in the given states (startup cleanup).
	FetchIssuesByStates(states []string) ([]serve.Issue, error)
}

// NewClient creates the appropriate tracker client based on kind.
func NewClient(kind, endpoint, apiKey, projectSlug string, activeStates, terminalStates []string) (Client, error) {
	switch kind {
	case "github":
		return NewGitHub(endpoint, apiKey, projectSlug, activeStates, terminalStates)
	default:
		return nil, fmt.Errorf("unsupported_tracker_kind: %q (supported: github)", kind)
	}
}
