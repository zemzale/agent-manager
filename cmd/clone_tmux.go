package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
)

func launchCloneInTmux(gitURL, workspacePath, command string, project *projects.Project) (bool, error) {
	if os.Getenv("TMUX") == "" {
		return false, nil
	}

	resolvedCommand := ai.ResolveCommand(command, workspacePath)
	sessionName := tmuxSessionName(project, gitURL)
	windowName := fmt.Sprintf("clone-%s", time.Now().Format("150405"))

	exists, err := tmuxSessionExists(sessionName)
	if err != nil {
		return false, err
	}

	var target string
	if exists {
		target, err = tmuxCreateWindow(sessionName, windowName, workspacePath)
	} else {
		target, err = tmuxCreateSessionWithWindow(sessionName, windowName, workspacePath)
	}
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

func tmuxSessionName(project *projects.Project, gitURL string) string {
	projectKey := projectKey(project, gitURL)
	return "am-" + sanitizeTmuxName(projectKey)
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

func sanitizeTmuxName(name string) string {
	var b strings.Builder
	lastDash := false

	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
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

func tmuxSessionExists(sessionName string) (bool, error) {
	cmd := exec.Command("tmux", "has-session", "-t", "="+sessionName)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, nil
	}

	return false, fmt.Errorf("check tmux session %q: %w", sessionName, err)
}

func tmuxCreateWindow(sessionName, windowName, workspacePath string) (string, error) {
	cmd := exec.Command(
		"tmux",
		"new-window",
		"-P",
		"-F",
		"#{session_name}:#{window_index}",
		"-t",
		"="+sessionName,
		"-c",
		workspacePath,
		"-n",
		windowName,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("create tmux window in %q: %w", sessionName, err)
	}

	target := strings.TrimSpace(string(out))
	if target == "" {
		return "", fmt.Errorf("tmux did not return window target")
	}

	return target, nil
}

func tmuxCreateSessionWithWindow(sessionName, windowName, workspacePath string) (string, error) {
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
