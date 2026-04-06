package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"jikime-adk/internal/team"
)

// boardCache holds cached store/registry instances to avoid re-creating them
// on every collectBoardData call (< 5s TTL).
var (
	boardCacheMu sync.Mutex
	cachedBoards = make(map[string]*boardCacheEntry)
)

type boardCacheEntry struct {
	reg     *team.Registry
	store   *team.Store
	created time.Time
}

// boardTaskItem is the wire format for a task card in the web UI.
type boardTaskItem struct {
	ID        string   `json:"id"`
	Subject   string   `json:"subject"`
	Owner     string   `json:"owner,omitempty"`
	BlockedBy []string `json:"blockedBy,omitempty"`
}

// boardMember is the wire format for an agent member card.
type boardMember struct {
	Name       string `json:"name"`
	User       string `json:"user,omitempty"`
	AgentType  string `json:"agentType"`
	InboxCount int    `json:"inboxCount"`
}

// boardMessage is the wire format for a message in the event log panel.
type boardMessage struct {
	From      string `json:"from"`
	To        string `json:"to,omitempty"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

// collectBoardData builds the full board payload for a team.
func collectBoardData(name string) (map[string]interface{}, error) {
	td := teamDir(name)
	if _, err := os.Stat(td); os.IsNotExist(err) {
		return nil, fmt.Errorf("team %q not found", name)
	}

	// Config
	cfg := struct {
		LeaderID    string `json:"leader_id"`
		Description string `json:"description"`
	}{}
	data, err := os.ReadFile(filepath.Join(td, "config.json"))
	if err == nil {
		if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse team config for %s: %v\n", name, jsonErr)
		}
	}

	// Reuse cached store/registry instances if < 5s old
	boardCacheMu.Lock()
	cached, cacheHit := cachedBoards[name]
	if cacheHit && time.Since(cached.created) >= 5*time.Second {
		cacheHit = false
	}
	boardCacheMu.Unlock()

	var reg *team.Registry
	var store *team.Store
	if cacheHit {
		reg = cached.reg
		store = cached.store
	} else {
		var regErr, storeErr error
		reg, regErr = team.NewRegistry(filepath.Join(td, "registry"))
		if regErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load registry for %s: %v\n", name, regErr)
		}
		store, storeErr = team.NewStore(filepath.Join(td, "tasks"))
		if storeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load task store for %s: %v\n", name, storeErr)
		}
		boardCacheMu.Lock()
		cachedBoards[name] = &boardCacheEntry{reg: reg, store: store, created: time.Now()}
		boardCacheMu.Unlock()
	}

	// Agents
	var agents []*team.AgentInfo
	if reg != nil {
		agents, _ = reg.List()
	}

	leaderName := cfg.LeaderID
	members := make([]boardMember, 0, len(agents))
	for _, a := range agents {
		if leaderName == "" && a.Role == "leader" {
			leaderName = a.ID
		}
		ib, _ := team.NewInbox(team.InboxDir(td, a.ID))
		inboxCount := 0
		if ib != nil {
			inboxCount, _ = ib.Count()
		}
		members = append(members, boardMember{
			Name:       a.ID,
			AgentType:  string(a.Role),
			InboxCount: inboxCount,
		})
	}

	// Tasks
	var allTasks []*team.Task
	if store != nil {
		allTasks, _ = store.List("", "")
	}

	taskGroups := map[string][]boardTaskItem{
		"pending":     {},
		"in_progress": {},
		"done":        {},
		"failed":      {},
		"blocked":     {},
	}
	summary := map[string]int{
		"pending": 0, "in_progress": 0, "done": 0, "failed": 0, "blocked": 0,
	}
	for _, t := range allTasks {
		status := string(t.Status)
		item := boardTaskItem{
			ID:      t.ID,
			Subject: t.Title,
			Owner:   t.AgentID,
		}
		if _, ok := taskGroups[status]; ok {
			taskGroups[status] = append(taskGroups[status], item)
			summary[status]++
		}
	}

	// Messages (event log, last 50)
	ti := team.NewTeamInbox(td)
	logMsgs, _ := ti.EventLog(50, "")
	messages := make([]boardMessage, 0, len(logMsgs))
	for _, m := range logMsgs {
		messages = append(messages, boardMessage{
			From:      m.From,
			To:        m.To,
			Type:      string(m.Kind),
			Timestamp: m.SentAt.Format(time.RFC3339),
			Content:   m.Body,
		})
	}

	return map[string]interface{}{
		"team": map[string]string{
			"name":        name,
			"leaderName":  leaderName,
			"description": cfg.Description,
		},
		"members":     members,
		"tasks":       taskGroups,
		"taskSummary": summary,
		"messages":    messages,
	}, nil
}
