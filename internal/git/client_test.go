package git

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCloneArgs(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		destination string
		branch      string
		expected    []string
	}{
		{
			name:        "without branch",
			url:         "https://github.com/user/repo.git",
			destination: "/tmp/repo",
			branch:      "",
			expected:    []string{"clone", "https://github.com/user/repo.git", "/tmp/repo"},
		},
		{
			name:        "with branch",
			url:         "git@github.com:user/repo.git",
			destination: "/tmp/repo",
			branch:      "develop",
			expected:    []string{"clone", "--branch", "develop", "git@github.com:user/repo.git", "/tmp/repo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cloneArgs(tc.url, tc.destination, tc.branch)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("cloneArgs() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestParseLSRemoteHeads(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name: "parses and sorts branches",
			output: "a1 refs/heads/feature/z\n" +
				"b2 refs/heads/main\n" +
				"c3 refs/heads/feature/a\n",
			expected: []string{"feature/a", "feature/z", "main"},
		},
		{
			name: "deduplicates and ignores non-head refs",
			output: "a1 refs/heads/main\n" +
				"b2 refs/tags/v1\n" +
				"c3 refs/heads/main\n",
			expected: []string{"main"},
		},
		{
			name:     "empty output",
			output:   "",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLSRemoteHeads([]byte(tc.output))
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("parseLSRemoteHeads() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestParseLSRemoteDefaultBranch(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "parses default branch from symref",
			output:   "ref: refs/heads/main\tHEAD\n1a2b3c HEAD\n",
			expected: "main",
		},
		{
			name:     "missing symref returns empty",
			output:   "1a2b3c HEAD\n",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLSRemoteDefaultBranch([]byte(tc.output))
			if got != tc.expected {
				t.Fatalf("parseLSRemoteDefaultBranch() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestCloneReturnsGitOutputOnFailure(t *testing.T) {
	client := NewClient()
	err := client.Clone("file:///tmp/this-repo-should-not-exist.git", t.TempDir(), "")
	if err == nil {
		t.Fatal("Clone() error = nil, want failure")
	}

	var exitErr interface{ ExitCode() int }
	if !errors.As(err, &exitErr) {
		t.Fatalf("Clone() error = %T, want wrapped exit error", err)
	}

	if !strings.Contains(err.Error(), "git clone failed") {
		t.Fatalf("Clone() error = %q, want git clone context", err)
	}

	if !strings.Contains(err.Error(), "does not appear to be a git repository") {
		t.Fatalf("Clone() error = %q, want git stderr output", err)
	}
}
