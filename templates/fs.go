package templates

import "embed"

//go:embed PRD.md TRD-Architecture.md TRD-Implementation.md TRD-Testing.md
var FS embed.FS
