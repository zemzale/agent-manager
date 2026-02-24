package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type Store struct {
	filePath string
}

type Project struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Commands []string `json:"commands,omitempty"`
}

func (s *Store) FilePath() string { return s.filePath }

// NewStore creates a store under ~/.config/agent-manger/config.json
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	base := filepath.Join(home, ".config", "agent-manger")
	return NewStoreWithBaseDir(base)
}

// NewStoreWithBaseDir is intended for tests; it stores under baseDir/config.json.
func NewStoreWithBaseDir(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("ensure base dir: %w", err)
	}
	fp := filepath.Join(baseDir, "config.json")
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(fp, []byte("{}"), 0o644); err != nil {
			return nil, fmt.Errorf("init config file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat config file: %w", err)
	}
	return &Store{filePath: fp}, nil
}

func (s *Store) load() (map[string]Project, error) {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("read store: %w", err)
	}

	// Try new format first
	var m map[string]Project
	if err := json.Unmarshal(b, &m); err == nil && m != nil {
		return m, nil
	}

	// Fall back to old format (map[string]string) for migration
	var old map[string]string
	if err := json.Unmarshal(b, &old); err != nil {
		return nil, fmt.Errorf("parse store: %w", err)
	}

	m = make(map[string]Project, len(old))
	for name, url := range old {
		m[name] = Project{URL: url}
	}
	return m, nil
}

func (s *Store) save(m map[string]Project) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	if err := os.WriteFile(s.filePath, b, 0o644); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	return nil
}

var nameRe = regexp.MustCompile(`^[a-zA-Z0-9._\-]+$`)

func validateName(name string) error {
	if name == "" || len(name) > 200 || !nameRe.MatchString(name) {
		return fmt.Errorf("invalid project name %q; use [a-zA-Z0-9._-]", name)
	}
	return nil
}

// Add adds a new project; errors if it already exists.
func (s *Store) Add(name, url string) error {
	if err := validateName(name); err != nil {
		return err
	}
	m, err := s.load()
	if err != nil {
		return err
	}
	if _, exists := m[name]; exists {
		return fmt.Errorf("project %q already exists", name)
	}
	m[name] = Project{URL: url}
	return s.save(m)
}

// Set creates or overwrites a project mapping.
func (s *Store) Set(name, url string) error {
	if err := validateName(name); err != nil {
		return err
	}
	m, err := s.load()
	if err != nil {
		return err
	}
	p := m[name]
	p.URL = url
	m[name] = p
	return s.save(m)
}

func (s *Store) Get(name string) (string, bool, error) {
	m, err := s.load()
	if err != nil {
		return "", false, err
	}
	p, ok := m[name]
	return p.URL, ok, nil
}

func (s *Store) GetProject(name string) (Project, bool, error) {
	m, err := s.load()
	if err != nil {
		return Project{}, false, err
	}
	p, ok := m[name]
	p.Name = name
	return p, ok, nil
}

func (s *Store) Remove(name string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := m[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	delete(m, name)
	return s.save(m)
}

func (s *Store) List() ([]Project, error) {
	m, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]Project, 0, len(m))
	for k, v := range m {
		out = append(out, Project{Name: k, URL: v.URL, Commands: v.Commands})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Store) AddCommand(name, command string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	p, ok := m[name]
	if !ok {
		return fmt.Errorf("project %q not found", name)
	}
	p.Commands = append(p.Commands, command)
	m[name] = p
	return s.save(m)
}

func (s *Store) RemoveCommand(name string, index int) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	p, ok := m[name]
	if !ok {
		return fmt.Errorf("project %q not found", name)
	}
	if index < 0 || index >= len(p.Commands) {
		return fmt.Errorf("invalid command index %d", index)
	}
	p.Commands = append(p.Commands[:index], p.Commands[index+1:]...)
	m[name] = p
	return s.save(m)
}
