//go:build unix

package update

import "syscall"

// Exec replaces the current process. Used after a successful self-update so
// the user's command runs on the new binary.
func Exec(argv0 string, argv, envv []string) error {
	return syscall.Exec(argv0, argv, envv)
}
