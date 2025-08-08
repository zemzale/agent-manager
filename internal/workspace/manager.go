package workspace

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Workspace struct {
	ID   string
	Path string
	URL  string
}

type Manager struct {
	baseDir string
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".agent-manager", "workspaces")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	return &Manager{baseDir: baseDir}, nil
}

func (m *Manager) Create(gitURL string) (*Workspace, error) {
	// Generate unique workspace ID
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate workspace ID: %w", err)
	}

	// Extract repository name from URL
	repoName := extractRepoName(gitURL)

	// Create workspace directory
	workspaceName := fmt.Sprintf("%s-%s", repoName, id)
	workspacePath := filepath.Join(m.baseDir, workspaceName)

	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	return &Workspace{
		ID:   id,
		Path: workspacePath,
		URL:  gitURL,
	}, nil
}

func (m *Manager) Remove(id string) error {
	// Find workspace by ID
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read workspaces directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-"+id) {
			workspacePath := filepath.Join(m.baseDir, entry.Name())
			if err := os.RemoveAll(workspacePath); err != nil {
				return fmt.Errorf("failed to remove workspace: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("workspace with ID %s not found", id)
}

func (m *Manager) List() ([]*Workspace, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspaces directory: %w", err)
	}

	var workspaces []*Workspace
	for _, entry := range entries {
		if entry.IsDir() {
			// Extract ID from directory name (format: reponame-id)
			parts := strings.Split(entry.Name(), "-")
			if len(parts) >= 2 {
				id := parts[len(parts)-1]
				workspaces = append(workspaces, &Workspace{
					ID:   id,
					Path: filepath.Join(m.baseDir, entry.Name()),
				})
			}
		}
	}

	return workspaces, nil
}

func generateID() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

func extractRepoName(gitURL string) string {
	// Handle different URL formats
	if strings.HasPrefix(gitURL, "http://") || strings.HasPrefix(gitURL, "https://") {
		u, err := url.Parse(gitURL)
		if err != nil {
			return "repo"
		}
		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	} else if strings.Contains(gitURL, ":") {
		// SSH format: git@github.com:user/repo.git
		parts := strings.Split(gitURL, ":")
		if len(parts) >= 2 {
			path := parts[len(parts)-1]
			path = strings.TrimSuffix(path, ".git")
			pathParts := strings.Split(path, "/")
			if len(pathParts) > 0 {
				return pathParts[len(pathParts)-1]
			}
		}
	}

	return "repo"
}
