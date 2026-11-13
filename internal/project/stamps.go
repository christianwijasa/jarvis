package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	StampSchema   = "jarvis.graph.stamp/v1"
	StampFileName = "stamps.json"
)

type GraphStamps struct {
	SchemaVersion string               `json:"schema_version"`
	BuiltAt       string               `json:"built_at"`
	Repos         map[string]RepoStamp `json:"repos"`
}

type RepoStamp struct {
	Commit string `json:"commit"`
	Ref    string `json:"ref,omitempty"`
}

type RepoFreshness struct {
	Name         string `json:"name"`
	GraphCommit  string `json:"graph_commit,omitempty"`
	OriginCommit string `json:"origin_commit,omitempty"`
	Ref          string `json:"ref,omitempty"`
	Graph        string `json:"graph"` // fresh, stale, missing, unknown
	Detail       string `json:"detail,omitempty"`
}

func (p *Project) StampsPath() (string, error) {
	d, err := p.GraphDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, StampFileName), nil
}

func ReadStamps(p *Project) (GraphStamps, error) {
	var out GraphStamps
	path, err := p.StampsPath()
	if err != nil {
		return out, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			out.Repos = map[string]RepoStamp{}
			return out, nil
		}
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("stamps.json: %w", err)
	}
	if out.Repos == nil {
		out.Repos = map[string]RepoStamp{}
	}
	return out, nil
}

func WriteStamps(p *Project, stamps GraphStamps) error {
	path, err := p.StampsPath()
	if err != nil {
		return err
	}
	if stamps.SchemaVersion == "" {
		stamps.SchemaVersion = StampSchema
	}
	if stamps.BuiltAt == "" {
		stamps.BuiltAt = time.Now().UTC().Format(time.RFC3339)
	}
	if stamps.Repos == nil {
		stamps.Repos = map[string]RepoStamp{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(stamps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func headSHA(dir string) (string, error) {
	b, err := GitRun(dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func originDefault(dir, yamlBranch string) (ref, sha string, err error) {
	var candidates []string
	if b := strings.TrimSpace(yamlBranch); b != "" {
		b = strings.TrimPrefix(b, "origin/")
		candidates = append(candidates, "origin/"+b)
	}
	if b, e := GitRun(dir, "symbolic-ref", "refs/remotes/origin/HEAD"); e == nil {
		s := strings.TrimSpace(string(b))
		s = strings.TrimPrefix(s, "refs/remotes/")
		if s != "" {
			candidates = append(candidates, s)
		}
	}
	candidates = append(candidates, "origin/main", "origin/master")
	seen := map[string]bool{}
	for _, c := range candidates {
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out, e := GitRun(dir, "rev-parse", "--verify", c)
		if e != nil {
			continue
		}
		return c, string(out), nil
	}
	return "", "", fmt.Errorf("no origin default branch")
}

func stampRepo(dir, yamlBranch string) (RepoStamp, error) {
	commit, err := headSHA(dir)
	if err != nil {
		return RepoStamp{}, err
	}
	st := RepoStamp{Commit: commit}
	if ref, _, e := originDefault(dir, yamlBranch); e == nil {
		st.Ref = ref
	}
	return st, nil
}

// CheckStale compares stamps.json to origin/<default-branch>.
// If fetch is true, each git checkout is fetched first.
func CheckStale(p *Project, fetch bool) ([]RepoFreshness, error) {
	stamps, err := ReadStamps(p)
	if err != nil {
		return nil, err
	}
	graphPath, err := p.GraphJSON()
	if err != nil {
		return nil, err
	}
	_, graphStatErr := os.Stat(graphPath)
	graphExists := graphStatErr == nil

	var out []RepoFreshness
	for _, r := range p.Repos {
		f := RepoFreshness{Name: r.Name}
		if st, ok := stamps.Repos[r.Name]; ok {
			f.GraphCommit = st.Commit
			if st.Ref != "" {
				f.Ref = st.Ref
			}
		}
		abs := p.AbsRepoPath(r)
		if abs == "" {
			f.Graph = graphExistsLabel(graphExists)
			f.Detail = "no local path"
			out = append(out, f)
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			f.Graph = graphExistsLabel(graphExists)
			f.Detail = "path missing"
			out = append(out, f)
			continue
		}
		if _, err := os.Stat(filepath.Join(abs, ".git")); err != nil {
			f.Graph = graphExistsLabel(graphExists)
			f.Detail = "not a git repo"
			out = append(out, f)
			continue
		}
		if fetch {
			_, _ = GitRun(abs, "fetch", "--quiet")
		}
		ref, originSHA, err := originDefault(abs, r.DefaultBranch)
		if err != nil {
			f.Graph = "unknown"
			f.Detail = err.Error()
			out = append(out, f)
			continue
		}
		f.Ref = ref
		f.OriginCommit = originSHA
		if !graphExists {
			f.Graph = "missing"
			out = append(out, f)
			continue
		}
		if f.GraphCommit == "" || f.GraphCommit != originSHA {
			f.Graph = "stale"
			out = append(out, f)
			continue
		}
		f.Graph = "fresh"
		out = append(out, f)
	}
	return out, nil
}

func graphExistsLabel(graphExists bool) string {
	if !graphExists {
		return "missing"
	}
	return "stale"
}

func needsExtract(graph string) bool {
	return graph == "stale" || graph == "missing"
}

func AllFresh(rows []RepoFreshness) bool {
	any := false
	for _, r := range rows {
		if r.Detail == "no local path" || r.Detail == "path missing" || r.Detail == "not a git repo" {
			continue
		}
		if r.Graph == "unknown" {
			continue
		}
		any = true
		if r.Graph != "fresh" {
			return false
		}
	}
	return any
}
