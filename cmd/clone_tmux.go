package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
)

func launchCloneInTmux(gitURL, workspaceID, workspacePath, command string, project *projects.Project) (bool, error) {
	if os.Getenv("TMUX") == "" {
		return false, nil
	}

	resolvedCommand := ai.ResolveCommand(command, workspacePath)
	slug := projectSlug(project, gitURL)
	sessionName := tmuxSessionName(slug, workspaceID)
	target, err := tmuxCreateSession(sessionName, slug, workspacePath)
	if err != nil {
		return false, err
	}

	if err := tmuxSendKeys(target, resolvedCommand); err != nil {
		return false, err
	}

	if err := tmuxSwitchClient(target); err != nil {
		return false, err
	}

	fmt.Printf("Opened tmux window %s\n", target)
	return true, nil
}

func tmuxSessionName(projectSlug, workspaceID string) string {
	return fmt.Sprintf("am-%s-%s", projectSlug, workspaceID)
}

func projectSlug(project *projects.Project, gitURL string) string {
	return slugify(projectKey(project, gitURL))
}

func projectKey(project *projects.Project, gitURL string) string {
	if project != nil && project.Name != "" {
		return project.Name
	}

	repo := git.ExtractRepoName(gitURL)
	if repo == "" {
		return "project"
	}

	return repo
}

func slugify(name string) string {
	var b strings.Builder
	lastDash := false

	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	clean := strings.Trim(b.String(), "-")
	if clean == "" {
		return "project"
	}

	return clean
}

func tmuxCreateSession(sessionName, windowName, workspacePath string) (string, error) {
	cmd := exec.Command(
		"tmux",
		"new-session",
		"-d",
		"-P",
		"-F",
		"#{session_name}:#{window_index}",
		"-s",
		sessionName,
		"-c",
		workspacePath,
		"-n",
		windowName,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("create tmux session %q: %w", sessionName, err)
	}

	target := strings.TrimSpace(string(out))
	if target == "" {
		return "", fmt.Errorf("tmux did not return session target")
	}

	return target, nil
}

func tmuxSendKeys(target, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, command, "C-m")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("send command to tmux target %q: %w", target, err)
	}
	return nil
}

func tmuxSwitchClient(target string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("switch tmux client to %q: %w", target, err)
	}
	return nil
}
