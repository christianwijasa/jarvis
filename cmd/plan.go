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
	var tier, from, typ string
	cmd := &cobra.Command{
		Use:   "plan <KEY>",
		Short: "Find or scaffold a PRD or TRD in the history repo",
		Long: `Find or scaffold planning documents for a work-item key under
history.path/docs/plans/<KEY>/.

--type prd writes only PRD.md (Product Requirements Document). --type trd
writes the Technical Requirements Documents. The default comes from the
role of the active project: product gives prd, engineer gives trd. A
project with no role, or no active project, means engineer. --repo and
--tier are for TRDs only. With --type prd, --from goes into the PRD
"Source notes" section instead of REQUIREMENT.md.

KEY is tracker-agnostic (Jira key, Trello card id, or any path-safe id).
This command does not call a model or a tracker API. It does not prompt.

If the active project (or jarvis.yaml) has a tracker URL, MCP id, or token,
JSON source is "tracker" and agents should fetch the ticket via the host MCP.
Otherwise source is "user" and agents use pasted markdown or --from.

Existing files are never overwritten. TRD files do not block a new
PRD.md, and PRD.md does not block a TRD scaffold.`,
		Example: `  jarvis plan TEAM-123
  jarvis plan TEAM-123 --repo api --tier A --json
  jarvis plan TEAM-123 --type prd --json
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
			planType, err := plan.ParseType(typ)
			if err != nil {
				return err
			}
			rr, err := ensureActiveRole(cmd, cfg)
			if err != nil {
				return err
			}
			if planType == "" {
				planType = rr.PlanType
			}
			tracker, tokenSet := plan.ResolveTracker(cfg)
			tierVal := strings.TrimSpace(tier)
			yamlRepos := plan.RepoKeys(cfg.Repos)
			if planType == plan.TypePRD {
				if cmd.Flags().Changed("repo") {
					return fmt.Errorf("--repo is for TRDs; do not pass it with --type prd")
				}
				if cmd.Flags().Changed("tier") {
					return fmt.Errorf("--tier is for TRDs; do not pass it with --type prd")
				}
				yamlRepos = nil
			} else if tierVal == "" {
				tierVal = strings.TrimSpace(cfg.Tiers.Default)
			}
			res, err := plan.Run(plan.Request{
				Type:        planType,
				Key:         key,
				HistoryRoot: history,
				YAMLRepos:   yamlRepos,
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
					"type":              res.Type,
					"role":              rr.Role,
					"role_source":       rr.Source,
					"prd_file":          res.PRDFile,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s  source=%s  dir=%s\n", res.Action, key, res.Source, res.Dir)
			for _, f := range res.Files {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", f)
			}
			if res.Type == plan.TypePRD {
				fmt.Fprintln(cmd.OutOrStdout(), "Load prd-planning and grilling before filling. Do not read project secrets.")
			} else {
				if res.PRDFile != "" {
					fmt.Fprintln(cmd.OutOrStdout(), "PRD.md found: read it first (trd-planning).")
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Load grill-with-docs and trd-planning before filling. Do not read project secrets.")
			}
			if res.Source == plan.SourceTracker && res.Tracker.MCP != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Fetch the work item with host MCP %s.\n", res.Tracker.MCP)
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	cmd.Flags().StringArrayVar(&repos, "repo", nil, "repo name from jarvis.yaml repos (repeatable)")
	cmd.Flags().StringVar(&tier, "tier", "", "story tier A, B, or C (default from jarvis.yaml)")
	cmd.Flags().StringVar(&from, "from", "", "requirement file: REQUIREMENT.md for a TRD, Source notes for a PRD")
	cmd.Flags().StringVar(&typ, "type", "", "prd or trd (default from the active project role)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print paths without writing")
	return cmd
}
