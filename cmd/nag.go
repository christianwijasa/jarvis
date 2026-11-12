package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/christianwijasa/jarvis/internal/buildinfo"
	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/update"
	"github.com/spf13/cobra"
)

const nagCheckTimeout = 5 * time.Second

var (
	nagNow         = func() time.Time { return time.Now() }
	nagCheckLatest = func() (update.Result, error) {
		var c update.Client
		return c.Apply(update.Options{
			Check:    true,
			Current:  buildinfo.Version,
			Progress: io.Discard,
			Timeout:  nagCheckTimeout,
		})
	}
	nagApply = func() (update.Result, error) {
		return applyUpdate(update.Options{
			Current:  buildinfo.Version,
			Progress: os.Stderr,
		})
	}
	nagExecSelf = func() error {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		return update.Exec(exe, os.Args, os.Environ())
	}
)

func nagUpdateIfNeeded(cmd *cobra.Command) {
	if !shouldNag(cmd) {
		return
	}
	home, err := config.HomeDir(nil)
	if err != nil {
		return
	}
	stamp := update.StampPath(home)
	now := nagNow()
	last, err := update.ReadStamp(stamp)
	if err != nil {
		return
	}
	if !update.CheckDue(last, now, nagIntervalDays()) {
		return
	}

	res, err := nagCheckLatest()
	_ = update.WriteStamp(stamp, now)
	if err != nil || !update.OfferUpdate(res.Current, res.Latest) {
		return
	}

	stderr := cmd.ErrOrStderr()
	fmt.Fprintf(stderr, "[jarvis] A newer version is available: %s → %s\n", res.Current, res.Latest)

	if envTruthy("JARVIS_DISABLE_UPDATE_PROMPT") {
		runNagInstall(stderr)
		return
	}

	fmt.Fprint(stderr, "[jarvis] Would you like to update? [Y/n] ")
	in := bufio.NewReader(cmd.InOrStdin())
	line, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintln(stderr)
		return
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		runNagInstall(stderr)
	default:
		fmt.Fprintln(stderr, "[jarvis] Skipping. Run jarvis update later, or set JARVIS_DISABLE_AUTO_UPDATE=1.")
	}
}

func runNagInstall(stderr io.Writer) {
	res, err := nagApply()
	if err != nil {
		fmt.Fprintf(stderr, "[jarvis] update failed: %v\n", err)
		return
	}
	if !res.Updated {
		fmt.Fprintf(stderr, "[jarvis] already %s\n", res.Latest)
		return
	}
	if err := nagExecSelf(); err != nil {
		fmt.Fprintf(stderr, "[jarvis] installed %s; re-run jarvis to use it\n", res.Latest)
	}
}

func shouldNag(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	switch cmd.Name() {
	case "update", "help", "agent-help", "completion":
		return false
	}
	if wantsJSON(cmd) {
		return false
	}
	if f := cmd.Flags().Lookup("help"); f != nil && f.Changed {
		return false
	}
	if envTruthy("JARVIS_DISABLE_AUTO_UPDATE") || envTruthy("CI") {
		return false
	}
	if buildinfo.Version == "dev" {
		return false
	}
	return stdinIsTerminal(cmd.InOrStdin())
}

func wantsJSON(cmd *cobra.Command) bool {
	f := cmd.Flags().Lookup("json")
	if f == nil {
		return false
	}
	v, err := cmd.Flags().GetBool("json")
	return err == nil && v
}

func nagIntervalDays() int {
	raw := strings.TrimSpace(os.Getenv("JARVIS_UPDATE_DAYS"))
	if raw == "" {
		return update.DefaultCheckIntervalDays
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return update.DefaultCheckIntervalDays
	}
	return n
}

func envTruthy(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
