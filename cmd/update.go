package cmd

import (
	"fmt"

	"github.com/christianwijasabfi/atlas/internal/buildinfo"
	"github.com/christianwijasabfi/atlas/internal/update"
	"github.com/spf13/cobra"
)

var applyUpdate = func(opts update.Options) (update.Result, error) {
	var c update.Client
	return c.Apply(opts)
}

func newUpdateCmd() *cobra.Command {
	var asJSON, check bool
	var repo string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Replace this binary with the latest tagged GitHub Release",
		Long: `Download the latest git tag from GitHub Releases (never main) and
replace the running atlas binary.

Needs gh on PATH. Asset names are atlas-<os>-<arch>. Checksums are
verified when SHA256SUMS is attached to the release.

--check prints current vs latest and downloads nothing.
This command does not prompt. Progress goes to stderr.

On a human TTY, other commands may offer the same install (zsh-style
[Y/n]). That nag is skipped without a TTY, with --json, for help
--agent, when the binary is "dev", and when ATLAS_DISABLE_AUTO_UPDATE
is set. ATLAS_UPDATE_DAYS (default 13) is the check interval.
ATLAS_DISABLE_UPDATE_PROMPT applies without asking.`,
		Example: `  atlas update
  atlas update --check
  atlas update --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := update.Options{
				Repo:     repo,
				Check:    check,
				Current:  buildinfo.Version,
				Progress: cmd.ErrOrStderr(),
			}
			res, err := applyUpdate(opts)
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), updateJSON{
					SchemaVersion: SchemaVersion,
					Kind:          "update",
					Current:       res.Current,
					Latest:        res.Latest,
					Updated:       res.Updated,
					Path:          res.Path,
					Check:         res.Check,
				})
			}
			if check {
				fmt.Fprintf(cmd.OutOrStdout(), "current: %s\nlatest:  %s\npath:    %s\n", res.Current, res.Latest, res.Path)
				return nil
			}
			if res.Updated {
				fmt.Fprintf(cmd.OutOrStdout(), "installed %s to %s\n", res.Latest, res.Path)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "already %s\n", res.Latest)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	cmd.Flags().BoolVar(&check, "check", false, "print current vs latest tag; do not download")
	cmd.Flags().StringVar(&repo, "repo", update.DefaultRepo, "GitHub owner/name of the atlas CLI")
	return cmd
}

type updateJSON struct {
	SchemaVersion string `json:"schema_version"`
	Kind          string `json:"kind"`
	Current       string `json:"current"`
	Latest        string `json:"latest"`
	Updated       bool   `json:"updated"`
	Path          string `json:"path"`
	Check         bool   `json:"check,omitempty"`
}
