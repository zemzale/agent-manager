package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/projects"
	"github.com/zemzale/agent-manager/internal/session"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:
  $ source <(agent-manager completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ agent-manager completion bash > /etc/bash_completion.d/agent-manager
  # macOS:
  $ agent-manager completion bash > /usr/local/etc/bash_completion.d/agent-manager

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ agent-manager completion zsh > "${fpath[1]}/_agent-manager"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ agent-manager completion fish | source
  # To load completions for each session, execute once:
  $ agent-manager completion fish > ~/.config/fish/completions/agent-manager.fish

PowerShell:
  PS> agent-manager completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> agent-manager completion powershell > agent-manager.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

// CustomCompletion provides autocomplete for project names and session IDs
func CustomCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// Get project names for completion
	projectStore, err := projects.NewStore()
	if err == nil {
		projects, err := projectStore.List()
		if err == nil {
			for _, p := range projects {
				if strings.HasPrefix(p.Name, toComplete) {
					completions = append(completions, p.Name)
				}
			}
		}
	}

	// Get recent session project names for completion
	sessionStore, err := session.NewStore()
	if err == nil {
		recentProjects, err := sessionStore.GetRecentProjectNames()
		if err == nil {
			for _, name := range recentProjects {
				if strings.HasPrefix(name, toComplete) {
					// Avoid duplicates
					found := false
					for _, existing := range completions {
						if existing == name {
							found = true
							break
						}
					}
					if !found {
						completions = append(completions, name)
					}
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
