package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zemzale/agent-manager/internal/ai"
	"github.com/zemzale/agent-manager/internal/session"
	"github.com/zemzale/agent-manager/internal/ui"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage agent-manager sessions",
	Long:  `List, restore, and manage your agent-manager sessions and workspace history.`,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent sessions",
	Long:  `Display a list of recent agent-manager sessions with their details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.NewStore()
		if err != nil {
			return fmt.Errorf("failed to create session store: %w", err)
		}

		sessions, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPROJECT\tGIT URL\tTOOL\tCREATED")
		fmt.Fprintln(w, strings.Repeat("-", 80))

		for _, s := range sessions {
			project := s.Project
			if project == "" {
				project = "-"
			}
			created := s.CreatedAt.Format("2006-01-02 15:04")
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.ID, project, s.GitURL, s.Tool, created)
		}

		return w.Flush()
	},
}

var sessionRestoreCmd = &cobra.Command{
	Use:   "restore [<session-id>]",
	Short: "Restore a previous session",
	Long:  `Restore a previous session by launching the same AI tool in the saved workspace. If no session ID is provided, an interactive picker will be shown.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.NewStore()
		if err != nil {
			return fmt.Errorf("failed to create session store: %w", err)
		}

		var sessionID string

		if len(args) == 0 {
			// Show picker
			sessions, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}
			if len(sessions) == 0 {
				return fmt.Errorf("no sessions found")
			}

			// Convert to ui.Session type
			uiSessions := make([]ui.Session, len(sessions))
			for i, s := range sessions {
				uiSessions[i] = ui.Session{
					ID:        s.ID,
					Project:   s.Project,
					GitURL:    s.GitURL,
					Tool:      s.Tool,
					CreatedAt: s.CreatedAt,
				}
			}

			selectedID, err := ui.SelectSession(uiSessions)
			if err != nil {
				return fmt.Errorf("session selection cancelled")
			}
			sessionID = selectedID
		} else {
			sessionID = args[0]
		}

		sess, err := store.Get(sessionID)
		if err != nil {
			return err
		}

		return restoreSession(sess)
	},
}

var sessionClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all session history",
	Long:  `Remove all sessions from the history. This cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.NewStore()
		if err != nil {
			return fmt.Errorf("failed to create session store: %w", err)
		}

		if err := store.Clear(); err != nil {
			return fmt.Errorf("failed to clear sessions: %w", err)
		}

		fmt.Println("Session history cleared.")
		return nil
	},
}

func restoreSession(sess *session.Session) error {
	// Check if workspace directory exists
	if _, err := os.Stat(sess.Workspace); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory no longer exists (was cleaned up): %s", sess.Workspace)
	}

	fmt.Printf("Restoring session: %s\n", sess.ID)
	fmt.Printf("Workspace: %s\n", sess.Workspace)
	fmt.Printf("AI Tool: %s\n", sess.Tool)

	// Launch AI tool
	aiClient := ai.NewClient(sess.Tool)
	if err := aiClient.Launch(sess.Workspace); err != nil {
		return fmt.Errorf("failed to launch AI tool: %w", err)
	}

	return nil
}

// sessionCompletion provides autocomplete for session IDs
func sessionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	store, err := session.NewStore()
	if err == nil {
		sessions, err := store.List()
		if err == nil {
			for _, s := range sessions {
				if strings.HasPrefix(s.ID, toComplete) {
					completions = append(completions, s.ID)
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(sessionCmd)

	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionRestoreCmd)
	sessionCmd.AddCommand(sessionClearCmd)

	// Add completion for session restore command
	sessionRestoreCmd.ValidArgsFunction = sessionCompletion
}
