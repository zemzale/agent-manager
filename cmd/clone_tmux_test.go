package cmd

import (
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

func TestSanitizeTmuxName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple", input: "Project_01", expected: "project_01"},
		{name: "collapses punctuation", input: "my/project::name", expected: "my-project-name"},
		{name: "fallback", input: "!!!", expected: "project"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitizeTmuxName(tc.input); got != tc.expected {
				t.Fatalf("sanitizeTmuxName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestBuildTmuxDispatchCommand(t *testing.T) {
	cmd := buildTmuxDispatchCommand("/usr/local/bin/agent-manager", []string{"clone", "my'repo", "--cmd", "opencode ."})
	expected := "'/usr/local/bin/agent-manager' 'clone' 'my'\"'\"'repo' '--cmd' 'opencode .' '--tmux-dispatched'"
	if cmd != expected {
		t.Fatalf("unexpected command\n got: %s\nwant: %s", cmd, expected)
	}

	withFlag := buildTmuxDispatchCommand("agent-manager", []string{"clone", "demo", "--tmux-dispatched"})
	expectedWithFlag := "'agent-manager' 'clone' 'demo' '--tmux-dispatched'"
	if withFlag != expectedWithFlag {
		t.Fatalf("expected no duplicate flag\n got: %s\nwant: %s", withFlag, expectedWithFlag)
	}
}
