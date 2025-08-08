package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent-manager",
	Short: "A CLI tool for managing AI-assisted development workspaces",
	Long: `Agent Manager helps you clone repositories into temporary workspaces
for AI-assisted development work. It handles cloning, launching AI tools,
and cleaning up workspaces when you're done.`,
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
