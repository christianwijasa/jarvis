package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd is the atlas entry point. Subcommands are stubs on purpose:
// the command list is still a proposal and should stay easy to change.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "atlas",
		Short:         "Planning guidance for a software team",
		Long:          rootLong,
		Example:       rootExample,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(newHelpCmd())
	cmd.AddCommand(
		newVersionCmd(),
		newUpdateCmd(),
		newPlanCmd(),
		newSkillsCmd(),
		newDoctorCmd(),
		newInitCmd(),
		newProjectCmd(),
		newAgentHelpCmd(),
	)
	return cmd
}

const rootLong = `Atlas is a small, tool-agnostic CLI that gives a software team
planning guidance.

Project-specific settings (Jira site, history repo, repo map, chosen
skills) live in atlas.yaml. Technical Requirements Documents (TRDs)
are written into a separate history repo, not into this CLI repository.

AI agents on any machine should start with:

  atlas help --agent
  atlas help --agent --json

Those commands print a compact guide (purpose, workflow, skill paths,
non-interactive flags, exit codes). They never prompt.

The command list below is a starting proposal. Implementations are
thin so they can change without a large rewrite.`

const rootExample = `  atlas help --agent
  atlas init --yes --agents-md
  atlas doctor
  atlas skills show grilling
  atlas plan TEAM-123
  atlas project use shop
  atlas project graph shop`
