package buildinfo

// Version is the CLI version. Set at build time with:
//
//	go build -ldflags "-X github.com/christianwijasa/jarvis/internal/buildinfo.Version=v0.1.0"
//
// Unset builds report "dev".
var Version = "dev"
