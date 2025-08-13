package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/projects"
)

var (
	force bool
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage saved projects (name -> git URL)",
	Long:  "Add and list projects so you can reference them by name in other commands like 'clone'.",
}

var projectsAddCmd = &cobra.Command{
	Use:   "add <name> <git-url>",
	Short: "Add a project mapping",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		url := args[1]

		store, err := projects.NewStore()
		if err != nil {
			return err
		}

		if force {
			if err := store.Set(name, url); err != nil {
				return err
			}
			fmt.Printf("Project %q set to %s\n", name, url)
			return nil
		}

		if err := store.Add(name, url); err != nil {
			return err
		}
		fmt.Printf("Project %q added: %s\n", name, url)
		return nil
	},
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := projects.NewStore()
		if err != nil {
			return err
		}
		items, err := store.List()
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("No projects saved. Add one with: agent-manager projects add <name> <git-url>")
			return nil
		}
		for _, p := range items {
			fmt.Printf("%-24s %s\n", p.Name, p.URL)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsAddCmd)
	projectsCmd.AddCommand(projectsListCmd)

	projectsAddCmd.Flags().BoolVar(&force, "force", false, "Overwrite if the project already exists")
}
