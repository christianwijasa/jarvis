package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd is the jarvis entry point. Subcommands are stubs on purpose:
// the command list is still a proposal and should stay easy to change.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "jarvis",
		Short:         "Planning guidance for a software team",
		Long:          rootLong,
		Example:       rootExample,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.PersistentPreRun = func(c *cobra.Command, _ []string) {
		nagUpdateIfNeeded(c)
	}
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

const rootLong = `Jarvis is a small, tool-agnostic CLI that gives a software team
planning guidance.

Project-specific settings (Jira site, history repo, repo map, chosen
skills) live in jarvis.yaml. Technical Requirements Documents (TRDs)
are written into a separate history repo, not into this CLI repository.

AI agents on any machine should start with:

  jarvis help --agent
  jarvis help --agent --json

Those commands print a compact guide (purpose, workflow, skill paths,
non-interactive flags, exit codes). They never prompt.

The command list below is a starting proposal. Implementations are
thin so they can change without a large rewrite.`

const rootExample = `  jarvis help --agent
  jarvis init --yes --agents-md
  jarvis doctor
  jarvis skills show grilling
  jarvis plan TEAM-123
  jarvis project use shop
  jarvis project graph shop`
