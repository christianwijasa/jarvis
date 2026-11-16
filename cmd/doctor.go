package cmd

import (
	"fmt"

	"github.com/christianwijasa/jarvis/internal/doctor"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check tools, jarvis.yaml, skills, and history path for the active role",
		Long: `Check that required tools are on PATH, jarvis.yaml exists, JARVIS_HOME
is resolvable, and installed skill links still point at the global install.

The checks follow the role of the active project. Engineer (the default)
needs git, gh, and graphify, and checks project repo paths and graph
stamps. Product needs git, the product skills, and a history.path git
checkout; gh, a history remote, and a tracker are warnings only.

If the active project has no role, a human TTY is asked once and the
answer is saved. Without a TTY, or with --json, the role row is a warning
with the fix command. Exit 1 if any check fails. Pass --json for a
schema_version object with role, role_source, and checks[].ok.`,
		Example: `  jarvis doctor
  jarvis doctor --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, _, _ := loadProjectConfig()
			if _, err := ensureActiveRole(cmd, cfg); err != nil {
				return err
			}
			rep := doctor.Collect("")
			if asJSON {
				if err := writeJSON(cmd.OutOrStdout(), doctorJSON{
					SchemaVersion: SchemaVersion,
					Kind:          "doctor",
					OK:            rep.OK,
					Role:          rep.Role,
					RoleSource:    rep.RoleSource,
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
	Role          string        `json:"role"`
	RoleSource    string        `json:"role_source"`
	Checks        []doctor.Item `json:"checks"`
}
