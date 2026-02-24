package ai

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Client struct {
	command string
}

func NewClient(command string) *Client {
	return &Client{command: command}
}

func (c *Client) Launch(workspacePath string) error {
	cmd := exec.Command("sh", "-c", c.command)
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
