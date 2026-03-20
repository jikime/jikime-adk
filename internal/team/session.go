package team

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// SessionStore saves and restores team state snapshots.
// Each session is stored at ~/.jikime/sessions/<teamName>/<sessionID>.json.
type SessionStore struct {
	sessDir string
}

// NewSessionStore returns a SessionStore for the given team.
// sessDir should be ~/.jikime/sessions/<teamName>/.
func NewSessionStore(sessDir string) (*SessionStore, error) {
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		return nil, fmt.Errorf("team/session: mkdir %s: %w", sessDir, err)
	}
	return &SessionStore{sessDir: sessDir}, nil
}

// Save writes a new session snapshot and returns its ID.
func (s *SessionStore) Save(teamName, description string, tasks []*Task, agents []*AgentInfo) (string, error) {
	sess := &Session{
		ID:          uuid.New().String(),
		TeamName:    teamName,
		Description: description,
		SavedAt:     time.Now(),
	}
	for _, t := range tasks {
		sess.Tasks = append(sess.Tasks, *t)
	}
	for _, a := range agents {
		sess.Agents = append(sess.Agents, *a)
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return "", fmt.Errorf("team/session: marshal: %w", err)
	}
	tmp := filepath.Join(s.sessDir, sess.ID+".json.tmp")
	dst := filepath.Join(s.sessDir, sess.ID+".json")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", fmt.Errorf("team/session: write: %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("team/session: rename: %w", err)
	}
	return sess.ID, nil
}

// Load returns the session with the given ID, or (nil, nil) if not found.
func (s *SessionStore) Load(sessionID string) (*Session, error) {
	data, err := os.ReadFile(filepath.Join(s.sessDir, sessionID+".json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/session: read %s: %w", sessionID, err)
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("team/session: unmarshal %s: %w", sessionID, err)
	}
	return &sess, nil
}

// Latest returns the most recently saved session, or nil if none exist.
func (s *SessionStore) Latest() (*Session, error) {
	entries, err := os.ReadDir(s.sessDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/session: readdir: %w", err)
	}
	var latest *Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		sess, err := s.Load(id)
		if err != nil || sess == nil {
			continue
		}
		if latest == nil || sess.SavedAt.After(latest.SavedAt) {
			latest = sess
		}
	}
	return latest, nil
}

// List returns all saved sessions, newest first.
func (s *SessionStore) List() ([]*Session, error) {
	entries, err := os.ReadDir(s.sessDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("team/session: readdir: %w", err)
	}
	var sessions []*Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		sess, err := s.Load(id)
		if err != nil || sess == nil {
			continue
		}
		sessions = append(sessions, sess)
	}
	// Sort newest first.
	for i := 0; i < len(sessions); i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[j].SavedAt.After(sessions[i].SavedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}
	return sessions, nil
}

// Delete removes a session file.
func (s *SessionStore) Delete(sessionID string) error {
	err := os.Remove(filepath.Join(s.sessDir, sessionID+".json"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// MarkRestored records the restoration time on a session.
func (s *SessionStore) MarkRestored(sessionID string) error {
	sess, err := s.Load(sessionID)
	if err != nil || sess == nil {
		return err
	}
	now := time.Now()
	sess.RestoredAt = &now
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.sessDir, sessionID+".json"), data, 0o644)
}
