package buildinfo

// Version is the CLI version. Set at build time with:
//
//	go build -ldflags "-X github.com/christianwijasabfi/atlas/internal/buildinfo.Version=v0.1.0"
//
// Unset builds report "dev".
var Version = "dev"
