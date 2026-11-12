package cmd

import (
	_ "embed"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed testdata/agent-help.md
var agentHelpMarkdown string

type agentHelpJSON struct {
	SchemaVersion     string     `json:"schema_version"`
	Kind              string     `json:"kind"`
	Markdown          string     `json:"markdown"`
	Workflow          []string   `json:"workflow"`
	ExitCodes         []exitCode `json:"exit_codes"`
	NonInteractive    []string   `json:"non_interactive_flags"`
	LoadSkill         string     `json:"load_skill"`
	JarvisHomeDefault string     `json:"jarvis_home_default"`
}

type exitCode struct {
	Code    string `json:"code"`
	Meaning string `json:"meaning"`
}

func newHelpCmd() *cobra.Command {
	var agent, asJSON bool
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Print help for jarvis or a subcommand.

For AI agents on any machine, use:

  jarvis help --agent
  jarvis help --agent --json
  jarvis agent-help`,
		Example: `  jarvis help
  jarvis help init
  jarvis help --agent
  jarvis help --agent --json`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && !agent {
				return errJSONNeedsAgent
			}
			if agent {
				if len(args) > 0 {
					return errAgentHelpNoArgs
				}
				return printAgentHelp(cmd, asJSON)
			}
			root := cmd.Root()
			target := root
			if len(args) > 0 {
				found, _, err := root.Find(args)
				if err != nil {
					return err
				}
				target = found
			}
			return target.Help()
		},
	}
	cmd.Flags().BoolVar(&agent, "agent", false, "print the compact AI-agent guide")
	addJSONFlag(cmd, &asJSON)
	return cmd
}

func newAgentHelpCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "agent-help",
		Short:   "Print the compact AI-agent guide (same as jarvis help --agent)",
		Long:    "Plain-markdown (or JSON) guide for other AI agents. No prompts. Same content as `jarvis help --agent`.",
		Example: "  jarvis agent-help\n  jarvis agent-help --json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printAgentHelp(cmd, asJSON)
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

func printAgentHelp(cmd *cobra.Command, asJSON bool) error {
	md := strings.TrimSuffix(agentHelpMarkdown, "\n") + "\n"
	if !asJSON {
		_, err := cmd.OutOrStdout().Write([]byte(md))
		return err
	}
	return writeJSON(cmd.OutOrStdout(), agentHelpJSON{
		SchemaVersion:     SchemaVersion,
		Kind:              "agent_help",
		Markdown:          md,
		Workflow:          []string{"help --agent", "init", "doctor", "skills", "skills show", "project current", "project graph", "plan"},
		ExitCodes:         []exitCode{{"0", "success"}, {"1", "error"}},
		NonInteractive:    []string{"--yes", "--recommended", "--skills", "--scope", "--copy", "--local-dir", "--agents-md", "--project", "--json", "--agent", "--repo", "--update", "--repos", "--fetch", "--check", "--from", "--dry-run", "--tier", "--tracker-site", "--tracker-token", "--tracker-mcp"},
		LoadSkill:         "jarvis skills show <name>",
		JarvisHomeDefault: "~/.jarvis",
	})
}

var (
	errJSONNeedsAgent  = errStr("jarvis help --json requires --agent")
	errAgentHelpNoArgs = errStr("jarvis help --agent does not take a command name")
)

type errStr string

func (e errStr) Error() string { return string(e) }
