package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update Atlas to the latest tagged release (not yet implemented)",
		Long: `When implemented, replace the current binary with the latest git tag
(for example v0.1.0). It will not follow the main branch.

No network calls in this skeleton. Does not prompt.`,
		Example: `  atlas update`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprint(cmd.OutOrStdout(), updateMessage)
			return nil
		},
	}
}

const updateMessage = `atlas update is not implemented yet.

When it is, this command will replace the current binary with the latest
tagged release. It will not follow the main branch. Releases are version
tags (for example v0.1.0), so a team can pin and roll forward on purpose.

No network calls are made in this skeleton.
`
