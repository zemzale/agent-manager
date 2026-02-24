package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/git"
	"github.com/zemzale/agent-manager/internal/projects"
	"github.com/zemzale/agent-manager/internal/ui"
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
	Use:   "add [name] [git-url]",
	Short: "Add a project mapping (auto-detects from current git repo if no args)",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name, url string

		switch len(args) {
		case 2:
			name, url = args[0], args[1]
		default:
			var err error
			url, err = resolveURL()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				name = args[0]
			} else {
				name = git.ExtractRepoName(url)
			}
		}

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

func resolveURL() (string, error) {
	client := git.NewClient()
	remotes, err := client.ListRemotes()
	if err != nil || len(remotes) == 0 {
		return ui.PromptURL()
	}

	var remote string
	if len(remotes) == 1 {
		remote = remotes[0]
	} else {
		remote, err = ui.SelectRemote(remotes)
		if err != nil {
			return "", err
		}
	}
	return client.GetRemoteURL(remote)
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

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Manage setup commands for a project",
}

var commandsListCmd = &cobra.Command{
	Use:   "list <project>",
	Short: "List setup commands for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := projects.NewStore()
		if err != nil {
			return err
		}
		p, ok, err := store.GetProject(args[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("project %q not found", args[0])
		}
		if len(p.Commands) == 0 {
			fmt.Println("No setup commands configured.")
			return nil
		}
		for i, c := range p.Commands {
			fmt.Printf("%d: %s\n", i, c)
		}
		return nil
	},
}

var commandsAddCmd = &cobra.Command{
	Use:   "add <project> <command>",
	Short: "Add a setup command to a project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := projects.NewStore()
		if err != nil {
			return err
		}
		if err := store.AddCommand(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Added command to %q\n", args[0])
		return nil
	},
}

var commandsRemoveCmd = &cobra.Command{
	Use:   "remove <project> <index>",
	Short: "Remove a setup command by index",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid index: %w", err)
		}
		store, err := projects.NewStore()
		if err != nil {
			return err
		}
		if err := store.RemoveCommand(args[0], idx); err != nil {
			return err
		}
		fmt.Printf("Removed command %d from %q\n", idx, args[0])
		return nil
	},
}

var commandsEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in $EDITOR (falls back to nvim)",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := projects.NewStore()
		if err != nil {
			return err
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nvim"
		}
		c := exec.Command(editor, store.FilePath())
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsAddCmd)
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(commandsCmd)

	commandsCmd.AddCommand(commandsListCmd)
	commandsCmd.AddCommand(commandsAddCmd)
	commandsCmd.AddCommand(commandsRemoveCmd)
	commandsCmd.AddCommand(commandsEditCmd)

	projectsAddCmd.Flags().BoolVar(&force, "force", false, "Overwrite if the project already exists")
}
