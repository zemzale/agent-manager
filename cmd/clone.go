package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/workspace"
)

var (
	aiTool        string
	keepWorkspace bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <git-url>",
	Short: "Clone a repository and launch an AI tool",
	Long: `Clone a Git repository into a temporary workspace and launch an AI tool for development.
The workspace will be created in ~/.agent-manager/workspaces/ with a unique name.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gitURL := args[0]

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

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&aiTool, "tool", "opencode", "AI tool to launch (opencode, cursor, etc.)")
	cloneCmd.Flags().BoolVar(&keepWorkspace, "keep", false, "Keep workspace after AI tool exits")
}
