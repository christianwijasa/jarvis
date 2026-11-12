package cmd

import (
	"fmt"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/plan"
	"github.com/spf13/cobra"
)

func newPlanCmd() *cobra.Command {
	var asJSON, dryRun bool
	var repos []string
	var tier, from string
	cmd := &cobra.Command{
		Use:   "plan <KEY>",
		Short: "Find or scaffold a TRD in the history repo",
		Long: `Find or scaffold Technical Requirements Documents for a work-item key
under history.path/docs/plans/<KEY>/.

KEY is tracker-agnostic (Jira key, Trello card id, or any path-safe id).
This command does not call a model or a tracker API. It does not prompt.

If the active project (or jarvis.yaml) has a tracker URL, MCP id, or token,
JSON source is "tracker" and agents should fetch the ticket via the host MCP.
Otherwise source is "user" and agents use pasted markdown or --from.

Existing TRD files are never overwritten.`,
		Example: `  jarvis plan TEAM-123
  jarvis plan TEAM-123 --repo api --tier A --json
  jarvis plan N-1 --from ./requirement.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			cfgPath, err := config.Find("")
			if err != nil {
				return fmt.Errorf("jarvis.yaml not found; run jarvis init")
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			history, err := plan.HistoryRoot(cfgPath, cfg)
			if err != nil {
				return err
			}
			tracker, tokenSet := plan.ResolveTracker(cfg)
			tierVal := strings.TrimSpace(tier)
			if tierVal == "" {
				tierVal = strings.TrimSpace(cfg.Tiers.Default)
			}
			res, err := plan.Run(plan.Request{
				Key:         key,
				HistoryRoot: history,
				YAMLRepos:   plan.RepoKeys(cfg.Repos),
				Selected:    repos,
				Tier:        tierVal,
				Tracker:     tracker,
				TokenSet:    tokenSet,
				FromPath:    from,
				DryRun:      dryRun,
			})
			if err != nil {
				return err
			}
			if asJSON {
				repoList := res.Repos
				if repoList == nil {
					repoList = []string{}
				}
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version":    SchemaVersion,
					"kind":              "plan",
					"issue":             key,
					"source":            res.Source,
					"tracker":           res.Tracker.Kind,
					"tracker_site":      res.Tracker.Site,
					"tracker_mcp":       res.Tracker.MCP,
					"tracker_token_set": res.TokenSet,
					"action":            res.Action,
					"history_path":      history,
					"plan_dir":          res.Dir,
					"tier":              res.Tier,
					"repos":             repoList,
					"files":             res.Files,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s  source=%s  dir=%s\n", res.Action, key, res.Source, res.Dir)
			for _, f := range res.Files {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", f)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Load grill-with-docs and trd-pr-planning before filling. Do not read project secrets.")
			if res.Source == plan.SourceTracker && res.Tracker.MCP != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Fetch the work item with host MCP %s.\n", res.Tracker.MCP)
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	cmd.Flags().StringArrayVar(&repos, "repo", nil, "repo name from jarvis.yaml repos (repeatable)")
	cmd.Flags().StringVar(&tier, "tier", "", "story tier A, B, or C (default from jarvis.yaml)")
	cmd.Flags().StringVar(&from, "from", "", "copy a requirement file to REQUIREMENT.md")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print paths without writing")
	return cmd
}
