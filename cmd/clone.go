package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
	"github.com/zemzale/agent-manager/internal/session"
	"github.com/zemzale/agent-manager/internal/workspace"
)

var (
	aiCommand     string
	keepWorkspace bool
	skipSetup     bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <git-url|project-name>",
	Short: "Clone a repository and launch an AI tool",
	Long: `Clone a Git repository into a temporary workspace and launch an AI tool for development.
The workspace will be created in ~/.agent-manager/workspaces/ with a unique name.
You can pass either a full Git URL (https/ssh) or a saved project name (see: agent-manager projects add).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		// Resolve target to a Git URL and get project if it exists
		gitURL, project, err := resolveTarget(target)
		if err != nil {
			return err
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
		if err := gitClient.Clone(gitURL, ws.Path); err != nil {
			wsManager.Remove(ws.ID)
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		fmt.Printf("Successfully cloned %s\n", gitURL)

		// Run setup commands if project has them and --skip-setup not set
		if !skipSetup && project != nil && len(project.Commands) > 0 {
			fmt.Println("Running setup commands...")
			for _, c := range project.Commands {
				fmt.Printf("  > %s\n", c)
				cmd := exec.Command("sh", "-c", c)
				cmd.Dir = ws.Path
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				if err := cmd.Run(); err != nil {
					wsManager.Remove(ws.ID)
					return fmt.Errorf("setup command failed: %w", err)
				}
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

		// Launch AI tool
		aiClient := ai.NewClient(aiCommand)
		if err := aiClient.Launch(ws.Path); err != nil {
			fmt.Printf("Warning: failed to launch AI tool: %v\n", err)
			fmt.Printf("You can manually navigate to: %s\n", ws.Path)
		}

		// Clean up workspace unless --keep is specified
		if !keepWorkspace {
			fmt.Println("Cleaning up workspace...")
			if err := wsManager.Remove(ws.ID); err != nil {
				fmt.Printf("Warning: failed to clean up workspace: %v\n", err)
			}
		} else {
			fmt.Printf("Workspace preserved at: %s\n", ws.Path)
		}

		return nil
	},
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

	cloneCmd.Flags().StringVar(&aiCommand, "cmd", "opencode .", "Command to run in the workspace")
	cloneCmd.Flags().BoolVar(&keepWorkspace, "keep", false, "Keep workspace after command exits")
	cloneCmd.Flags().BoolVar(&skipSetup, "skip-setup", false, "Skip project setup commands")
}
