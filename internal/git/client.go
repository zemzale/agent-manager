package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Clone(url, destination, branch string) error {
	cmd := exec.Command("git", cloneArgs(url, destination, branch)...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

func cloneArgs(url, destination, branch string) []string {
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, url, destination)
	return args
}

func (c *Client) ListRemotes() ([]string, error) {
	cmd := exec.Command("git", "remote")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.TrimSpace(string(out))
	if lines == "" {
		return nil, nil
	}
	return strings.Split(lines, "\n"), nil
}

func (c *Client) GetRemoteURL(name string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", name)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get remote url: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) ListRemoteBranches(url string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", url)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list remote branches: %w", err)
	}
	return parseLSRemoteHeads(out), nil
}

func (c *Client) GetRemoteDefaultBranch(url string) (string, error) {
	cmd := exec.Command("git", "ls-remote", "--symref", url, "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get remote default branch: %w", err)
	}

	return parseLSRemoteDefaultBranch(out), nil
}

func parseLSRemoteHeads(out []byte) []string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}

	branches := make([]string, 0, len(lines))
	seen := make(map[string]struct{})

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		const branchPrefix = "refs/heads/"
		ref := fields[1]
		if !strings.HasPrefix(ref, branchPrefix) {
			continue
		}

		branch := strings.TrimPrefix(ref, branchPrefix)
		if branch == "" {
			continue
		}

		if _, exists := seen[branch]; exists {
			continue
		}

		seen[branch] = struct{}{}
		branches = append(branches, branch)
	}

	sort.Strings(branches)
	return branches
}

func parseLSRemoteDefaultBranch(out []byte) string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	const symrefPrefix = "ref: refs/heads/"

	for _, line := range lines {
		if !strings.HasPrefix(line, symrefPrefix) {
			continue
		}

		branchWithSuffix := strings.TrimPrefix(line, symrefPrefix)
		fields := strings.Fields(branchWithSuffix)
		if len(fields) == 0 {
			continue
		}

		return fields[0]
	}

	return ""
}
