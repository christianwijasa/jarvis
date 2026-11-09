package cmd

import (
	"fmt"

	"github.com/christianwijasabfi/atlas/internal/doctor"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check git, gh, graphify, atlas.yaml, and skill links",
		Long: `Check that required tools are on PATH, atlas.yaml exists, ATLAS_HOME
is resolvable, and installed skill links still point at the global install.

Does not prompt. Exit 1 if any check fails. Pass --json for a schema_version
object with checks[].ok.`,
		Example: `  atlas doctor
  atlas doctor --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rep := doctor.Collect("")
			if asJSON {
				if err := writeJSON(cmd.OutOrStdout(), doctorJSON{
					SchemaVersion: SchemaVersion,
					Kind:          "doctor",
					OK:            rep.OK,
					Checks:        rep.Checks,
				}); err != nil {
					return err
				}
			} else {
				doctor.WriteText(cmd.OutOrStdout(), rep)
			}
			if !rep.OK {
				return fmt.Errorf("doctor found problems")
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

type doctorJSON struct {
	SchemaVersion string        `json:"schema_version"`
	Kind          string        `json:"kind"`
	OK            bool          `json:"ok"`
	Checks        []doctor.Item `json:"checks"`
}
