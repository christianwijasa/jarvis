package upstream

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// Remote looks up the latest commit for an upstream path. Tests mock this.
type Remote interface {
	LatestCommit(repoURL, path string) (sha string, err error)
}

// Result is one registry row after a read-only upstream check.
type Result struct {
	Entry   Entry
	Status  string // unchanged, changed, unpinned, original, error
	Pinned  string
	Latest  string
	Compare string
	Detail  string
}

// Check compares each entry with a pinned commit against Remote.
// It never writes, fetches a clone, or applies a patch.
func Check(reg *Registry, remote Remote) []Result {
	out := make([]Result, 0, len(reg.Entries))
	for _, e := range reg.Entries {
		out = append(out, checkOne(e, remote))
	}
	return out
}

func checkOne(e Entry, remote Remote) Result {
	r := Result{Entry: e, Pinned: displayPin(e.PinnedCommit)}
	if e.Mode == ModeOriginal || !e.HasUpstream() {
		r.Status = "original"
		r.Detail = "no upstream; Jarvis-written"
		return r
	}
	if !e.PinnedKnown() {
		r.Status = "unpinned"
		r.Detail = "pinned commit unknown; will not guess"
		return r
	}
	latest, err := remote.LatestCommit(e.UpstreamRepo, e.UpstreamPath)
	if err != nil {
		r.Status = "error"
		r.Detail = err.Error()
		return r
	}
	r.Latest = latest
	r.Compare = compareURL(e.UpstreamRepo, e.PinnedCommit, latest)
	if sameCommit(e.PinnedCommit, latest) {
		r.Status = "unchanged"
		r.Detail = "upstream path matches pin"
		return r
	}
	r.Status = "changed"
	r.Detail = "upstream path has commits after the pin; adopt only in a reviewed PR"
	return r
}

func displayPin(p string) string {
	if strings.TrimSpace(p) == "" {
		return "n/a"
	}
	return p
}

func sameCommit(a, b string) bool {
	a, b = strings.ToLower(strings.TrimSpace(a)), strings.ToLower(strings.TrimSpace(b))
	n := 7
	if len(a) < n || len(b) < n {
		return a == b
	}
	if strings.HasPrefix(a, b) || strings.HasPrefix(b, a) {
		return true
	}
	return a == b
}

func compareURL(repo, pinned, latest string) string {
	owner, name, ok := githubOwnerRepo(repo)
	if !ok {
		return ""
	}
	short := func(s string) string {
		if len(s) > 12 {
			return s[:12]
		}
		return s
	}
	return fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s", owner, name, short(pinned), short(latest))
}

func githubOwnerRepo(repoURL string) (owner, repo string, ok bool) {
	u := strings.TrimSuffix(strings.TrimSpace(repoURL), ".git")
	u = strings.TrimSuffix(u, "/")
	switch {
	case strings.HasPrefix(u, "https://github.com/"):
		u = strings.TrimPrefix(u, "https://github.com/")
	case strings.HasPrefix(u, "http://github.com/"):
		u = strings.TrimPrefix(u, "http://github.com/")
	default:
		return "", "", false
	}
	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// GHRemote uses `gh api` for a path-level latest commit, then `git ls-remote`
// if gh fails. Read-only.
type GHRemote struct {
	Run func(name string, args ...string) ([]byte, error)
}

func (r GHRemote) run(name string, args ...string) ([]byte, error) {
	if r.Run != nil {
		return r.Run(name, args...)
	}
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func (r GHRemote) LatestCommit(repoURL, path string) (string, error) {
	owner, repo, ok := githubOwnerRepo(repoURL)
	if ok {
		q := url.Values{}
		q.Set("per_page", "1")
		if strings.TrimSpace(path) != "" {
			q.Set("path", path)
		}
		endpoint := fmt.Sprintf("repos/%s/%s/commits?%s", owner, repo, q.Encode())
		out, err := r.run("gh", "api", endpoint)
		if err == nil {
			sha, parseErr := parseGHCommits(out)
			if parseErr == nil {
				return sha, nil
			}
			err = parseErr
		}
		// fall through to ls-remote
		_ = err
	}
	return r.lsRemote(repoURL)
}

func (r GHRemote) lsRemote(repoURL string) (string, error) {
	out, err := r.run("git", "ls-remote", repoURL, "HEAD")
	if err != nil {
		return "", fmt.Errorf("git ls-remote: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", fmt.Errorf("git ls-remote: empty output")
	}
	return fields[0], nil
}

func parseGHCommits(raw []byte) (string, error) {
	var rows []struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return "", fmt.Errorf("gh api: %w", err)
	}
	if len(rows) == 0 || rows[0].SHA == "" {
		return "", fmt.Errorf("gh api: no commits")
	}
	return rows[0].SHA, nil
}
