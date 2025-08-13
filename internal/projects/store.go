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
	Name string `json:"name"`
	URL  string `json:"url"`
}

// NewStore creates a store under ~/.config/agent-manger/config.json
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	base := filepath.Join(home, ".config", "agent-manger")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, fmt.Errorf("ensure base dir: %w", err)
	}
	fp := filepath.Join(base, "config.json")
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(fp, []byte("{}"), 0o644); err != nil {
			return nil, fmt.Errorf("init config file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat config file: %w", err)
	}
	return &Store{filePath: fp}, nil
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

func (s *Store) load() (map[string]string, error) {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("read store: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse store: %w", err)
	}
	if m == nil {
		m = make(map[string]string)
	}
	return m, nil
}

func (s *Store) save(m map[string]string) error {
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
	m[name] = url
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
	m[name] = url
	return s.save(m)
}

func (s *Store) Get(name string) (string, bool, error) {
	m, err := s.load()
	if err != nil {
		return "", false, err
	}
	u, ok := m[name]
	return u, ok, nil
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
		out = append(out, Project{Name: k, URL: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
