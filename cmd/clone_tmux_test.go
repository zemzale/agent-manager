package cmd

import (
	"os"
	"testing"

	"github.com/zemzale/agent-manager/internal/projects"
)

func TestProjectKey(t *testing.T) {
	t.Run("uses explicit project name", func(t *testing.T) {
		key := projectKey(&projects.Project{Name: "my-project"}, "https://github.com/user/repo.git")
		if key != "my-project" {
			t.Fatalf("expected project key to prefer project name, got %q", key)
		}
	})

	t.Run("falls back to repository name", func(t *testing.T) {
		key := projectKey(nil, "git@github.com:user/demo-repo.git")
		if key != "demo-repo" {
			t.Fatalf("expected repo-derived key, got %q", key)
		}
	})
}

func TestProjectSlug(t *testing.T) {
	t.Run("uses explicit project name", func(t *testing.T) {
		got := projectSlug(&projects.Project{Name: "My Project_01"}, "https://github.com/user/repo.git")
		if got != "my-project-01" {
			t.Fatalf("projectSlug() = %q, want %q", got, "my-project-01")
		}
	})

	t.Run("falls back to repository name", func(t *testing.T) {
		got := projectSlug(nil, "git@github.com:user/demo-repo.git")
		if got != "demo-repo" {
			t.Fatalf("projectSlug() = %q, want %q", got, "demo-repo")
		}
	})
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple", input: "Project_01", expected: "project-01"},
		{name: "collapses punctuation", input: "my/project::name", expected: "my-project-name"},
		{name: "fallback", input: "!!!", expected: "project"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := slugify(tc.input); got != tc.expected {
				t.Fatalf("slugify(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestTmuxSessionName(t *testing.T) {
	got := tmuxSessionName("demo-repo", "abc12345")
	if got != "am-demo-repo-abc12345" {
		t.Fatalf("tmuxSessionName() = %q, want %q", got, "am-demo-repo-abc12345")
	}
}

func TestLaunchCloneInTmuxNoopOutsideTmux(t *testing.T) {
	original := os.Getenv("TMUX")
	t.Cleanup(func() {
		if original == "" {
			_ = os.Unsetenv("TMUX")
			return
		}
		_ = os.Setenv("TMUX", original)
	})

	if err := os.Unsetenv("TMUX"); err != nil {
		t.Fatalf("unset TMUX: %v", err)
	}

	launched, err := launchCloneInTmux("https://github.com/user/repo.git", "abc12345", "/tmp/repo", "opencode {workspace}", nil)
	if err != nil {
		t.Fatalf("launchCloneInTmux() error = %v", err)
	}
	if launched {
		t.Fatal("launchCloneInTmux() = true, want false outside tmux")
	}
}
