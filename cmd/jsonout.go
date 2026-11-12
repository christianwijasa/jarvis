package cmd

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
)

// SchemaVersion is included in every --json payload so agents can
// detect incompatible shapes.
const SchemaVersion = "jarvis.cli/v1"

func addJSONFlag(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, "json", false, "machine-readable JSON (schema_version "+SchemaVersion+")")
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
