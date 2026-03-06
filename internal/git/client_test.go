package git

import (
	"reflect"
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
