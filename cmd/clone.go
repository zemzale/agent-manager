package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
	"github.com/zemzale/agent-manager/internal/workspace"
)

var (
	aiTool        string
	keepWorkspace bool
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

		// Resolve target to a Git URL if a project name was provided.
		gitURL, err := resolveGitURL(target)
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
			// Clean up on failure
			wsManager.Remove(ws.ID)
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		fmt.Printf("Successfully cloned %s\n", gitURL)

		// Launch AI tool
		aiClient := ai.NewClient(aiTool)
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

func resolveGitURL(target string) (string, error) {
	// Heuristic: URL if starts with http(s) or looks like SSH (contains ':')
	isURL := strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "https://") ||
		strings.Contains(target, ":")

	if isURL {
		return target, nil
	}

	store, err := projects.NewStore()
	if err != nil {
		return "", err
	}
	if u, ok, err := store.Get(target); err != nil {
		return "", err
	} else if ok {
		return u, nil
	}
	return "", fmt.Errorf("unknown project %q. Add it with: agent-manager projects add %s <git-url>", target, target)
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&aiTool, "tool", "opencode", "AI tool to launch (opencode, cursor, etc.)")
	cloneCmd.Flags().BoolVar(&keepWorkspace, "keep", false, "Keep workspace after AI tool exits")
}
