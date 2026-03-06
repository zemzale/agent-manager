package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
)

func dispatchCloneToTmux(gitURL string, project *projects.Project) (bool, error) {
	if tmuxDispatched || os.Getenv("TMUX") == "" {
		return false, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("resolve executable for tmux dispatch: %w", err)
	}

	command := buildTmuxDispatchCommand(exe, os.Args[1:])
	sessionName := tmuxSessionName(project, gitURL)
	windowName := fmt.Sprintf("clone-%s", time.Now().Format("150405"))

	exists, err := tmuxSessionExists(sessionName)
	if err != nil {
		return false, err
	}

	var target string
	if exists {
		target, err = tmuxCreateWindow(sessionName, windowName, command)
	} else {
		target, err = tmuxCreateSessionWithWindow(sessionName, windowName, command)
	}
	if err != nil {
		return false, err
	}

	if err := tmuxSwitchClient(target); err != nil {
		return false, err
	}

	fmt.Printf("Started clone in tmux window %s\n", target)
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

func buildTmuxDispatchCommand(executable string, args []string) string {
	clonedArgs := append([]string{}, args...)
	if !hasFlag(clonedArgs, "--tmux-dispatched") {
		clonedArgs = append(clonedArgs, "--tmux-dispatched")
	}

	parts := append([]string{executable}, clonedArgs...)
	return shellJoin(parts)
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

func shellJoin(parts []string) string {
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		quoted = append(quoted, shellQuote(p))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
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

func tmuxCreateWindow(sessionName, windowName, command string) (string, error) {
	cmd := exec.Command(
		"tmux",
		"new-window",
		"-P",
		"-F",
		"#{session_name}:#{window_index}",
		"-t",
		"="+sessionName,
		"-n",
		windowName,
		command,
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

func tmuxCreateSessionWithWindow(sessionName, windowName, command string) (string, error) {
	cmd := exec.Command(
		"tmux",
		"new-session",
		"-d",
		"-P",
		"-F",
		"#{session_name}:#{window_index}",
		"-s",
		sessionName,
		"-n",
		windowName,
		command,
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

func tmuxSwitchClient(target string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("switch tmux client to %q: %w", target, err)
	}
	return nil
}
