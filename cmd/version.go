package cmd

import (
	"fmt"

	"github.com/christianwijasa/jarvis/internal/buildinfo"
	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/skills"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print CLI version, skill-pack version, and config path",
		Long: `Print the binary version (set at build time with ldflags), the
embedded skill-pack version, and the path of jarvis.yaml if one is found.

Does not prompt. Safe for agents. Pass --json for a schema_version object.`,
		Example: `  jarvis version
  jarvis version --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath := ""
			if path, err := config.Find(""); err == nil {
				cfgPath = path
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), versionJSON{
					SchemaVersion: SchemaVersion,
					Kind:          "version",
					CLIVersion:    buildinfo.Version,
					SkillPack:     skills.Version(),
					ConfigFile:    cfgPath,
				})
			}
			if cfgPath == "" {
				cfgPath = "(none found)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "CLI version: %s\n", buildinfo.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Skill pack:  %s\n", skills.Version())
			fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n", cfgPath)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

type versionJSON struct {
	SchemaVersion string `json:"schema_version"`
	Kind          string `json:"kind"`
	CLIVersion    string `json:"cli_version"`
	SkillPack     string `json:"skill_pack"`
	ConfigFile    string `json:"config_file"`
}
