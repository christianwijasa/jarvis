//go:build !unix

package update

import (
	"fmt"
	"os"
)

var (
	errEACCES error = os.ErrPermission
	errEPERM  error = os.ErrPermission
)

// Exec is not available off unix; the caller should tell the user to re-run.
func Exec(argv0 string, argv, envv []string) error {
	return fmt.Errorf("re-exec is not supported on this OS")
}
