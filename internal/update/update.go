package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultRepo is the GitHub owner/name atlas update queries.
const DefaultRepo = "christianwijasabfi/atlas"

// Options control a single update attempt.
type Options struct {
	Repo     string
	Check    bool
	Current  string
	Dest     string
	GOOS     string
	GOARCH   string
	Progress io.Writer
	Timeout  time.Duration
}

// Result is the outcome of Apply (also the --json payload fields).
type Result struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Updated bool   `json:"updated"`
	Path    string `json:"path"`
	Check   bool   `json:"check,omitempty"`
}

// Client shells out to gh. Tests replace Run and LookPath.
type Client struct {
	Run      func(name string, args ...string) ([]byte, error)
	LookPath func(string) (string, error)
}

func (c Client) run(timeout time.Duration, name string, args ...string) ([]byte, error) {
	if c.Run != nil {
		return c.Run(name, args...)
	}
	var cmd *exec.Cmd
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, name, args...)
	} else {
		cmd = exec.Command(name, args...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("gh %s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}

func (c Client) lookPath(name string) (string, error) {
	if c.LookPath != nil {
		return c.LookPath(name)
	}
	return exec.LookPath(name)
}

func (o Options) progress() io.Writer {
	if o.Progress != nil {
		return o.Progress
	}
	return io.Discard
}

func writeProgress(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, format, args...)
}

// Apply fetches the latest tagged GitHub Release (never main) and optionally
// replaces Dest with the matching atlas-<os>-<arch> asset.
func (c Client) Apply(opts Options) (Result, error) {
	var res Result
	res.Current = strings.TrimSpace(opts.Current)
	res.Check = opts.Check

	repo := strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = DefaultRepo
	}
	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	if goos == "windows" {
		return res, fmt.Errorf("atlas update does not support windows")
	}

	dest := strings.TrimSpace(opts.Dest)
	if dest == "" {
		exe, err := os.Executable()
		if err != nil {
			return res, fmt.Errorf("cannot resolve current binary: %w", err)
		}
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		dest = exe
	}
	res.Path = dest

	if _, err := c.lookPath("gh"); err != nil {
		return res, fmt.Errorf("gh is not on PATH; atlas doctor checks this")
	}

	prog := opts.progress()
	writeProgress(prog, "checking latest release of %s\n", repo)

	raw, err := c.run(opts.Timeout, "gh", "release", "view", "--repo", repo, "--json", "tagName,assets")
	if err != nil {
		return res, fmt.Errorf("no GitHub Releases for %s (tag a v* release first): %w", repo, err)
	}
	var rel ghRelease
	if err := json.Unmarshal(raw, &rel); err != nil {
		return res, fmt.Errorf("parse gh release view: %w", err)
	}
	res.Latest = strings.TrimSpace(rel.TagName)
	if res.Latest == "" {
		return res, fmt.Errorf("no GitHub Releases for %s (tag a v* release first)", repo)
	}

	if !opts.Check && res.Current != "dev" && res.Current == res.Latest {
		return res, nil
	}
	if opts.Check {
		return res, nil
	}

	asset := fmt.Sprintf("atlas-%s-%s", goos, goarch)
	if !rel.hasAsset(asset) {
		return res, fmt.Errorf("release %s has no asset %s", res.Latest, asset)
	}

	writeProgress(prog, "downloading %s %s\n", res.Latest, asset)
	work, err := os.MkdirTemp("", "atlas-update-")
	if err != nil {
		return res, err
	}
	defer os.RemoveAll(work)

	if _, err := c.run(opts.Timeout, "gh", "release", "download", res.Latest, "--repo", repo, "--dir", work, "--pattern", asset, "--clobber"); err != nil {
		return res, err
	}
	binPath := filepath.Join(work, asset)
	body, err := os.ReadFile(binPath)
	if err != nil {
		return res, fmt.Errorf("read downloaded %s: %w", asset, err)
	}
	if _, err := c.run(opts.Timeout, "gh", "release", "download", res.Latest, "--repo", repo, "--dir", work, "--pattern", "SHA256SUMS", "--clobber"); err == nil {
		sumBody, err := os.ReadFile(filepath.Join(work, "SHA256SUMS"))
		if err != nil {
			return res, err
		}
		want, ok := checksumFor(sumBody, asset)
		if !ok {
			return res, fmt.Errorf("SHA256SUMS has no entry for %s", asset)
		}
		sum := sha256.Sum256(body)
		got := hex.EncodeToString(sum[:])
		if !strings.EqualFold(got, want) {
			return res, fmt.Errorf("checksum mismatch for %s: got %s want %s", asset, got, want)
		}
	}

	if err := replaceBinary(dest, body); err != nil {
		return res, err
	}
	res.Updated = true
	return res, nil
}

type ghRelease struct {
	TagName string `json:"tagName"`
	Assets  []struct {
		Name string `json:"name"`
	} `json:"assets"`
}

func (r ghRelease) hasAsset(name string) bool {
	for _, a := range r.Assets {
		if a.Name == name {
			return true
		}
	}
	return false
}

func checksumFor(sums []byte, name string) (string, bool) {
	for _, line := range strings.Split(string(sums), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[len(fields)-1] == name {
			return fields[0], true
		}
	}
	return "", false
}

func replaceBinary(dest string, body []byte) error {
	mode := os.FileMode(0o755)
	if st, err := os.Stat(dest); err == nil {
		mode = st.Mode()
	}
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, "atlas.new.*")
	if err != nil {
		return permissionHint(dest, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return permissionHint(dest, err)
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return permissionHint(dest, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return permissionHint(dest, err)
	}
	return nil
}

func permissionHint(dest string, err error) error {
	if err == nil {
		return nil
	}
	if os.IsPermission(err) {
		return fmt.Errorf("cannot replace %s: %w (try: sudo atlas update)", dest, err)
	}
	return fmt.Errorf("cannot replace %s: %w", dest, err)
}
