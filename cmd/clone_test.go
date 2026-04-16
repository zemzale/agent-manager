package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSetupCommandsContinuesAfterFailure(t *testing.T) {
	workspacePath := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	commands := []string{
		"touch first.txt",
		"cp missing-file.txt missing-copy.txt",
		"touch second.txt",
	}

	failures := runSetupCommands(workspacePath, commands, &stdout, &stderr)
	if len(failures) != 1 {
		t.Fatalf("runSetupCommands() failures = %d, want 1", len(failures))
	}

	if failures[0].command != commands[1] {
		t.Fatalf("runSetupCommands() failed command = %q, want %q", failures[0].command, commands[1])
	}

	for _, name := range []string{"first.txt", "second.txt"} {
		if _, err := os.Stat(filepath.Join(workspacePath, name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}

	if !strings.Contains(stderr.String(), "Warning: setup command failed") {
		t.Fatalf("stderr = %q, want failure warning", stderr.String())
	}
}
