package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/christianwijasabfi/atlas/internal/buildinfo"
	"github.com/christianwijasabfi/atlas/internal/update"
)

func TestShouldNagSkipsAgentsAndDev(t *testing.T) {
	root := NewRootCmd()
	root.SetIn(bytes.NewReader(nil))
	updateCmd, _, err := root.Find([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	if shouldNag(updateCmd) {
		t.Fatal("update must not nag")
	}
	agentHelp, _, err := root.Find([]string{"agent-help"})
	if err != nil {
		t.Fatal(err)
	}
	if shouldNag(agentHelp) {
		t.Fatal("agent-help must not nag")
	}
	if shouldNag(root) {
		t.Fatal("dev / non-TTY root must not nag")
	}
}

func TestNagPromptsAndApplies(t *testing.T) {
	restore := stubNag(t, "v0.1.0", "v0.2.0")
	defer restore()

	applied := false
	execed := false
	nagApply = func() (update.Result, error) {
		applied = true
		return update.Result{Current: "v0.1.0", Latest: "v0.2.0", Updated: true}, nil
	}
	nagExecSelf = func() error {
		execed = true
		return nil
	}

	out, err := executeNag(t, strings.NewReader("\n"), "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Would you like to update?") {
		t.Fatalf("missing prompt:\n%s", out)
	}
	if !applied || !execed {
		t.Fatalf("applied=%v execed=%v", applied, execed)
	}
}

func TestNagDeclineDoesNotApply(t *testing.T) {
	restore := stubNag(t, "v0.1.0", "v0.2.0")
	defer restore()

	nagApply = func() (update.Result, error) {
		t.Fatal("should not apply")
		return update.Result{}, nil
	}

	out, err := executeNag(t, strings.NewReader("n\n"), "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Skipping") {
		t.Fatalf("missing skip:\n%s", out)
	}
}

func TestNagSilentWhenDisablePrompt(t *testing.T) {
	restore := stubNag(t, "v0.1.0", "v0.2.0")
	defer restore()
	t.Setenv("ATLAS_DISABLE_UPDATE_PROMPT", "1")

	applied := false
	nagApply = func() (update.Result, error) {
		applied = true
		return update.Result{Current: "v0.1.0", Latest: "v0.2.0", Updated: true}, nil
	}
	nagExecSelf = func() error { return nil }

	out, err := executeNag(t, strings.NewReader(""), "version")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Would you like to update?") {
		t.Fatalf("prompt should be off:\n%s", out)
	}
	if !applied {
		t.Fatal("expected auto apply")
	}
}

func TestNagRespectsInterval(t *testing.T) {
	restore := stubNag(t, "v0.1.0", "v0.2.0")
	defer restore()

	home := os.Getenv("ATLAS_HOME")
	if err := update.WriteStamp(update.StampPath(home), nagNow()); err != nil {
		t.Fatal(err)
	}
	checked := false
	nagCheckLatest = func() (update.Result, error) {
		checked = true
		return update.Result{}, nil
	}

	if _, err := executeNag(t, strings.NewReader(""), "version"); err != nil {
		t.Fatal(err)
	}
	if checked {
		t.Fatal("fresh stamp must skip network")
	}
}

func TestNagSkipsJSON(t *testing.T) {
	restore := stubNag(t, "v0.1.0", "v0.2.0")
	defer restore()
	nagCheckLatest = func() (update.Result, error) {
		t.Fatal("json must not check")
		return update.Result{}, nil
	}
	if _, err := executeNag(t, strings.NewReader(""), "version", "--json"); err != nil {
		t.Fatal(err)
	}
}

func stubNag(t *testing.T, current, latest string) func() {
	t.Helper()
	t.Setenv("CI", "")
	t.Setenv("ATLAS_DISABLE_AUTO_UPDATE", "")
	t.Setenv("ATLAS_DISABLE_UPDATE_PROMPT", "")
	t.Setenv("ATLAS_UPDATE_DAYS", "13")
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)

	oldVer := buildinfo.Version
	buildinfo.Version = current
	oldTerm := terminalCheck
	terminalCheck = func(io.Reader) bool { return true }
	oldNow := nagNow
	nagNow = func() time.Time { return time.Unix(1_700_000_000, 0) }
	oldCheck := nagCheckLatest
	nagCheckLatest = func() (update.Result, error) {
		return update.Result{Current: current, Latest: latest}, nil
	}
	oldApply := nagApply
	oldExec := nagExecSelf
	nagApply = func() (update.Result, error) {
		t.Fatal("unexpected apply")
		return update.Result{}, nil
	}
	nagExecSelf = func() error {
		t.Fatal("unexpected exec")
		return nil
	}

	return func() {
		buildinfo.Version = oldVer
		terminalCheck = oldTerm
		nagNow = oldNow
		nagCheckLatest = oldCheck
		nagApply = oldApply
		nagExecSelf = oldExec
	}
}

func executeNag(t *testing.T, stdin io.Reader, args ...string) (string, error) {
	t.Helper()
	return executeIn(t, stdin, args...)
}

func TestStampPathUsesAtlasHome(t *testing.T) {
	dir := t.TempDir()
	if !strings.HasPrefix(update.StampPath(dir), dir) {
		t.Fatal(update.StampPath(dir))
	}
	if _, err := os.Stat(filepath.Dir(update.StampPath(dir))); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
