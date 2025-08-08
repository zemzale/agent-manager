package git

import (
	"fmt"
	"os/exec"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Clone(url, destination string) error {
	cmd := exec.Command("git", "clone", url, destination)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}
