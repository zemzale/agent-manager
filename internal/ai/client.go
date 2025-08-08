package ai

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Client struct {
	tool string
}

func NewClient(tool string) *Client {
	return &Client{tool: tool}
}

func (c *Client) Launch(workspacePath string) error {
	// Change to the workspace directory
	if err := os.Chdir(workspacePath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Launch the AI tool
	var cmd *exec.Cmd
	switch c.tool {
	case "opencode":
		cmd = exec.Command("opencode", ".")
	case "cursor":
		cmd = exec.Command("cursor", ".")
	case "code":
		cmd = exec.Command("code", ".")
	default:
		return fmt.Errorf("unsupported AI tool: %s", c.tool)
	}

	// Set up the command to run in the foreground
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command and wait for it to complete
	if err := cmd.Run(); err != nil {
		// Check if it's an exit error (tool was closed normally)
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				// If the tool was closed normally (exit code 0 or 1), don't treat as error
				if status.ExitStatus() <= 1 {
					return nil
				}
			}
		}
		return fmt.Errorf("failed to launch %s: %w", c.tool, err)
	}

	return nil
}
