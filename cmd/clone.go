package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
	"github.com/zemzale/agent-manager/internal/session"
	"github.com/zemzale/agent-manager/internal/ui"
	"github.com/zemzale/agent-manager/internal/workspace"
)

var (
	aiCommand      string
	cloneBranch    string
	skipSetup      bool
	tmuxDispatched bool
)

const (
	cloneCustomRepoChoice    = "__custom_repo__"
	cloneDefaultBranchChoice = "__default_branch__"
	cloneManualBranchChoice  = "__manual_branch__"
)

type repoSelection struct {
	gitURL  string
	project *projects.Project
	saved   bool
}

var cloneCmd = &cobra.Command{
	Use:   "clone [git-url|project-name]",
	Short: "Clone a repository and launch an AI tool",
	Long: `Clone a Git repository into a temporary workspace and launch an AI tool for development.
The workspace will be created in ~/.agent-manager/workspaces/ with a unique name.
You can pass either a full Git URL (https/ssh) or a saved project name (see: agent-manager projects add).
If no target is provided, an interactive picker is shown.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			gitURL  string
			project *projects.Project
			err     error
		)

		if len(args) == 0 {
			gitURL, project, err = resolveInteractiveCloneInputs(cmd)
			if err != nil {
				return err
			}
		} else {
			target := args[0]

			// Resolve target to a Git URL and get project if it exists
			gitURL, project, err = resolveTarget(target)
			if err != nil {
				return err
			}

		}

		// Create workspace manager
		wsManager, err := workspace.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create workspace manager: %w", err)
		}

		// Create a new workspace
		ws, err := wsManager.Create(gitURL)
		if err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		fmt.Printf("Created workspace: %s\n", ws.Path)

		// Clone the repository
		gitClient := git.NewClient()
		if err := gitClient.Clone(gitURL, ws.Path, cloneBranch); err != nil {
			wsManager.Remove(ws.ID)
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		fmt.Printf("Successfully cloned %s\n", gitURL)

		// Run setup commands if project has them and --skip-setup not set.
		if !skipSetup && project != nil && len(project.Commands) > 0 {
			failures := runSetupCommands(ws.Path, project.Commands, os.Stdout, os.Stderr)
			if len(failures) > 0 {
				fmt.Printf("Warning: %d setup command(s) failed; continuing with cloned workspace\n", len(failures))
			}
		}

		// Save session to history
		sessionStore, err := session.NewStore()
		if err == nil {
			projectName := ""
			if project != nil {
				projectName = project.Name
			}
			newSession := session.Session{
				ID:        ws.ID,
				GitURL:    gitURL,
				Project:   projectName,
				Tool:      aiCommand,
				Workspace: ws.Path,
			}
			if err := sessionStore.Add(newSession); err != nil {
				fmt.Printf("Warning: failed to save session: %v\n", err)
			}
		}

		// Launch AI tool, preferring a dedicated tmux session when invoked from tmux.
		launchedInTmux, err := launchCloneInTmux(gitURL, ws.ID, ws.Path, aiCommand, project)
		if err != nil {
			fmt.Printf("Warning: failed to open tmux workspace: %v\n", err)
		}
		if !launchedInTmux {
			aiClient := ai.NewClient(aiCommand)
			if err := aiClient.Launch(ws.Path); err != nil {
				fmt.Printf("Warning: failed to launch AI tool: %v\n", err)
				fmt.Printf("You can manually navigate to: %s\n", ws.Path)
			}
		}

		fmt.Printf("Workspace preserved at: %s\n", ws.Path)

		return nil
	},
}

type setupCommandFailure struct {
	command string
	err     error
}

func runSetupCommands(workspacePath string, commands []string, stdout, stderr io.Writer) []setupCommandFailure {
	fmt.Fprintln(stdout, "Running setup commands...")

	failures := make([]setupCommandFailure, 0)
	for _, command := range commands {
		fmt.Fprintf(stdout, "  > %s\n", command)

		setupCmd := exec.Command("sh", "-c", command)
		setupCmd.Dir = workspacePath
		setupCmd.Stdout = stdout
		setupCmd.Stderr = stderr
		if err := setupCmd.Run(); err != nil {
			fmt.Fprintf(stderr, "Warning: setup command failed (%q): %v\n", command, err)
			failures = append(failures, setupCommandFailure{command: command, err: err})
		}
	}

	return failures
}

func resolveInteractiveCloneInputs(cmd *cobra.Command) (string, *projects.Project, error) {
	choices, selections, err := buildInteractiveRepoChoices()
	if err != nil {
		return "", nil, err
	}

	selectedRepo, err := ui.SelectChoice("Choose repository", choices)
	if err != nil {
		return "", nil, fmt.Errorf("repository selection cancelled")
	}

	var (
		gitURL     string
		project    *projects.Project
		shouldSave bool
	)

	if selectedRepo == cloneCustomRepoChoice {
		customURL, err := ui.PromptText("Enter git URL", "")
		if err != nil {
			return "", nil, fmt.Errorf("git URL input cancelled")
		}

		gitURL = strings.TrimSpace(customURL)
		if gitURL == "" {
			return "", nil, fmt.Errorf("git URL cannot be empty")
		}
		shouldSave = true
	} else {
		selection, ok := selections[selectedRepo]
		if !ok {
			return "", nil, fmt.Errorf("invalid repository selection")
		}

		gitURL = selection.gitURL
		project = selection.project
		shouldSave = !selection.saved
	}

	if !cmd.Flags().Changed("branch") {
		selectedBranch, err := promptCloneBranchSelection(gitURL)
		if err != nil {
			return "", nil, err
		}
		cloneBranch = selectedBranch
	}

	if shouldSave {
		savedProject, err := maybeSaveProject(gitURL)
		if err != nil {
			return "", nil, err
		}
		if savedProject != nil {
			project = savedProject
		}
	}

	return gitURL, project, nil
}

func buildInteractiveRepoChoices() ([]ui.Choice, map[string]repoSelection, error) {
	choices := []ui.Choice{}
	selections := make(map[string]repoSelection)
	seenURLs := make(map[string]struct{})

	projectStore, err := projects.NewStore()
	if err != nil {
		return nil, nil, err
	}

	savedProjects, err := projectStore.List()
	if err != nil {
		return nil, nil, err
	}

	for _, p := range savedProjects {
		key := "project:" + p.Name
		choices = append(choices, ui.Choice{
			Label: fmt.Sprintf("%s -> %s", p.Name, p.URL),
			Value: key,
		})

		project := p
		selections[key] = repoSelection{
			gitURL:  p.URL,
			project: &project,
			saved:   true,
		}
		seenURLs[p.URL] = struct{}{}
	}

	sessionStore, err := session.NewStore()
	if err == nil {
		sessions, err := sessionStore.List()
		if err == nil {
			recentIndex := 0
			for _, s := range sessions {
				url := strings.TrimSpace(s.GitURL)
				if url == "" {
					continue
				}
				if _, exists := seenURLs[url]; exists {
					continue
				}

				label := fmt.Sprintf("%s (recent)", url)
				if s.Project != "" {
					label = fmt.Sprintf("%s -> %s (recent)", s.Project, url)
				}

				key := fmt.Sprintf("recent:%d", recentIndex)
				recentIndex++

				choices = append(choices, ui.Choice{Label: label, Value: key})
				selections[key] = repoSelection{gitURL: url, saved: false}
				seenURLs[url] = struct{}{}
			}
		}
	}

	choices = append(choices, ui.Choice{
		Label: "Enter custom git URL",
		Value: cloneCustomRepoChoice,
	})

	return choices, selections, nil
}

func promptCloneBranchSelection(gitURL string) (string, error) {
	gitClient := git.NewClient()

	defaultBranch, err := gitClient.GetRemoteDefaultBranch(gitURL)
	if err != nil {
		defaultBranch = ""
	}

	branches, err := gitClient.ListRemoteBranches(gitURL)
	if err != nil {
		fmt.Printf("Warning: failed to list remote branches: %v\n", err)
		branches = nil
	}

	choices := make([]ui.Choice, 0, len(branches)+2)
	defaultLabel := "Default branch"
	if defaultBranch != "" {
		defaultLabel = fmt.Sprintf("Default branch (%s)", defaultBranch)
	}
	choices = append(choices, ui.Choice{Label: defaultLabel, Value: cloneDefaultBranchChoice})

	for _, branch := range branches {
		if branch == defaultBranch {
			continue
		}
		choices = append(choices, ui.Choice{Label: branch, Value: branch})
	}

	choices = append(choices, ui.Choice{
		Label: "Enter branch name manually",
		Value: cloneManualBranchChoice,
	})

	selectedBranch, err := ui.SelectChoice("Choose branch", choices)
	if err != nil {
		return "", fmt.Errorf("branch selection cancelled")
	}

	switch selectedBranch {
	case cloneDefaultBranchChoice:
		return "", nil
	case cloneManualBranchChoice:
		manualBranch, err := ui.PromptText("Enter branch name (leave empty for default)", "")
		if err != nil {
			return "", fmt.Errorf("branch input cancelled")
		}
		return strings.TrimSpace(manualBranch), nil
	default:
		return selectedBranch, nil
	}
}

func maybeSaveProject(gitURL string) (*projects.Project, error) {
	saveProject, err := ui.ConfirmChoice("Save this repository for future use?", true)
	if err != nil {
		return nil, fmt.Errorf("save-project selection cancelled")
	}
	if !saveProject {
		return nil, nil
	}

	defaultName := git.ExtractRepoName(gitURL)
	projectName, err := ui.PromptText("Project name", defaultName)
	if err != nil {
		return nil, fmt.Errorf("project name input cancelled")
	}

	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	store, err := projects.NewStore()
	if err != nil {
		return nil, err
	}

	existingProject, exists, err := store.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	if exists {
		if existingProject.URL != gitURL {
			overwrite, err := ui.ConfirmChoice(fmt.Sprintf("Project %q already exists. Overwrite URL?", projectName), false)
			if err != nil {
				return nil, fmt.Errorf("overwrite selection cancelled")
			}
			if !overwrite {
				fmt.Printf("Skipped saving project %q\n", projectName)
				return nil, nil
			}

			if err := store.Set(projectName, gitURL); err != nil {
				return nil, err
			}
			fmt.Printf("Project %q updated: %s\n", projectName, gitURL)
		} else {
			fmt.Printf("Project %q already saved\n", projectName)
		}
	} else {
		if err := store.Add(projectName, gitURL); err != nil {
			return nil, err
		}
		fmt.Printf("Project %q added: %s\n", projectName, gitURL)
	}

	p, ok, err := store.GetProject(projectName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &p, nil
}

func resolveTarget(target string) (string, *projects.Project, error) {
	isURL := strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "https://") ||
		strings.Contains(target, ":")

	if isURL {
		return target, nil, nil
	}

	store, err := projects.NewStore()
	if err != nil {
		return "", nil, err
	}
	p, ok, err := store.GetProject(target)
	if err != nil {
		return "", nil, err
	}
	if ok {
		return p.URL, &p, nil
	}
	return "", nil, fmt.Errorf("unknown project %q. Add it with: agent-manager projects add %s <git-url>", target, target)
}

func resolveGitURL(target string) (string, error) {
	url, _, err := resolveTarget(target)
	return url, err
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	// Add custom completion for the clone argument
	cloneCmd.ValidArgsFunction = CustomCompletion

	cloneCmd.Flags().StringVar(&aiCommand, "cmd", "opencode {workspace}", "Command to run in the workspace")
	cloneCmd.Flags().StringVarP(&cloneBranch, "branch", "b", "", "Branch to check out after clone")
	cloneCmd.Flags().BoolVar(&skipSetup, "skip-setup", false, "Skip project setup commands")
	cloneCmd.Flags().BoolVar(&tmuxDispatched, "tmux-dispatched", false, "Internal flag for tmux dispatch")
	_ = cloneCmd.Flags().MarkHidden("tmux-dispatched")
}
