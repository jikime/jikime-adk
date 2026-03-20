package team

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

// Inbox manages a FIFO message queue for a single agent.
// Messages are stored as JSON files under:
//
//	~/.jikime/teams/<team>/inbox/<agentID>/
//
// File names are prefixed with an RFC3339Nano timestamp so that
// os.ReadDir returns them in chronological order.
type Inbox struct {
	dir string // absolute path to this agent's inbox directory
}

// NewInbox returns an Inbox rooted at dir, creating the directory as needed.
func NewInbox(dir string) (*Inbox, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("team/inbox: mkdir %s: %w", dir, err)
	}
	return &Inbox{dir: dir}, nil
}

// InboxDir returns the inbox directory for agentID inside teamDir.
func InboxDir(teamDir, agentID string) string {
	return filepath.Join(teamDir, "inbox", agentID)
}

// Send writes a message to this inbox atomically (tmp + rename).
func (b *Inbox) Send(msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now()
	}

	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("team/inbox: marshal: %w", err)
	}

	// Prefix with timestamp so files sort in arrival order.
	ts := msg.SentAt.UTC().Format("20060102T150405.000000000Z")
	name := fmt.Sprintf("%s-%s.json", ts, msg.ID)
	dst := filepath.Join(b.dir, name)
	tmp := dst + ".tmp"

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("team/inbox: write tmp: %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("team/inbox: rename: %w", err)
	}
	return nil
}

// Receive reads up to limit messages from the inbox in FIFO order and
// removes them from disk. Pass limit ≤ 0 to receive all pending messages.
func (b *Inbox) Receive(limit int) ([]*Message, error) {
	files, err := b.sortedFiles()
	if err != nil {
		return nil, err
	}
	var msgs []*Message
	for _, f := range files {
		if limit > 0 && len(msgs) >= limit {
			break
		}
		path := filepath.Join(b.dir, f)
		msg, err := readMessage(path)
		if err != nil {
			continue // skip corrupt files
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return msgs, fmt.Errorf("team/inbox: remove %s: %w", f, err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// Peek returns up to limit messages without removing them.
// Pass limit ≤ 0 to peek at all pending messages.
func (b *Inbox) Peek(limit int) ([]*Message, error) {
	files, err := b.sortedFiles()
	if err != nil {
		return nil, err
	}
	var msgs []*Message
	for _, f := range files {
		if limit > 0 && len(msgs) >= limit {
			break
		}
		msg, err := readMessage(filepath.Join(b.dir, f))
		if err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// Count returns the number of pending (unread) messages.
func (b *Inbox) Count() (int, error) {
	files, err := b.sortedFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// Watch calls fn for each new message that arrives in the inbox.
// It blocks until ctx.Done() is closed or an unrecoverable error occurs.
// Each received message is consumed (removed from disk).
func (b *Inbox) Watch(done <-chan struct{}, fn func(*Message)) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("team/inbox: watcher: %w", err)
	}
	defer w.Close()

	if err := w.Add(b.dir); err != nil {
		return fmt.Errorf("team/inbox: watch dir %s: %w", b.dir, err)
	}

	// Drain any messages already present before the watcher started.
	if msgs, err := b.Receive(0); err == nil {
		for _, m := range msgs {
			fn(m)
		}
	}

	for {
		select {
		case <-done:
			return nil
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
				continue
			}
			if filepath.Ext(event.Name) != ".json" {
				continue
			}
			// Small sleep to let the writer finish the atomic rename.
			time.Sleep(5 * time.Millisecond)
			msgs, err := b.Receive(0)
			if err != nil {
				continue
			}
			for _, m := range msgs {
				fn(m)
			}
		case <-w.Errors:
			// Non-fatal watcher errors are ignored.
		}
	}
}

// sortedFiles returns JSON file names in chronological order (excluding .tmp).
func (b *Inbox) sortedFiles() ([]string, error) {
	entries, err := os.ReadDir(b.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/inbox: readdir %s: %w", b.dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if filepath.Ext(n) == ".json" {
			names = append(names, n)
		}
	}
	sort.Strings(names) // timestamp prefix ensures chronological order
	return names, nil
}

func readMessage(path string) (*Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// eventLogPath returns the path for the team-wide message event log.
func eventLogPath(teamDir string) string {
	return filepath.Join(teamDir, "inbox", "event-log.jsonl")
}

// appendEventLog appends a single message entry to the team event log (JSONL).
// Errors are silently ignored so that messaging always succeeds.
func appendEventLog(teamDir string, msg *Message) {
	line, err := json.Marshal(msg)
	if err != nil {
		return
	}
	line = append(line, '\n')
	// Ensure inbox directory exists.
	_ = os.MkdirAll(filepath.Join(teamDir, "inbox"), 0o755)
	f, err := os.OpenFile(eventLogPath(teamDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(line)
}

// TeamInbox is a helper that manages inboxes for all agents in a team.
type TeamInbox struct {
	teamDir string
}

// NewTeamInbox returns a TeamInbox for the given team directory.
func NewTeamInbox(teamDir string) *TeamInbox {
	return &TeamInbox{teamDir: teamDir}
}

// For returns the Inbox for a specific agent, creating its directory if needed.
func (ti *TeamInbox) For(agentID string) (*Inbox, error) {
	return NewInbox(InboxDir(ti.teamDir, agentID))
}

// Send delivers a message to the recipient's inbox and records it in the event log.
func (ti *TeamInbox) Send(msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now()
	}
	ib, err := ti.For(msg.To)
	if err != nil {
		return err
	}
	if err := ib.Send(msg); err != nil {
		return err
	}
	appendEventLog(ti.teamDir, msg)
	return nil
}

// Broadcast sends msg to all agentIDs except msg.From and records in the event log.
func (ti *TeamInbox) Broadcast(msg *Message, agentIDs []string) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now()
	}
	// Record the broadcast event once (not per-recipient copy).
	appendEventLog(ti.teamDir, msg)

	for _, id := range agentIDs {
		if id == msg.From {
			continue
		}
		cp := *msg
		cp.ID = uuid.New().String()
		cp.To = id
		cp.Kind = MessageKindBroadcast
		cp.SentAt = time.Now()
		ib, err := ti.For(id)
		if err != nil {
			return err
		}
		if err := ib.Send(&cp); err != nil {
			return err
		}
	}
	return nil
}

// EventLog returns up to limit messages from the team event log (oldest-first).
// Pass fromAgent to filter by sender; pass limit ≤ 0 for all entries.
func (ti *TeamInbox) EventLog(limit int, fromAgent string) ([]*Message, error) {
	path := eventLogPath(ti.teamDir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/inbox: event log read: %w", err)
	}

	var msgs []*Message
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var m Message
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		if fromAgent != "" && m.From != fromAgent {
			continue
		}
		msgs = append(msgs, &m)
	}

	// Event log is oldest-first (appended in order); apply limit from the end.
	if limit > 0 && len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}
	return msgs, nil
}

// splitLines splits JSONL content into non-empty lines.
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := data[start:i]
			if len(line) > 0 {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
