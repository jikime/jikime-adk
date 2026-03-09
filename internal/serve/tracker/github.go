package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"jikime-adk/internal/serve"
)

const (
	githubDefaultEndpoint = "https://api.github.com"
	githubPageSize        = 50
	githubNetworkTimeout  = 30 * time.Second
)

// GitHub is an issue tracker client backed by the GitHub REST API.
// active_states map to GitHub labels; issues without those labels are skipped.
// terminal_states correspond to "closed" GitHub state.
type GitHub struct {
	endpoint       string
	apiKey         string
	owner          string
	repo           string
	activeStates   []string
	terminalStates []string
	httpClient     *http.Client
}

// NewGitHub creates a GitHub tracker client.
// projectSlug must be "owner/repo" format.
func NewGitHub(endpoint, apiKey, projectSlug string, activeStates, terminalStates []string) (*GitHub, error) {
	parts := strings.SplitN(projectSlug, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("github tracker: project_slug must be 'owner/repo', got %q", projectSlug)
	}
	if endpoint == "" {
		endpoint = githubDefaultEndpoint
	}
	return &GitHub{
		endpoint:       strings.TrimRight(endpoint, "/"),
		apiKey:         apiKey,
		owner:          parts[0],
		repo:           parts[1],
		activeStates:   activeStates,
		terminalStates: terminalStates,
		httpClient:     &http.Client{Timeout: githubNetworkTimeout},
	}, nil
}

// FetchCandidateIssues returns open issues that have at least one active-state label.
// Issues without any active-state label are excluded.
func (g *GitHub) FetchCandidateIssues() ([]serve.Issue, error) {
	var all []serve.Issue
	page := 1

	for {
		issues, hasMore, err := g.fetchPage("open", page)
		if err != nil {
			return nil, err
		}
		all = append(all, issues...)
		if !hasMore {
			break
		}
		page++
	}

	// Filter: only issues with at least one active-state label
	filtered := make([]serve.Issue, 0, len(all))
	for _, issue := range all {
		if g.hasActiveLabel(issue.Labels) {
			filtered = append(filtered, issue)
		}
	}
	return filtered, nil
}

// FetchIssueStatesByIDs returns current state for specific GitHub issue numbers.
// IDs are GitHub issue numbers (e.g. "123").
func (g *GitHub) FetchIssueStatesByIDs(ids []string) ([]serve.Issue, error) {
	var result []serve.Issue
	for _, id := range ids {
		issue, err := g.fetchIssueByID(id)
		if err != nil {
			return nil, err
		}
		if issue != nil {
			result = append(result, *issue)
		}
	}
	return result, nil
}

// FetchIssuesByStates returns issues in the given states.
// For GitHub, "closed" terminal states → fetch closed issues.
func (g *GitHub) FetchIssuesByStates(states []string) ([]serve.Issue, error) {
	if len(states) == 0 {
		return nil, nil
	}

	// Determine GitHub state filter
	githubState := "closed"
	for _, s := range states {
		sl := strings.ToLower(strings.TrimSpace(s))
		if sl == "open" || sl == "todo" || sl == "in progress" {
			githubState = "open"
			break
		}
	}

	var all []serve.Issue
	page := 1
	for {
		issues, hasMore, err := g.fetchPage(githubState, page)
		if err != nil {
			return nil, err
		}
		all = append(all, issues...)
		if !hasMore {
			break
		}
		page++
	}
	return all, nil
}

// --- Internal helpers ---

type githubIssue struct {
	Number    int            `json:"number"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	State     string         `json:"state"` // "open" or "closed"
	HTMLURL   string         `json:"html_url"`
	Labels    []githubLabel  `json:"labels"`
	Milestone *githubMilestone `json:"milestone"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type githubLabel struct {
	Name string `json:"name"`
}

type githubMilestone struct {
	Title string `json:"title"`
}

func (g *GitHub) fetchPage(state string, page int) ([]serve.Issue, bool, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/issues", g.endpoint, g.owner, g.repo)
	params := url.Values{}
	params.Set("state", state)
	params.Set("per_page", strconv.Itoa(githubPageSize))
	params.Set("page", strconv.Itoa(page))
	params.Set("sort", "created")
	params.Set("direction", "asc")
	fullURL := apiURL + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("github_api_request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("github_api_request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("github_api_status: %d", resp.StatusCode)
	}

	var raw []githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, false, fmt.Errorf("github_unknown_payload: %w", err)
	}

	issues := make([]serve.Issue, 0, len(raw))
	for _, r := range raw {
		// Skip pull requests (GitHub API includes PRs in issues endpoint)
		if strings.Contains(r.HTMLURL, "/pull/") {
			continue
		}
		issues = append(issues, g.normalize(r))
	}

	hasMore := len(raw) == githubPageSize
	return issues, hasMore, nil
}

func (g *GitHub) fetchIssueByID(id string) (*serve.Issue, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/issues/%s", g.endpoint, g.owner, g.repo, id)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("github_api_request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github_api_request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github_api_status: %d", resp.StatusCode)
	}

	var raw githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("github_unknown_payload: %w", err)
	}

	issue := g.normalize(raw)
	return &issue, nil
}

// normalize converts a GitHub issue to the canonical Issue type.
// State is synthesized from GitHub state + labels.
func (g *GitHub) normalize(r githubIssue) serve.Issue {
	labels := make([]string, 0, len(r.Labels))
	for _, l := range r.Labels {
		labels = append(labels, strings.ToLower(l.Name))
	}

	// Derive logical state from GitHub state + labels
	state := g.deriveState(r.State, labels)

	id := strconv.Itoa(r.Number)

	var createdAt, updatedAt *time.Time
	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		createdAt = &t
	}
	if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
		updatedAt = &t
	}

	return serve.Issue{
		ID:          id,
		Identifier:  fmt.Sprintf("%s/%s#%s", g.owner, g.repo, id),
		Title:       r.Title,
		Description: r.Body,
		State:       state,
		URL:         r.HTMLURL,
		Labels:      labels,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// deriveState maps GitHub issue state + labels to a logical state name.
// Priority: closed → "Done"; label match → label name; open → "In Progress".
func (g *GitHub) deriveState(githubState string, labels []string) string {
	if githubState == "closed" {
		return "Done"
	}
	// Match active state from labels (first match wins)
	for _, activeState := range g.activeStates {
		slug := strings.ToLower(strings.ReplaceAll(activeState, " ", "-"))
		for _, label := range labels {
			if label == slug || label == strings.ToLower(activeState) {
				return activeState
			}
		}
	}
	return "In Progress"
}

// hasActiveLabel returns true if any label matches an active state.
func (g *GitHub) hasActiveLabel(labels []string) bool {
	if len(g.activeStates) == 0 {
		return true // no filter configured → include all
	}
	labelSet := make(map[string]bool, len(labels))
	for _, l := range labels {
		labelSet[l] = true
	}
	for _, activeState := range g.activeStates {
		slug := strings.ToLower(strings.ReplaceAll(activeState, " ", "-"))
		if labelSet[slug] || labelSet[strings.ToLower(activeState)] {
			return true
		}
	}
	return false
}
