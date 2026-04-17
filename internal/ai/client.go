package ai

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Client struct {
	command string
}

func NewClient(command string) *Client {
	return &Client{command: command}
}

func (c *Client) Launch(workspacePath string) error {
	cmd := exec.Command("sh", "-c", ResolveCommand(c.command, workspacePath))
	cmd.Dir = workspacePath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() <= 1 {
					return nil
				}
			}
		}
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}

func ResolveCommand(command, workspacePath string) string {
	trimmed := strings.TrimSpace(command)
	if trimmed == "opencode ." {
		return "opencode " + shellQuote(workspacePath)
	}

	if strings.Contains(command, "{workspace}") {
		return strings.ReplaceAll(command, "{workspace}", shellQuote(workspacePath))
	}

	return command
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
