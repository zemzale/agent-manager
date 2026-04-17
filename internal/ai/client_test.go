package ai

import "testing"

func TestResolveCommand(t *testing.T) {
	workspacePath := "/tmp/my project's workspace"

	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "legacy opencode default uses explicit workspace",
			command:  "opencode .",
			expected: "opencode '/tmp/my project'\"'\"'s workspace'",
		},
		{
			name:     "workspace placeholder is substituted",
			command:  "opencode {workspace}",
			expected: "opencode '/tmp/my project'\"'\"'s workspace'",
		},
		{
			name:     "custom command remains unchanged without placeholder",
			command:  "my-tool --flag",
			expected: "my-tool --flag",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveCommand(tc.command, workspacePath); got != tc.expected {
				t.Fatalf("ResolveCommand(%q) = %q, want %q", tc.command, got, tc.expected)
			}
		})
	}
}
