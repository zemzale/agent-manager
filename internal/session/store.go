package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	filePath string
}

type Session struct {
	ID        string    `json:"id"`
	GitURL    string    `json:"git_url"`
	Project   string    `json:"project,omitempty"`
	Tool      string    `json:"tool"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"created_at"`
}

type sessionsData struct {
	Sessions []Session `json:"sessions"`
}

// NewStore creates a store under ~/.config/agent-manger/sessions.json
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	base := filepath.Join(home, ".config", "agent-manger")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, fmt.Errorf("ensure base dir: %w", err)
	}
	fp := filepath.Join(base, "sessions.json")
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(fp, []byte(`{"sessions": []}`), 0o644); err != nil {
			return nil, fmt.Errorf("init sessions file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat sessions file: %w", err)
	}
	return &Store{filePath: fp}, nil
}

// NewStoreWithBaseDir is intended for tests; it stores under baseDir/sessions.json.
func NewStoreWithBaseDir(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("ensure base dir: %w", err)
	}
	fp := filepath.Join(baseDir, "sessions.json")
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(fp, []byte(`{"sessions": []}`), 0o644); err != nil {
			return nil, fmt.Errorf("init sessions file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat sessions file: %w", err)
	}
	return &Store{filePath: fp}, nil
}

func (s *Store) load() (*sessionsData, error) {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("read store: %w", err)
	}
	var data sessionsData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse store: %w", err)
	}
	return &data, nil
}

func (s *Store) save(data *sessionsData) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	if err := os.WriteFile(s.filePath, b, 0o644); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	return nil
}

// Add adds a new session to the history (keeps only last 50 sessions)
func (s *Store) Add(session Session) error {
	data, err := s.load()
	if err != nil {
		return err
	}

	// Add new session at the beginning
	session.CreatedAt = time.Now()
	data.Sessions = append([]Session{session}, data.Sessions...)

	// Keep only last 50 sessions
	if len(data.Sessions) > 50 {
		data.Sessions = data.Sessions[:50]
	}

	return s.save(data)
}

// List returns all sessions sorted by creation time (newest first)
func (s *Store) List() ([]Session, error) {
	data, err := s.load()
	if err != nil {
		return nil, err
	}
	return data.Sessions, nil
}

// Get returns a session by ID
func (s *Store) Get(id string) (*Session, error) {
	data, err := s.load()
	if err != nil {
		return nil, err
	}
	for _, session := range data.Sessions {
		if session.ID == id {
			return &session, nil
		}
	}
	return nil, fmt.Errorf("session %q not found", id)
}

// Remove removes a session by ID
func (s *Store) Remove(id string) error {
	data, err := s.load()
	if err != nil {
		return err
	}

	found := false
	newSessions := make([]Session, 0, len(data.Sessions))
	for _, session := range data.Sessions {
		if session.ID != id {
			newSessions = append(newSessions, session)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("session %q not found", id)
	}

	data.Sessions = newSessions
	return s.save(data)
}

// Clear removes all sessions
func (s *Store) Clear() error {
	data := &sessionsData{Sessions: []Session{}}
	return s.save(data)
}

// GetRecentProjectNames returns unique project names from recent sessions
func (s *Store) GetRecentProjectNames() ([]string, error) {
	data, err := s.load()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	names := []string{}

	for _, session := range data.Sessions {
		if session.Project != "" && !seen[session.Project] {
			seen[session.Project] = true
			names = append(names, session.Project)
		}
	}

	return names, nil
}
