package cmd

import (
	"fmt"

	"github.com/christianwijasabfi/atlas/internal/config"
	"github.com/spf13/cobra"
)

func newPlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan <JIRA-KEY>",
		Short: "Scaffold a TRD into the history repo (not yet implemented)",
		Long: `Would scaffold a Technical Requirements Document for the given Jira
issue key into history.path from atlas.yaml, using skill trd-pr-planning.

This command does not call a model or the Jira API. It does not prompt.
Load the skill first with: atlas skills show trd-pr-planning`,
		Example: `  atlas plan TEAM-123
  atlas skills show trd-pr-planning`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			out := cmd.OutOrStdout()

			fmt.Fprintf(out, "Would plan %s using skill trd-pr-planning.\n", key)
			fmt.Fprintln(out, "This command does not call a model or the Jira API yet.")

			path, err := config.Find("")
			if err != nil {
				fmt.Fprintln(out, "No atlas.yaml found. History repo path is unknown until you run atlas init or add a config file.")
				fmt.Fprintln(out, "A later version will copy templates/trd.md into that history repo.")
				return nil
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			history := cfg.History.Path
			if history == "" {
				history = "(history.path is empty in atlas.yaml)"
			}
			fmt.Fprintf(out, "Using config: %s\n", path)
			fmt.Fprintf(out, "Would scaffold templates/trd.md into the history repo at %s.\n", history)
			return nil
		},
	}
}
