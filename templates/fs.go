package templates

import "embed"

//go:embed TRD-Architecture.md TRD-Implementation.md TRD-Testing.md
var FS embed.FS
