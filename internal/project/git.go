package project

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GitRun runs git. Tests replace this.
var GitRun = func(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return bytes.TrimSpace(out), nil
}

type CloneResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // cloned, skipped, failed
	Path   string `json:"path,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func Clone(p *Project, parentDir string, depth int) ([]CloneResult, error) {
	var results []CloneResult
	changed := false
	for i, r := range p.Repos {
		res := CloneResult{Name: r.Name}
		dest := ExpandPath(r.Path)
		if dest == "" {
			if parentDir == "" {
				wd, err := os.Getwd()
				if err != nil {
					res.Status = "failed"
					res.Detail = "no path and no --dir; " + err.Error()
					results = append(results, res)
					continue
				}
				parentDir = wd
			}
			dest = filepath.Join(ExpandPath(parentDir), r.Name)
		}
		res.Path = dest
		if st, err := os.Stat(dest); err == nil {
			res.Status = "skipped"
			if st.IsDir() {
				res.Detail = "path exists"
			} else {
				res.Detail = "path exists (not a directory)"
			}
			if r.Path == "" {
				p.Repos[i].Path = dest
				changed = true
			}
			results = append(results, res)
			continue
		}
		if strings.TrimSpace(r.URL) == "" {
			res.Status = "skipped"
			res.Detail = "no url"
			results = append(results, res)
			continue
		}
		args := []string{"clone"}
		if depth > 0 {
			args = append(args, "--depth", strconv.Itoa(depth))
		}
		if r.DefaultBranch != "" {
			args = append(args, "--branch", r.DefaultBranch)
		}
		args = append(args, r.URL, dest)
		if _, err := GitRun("", args...); err != nil {
			res.Status = "failed"
			res.Detail = err.Error()
			results = append(results, res)
			continue
		}
		res.Status = "cloned"
		p.Repos[i].Path = dest
		changed = true
		results = append(results, res)
	}
	if changed {
		if err := Save(p); err != nil {
			return results, err
		}
	}
	return results, nil
}

type RepoStatus struct {
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	Exists   bool   `json:"exists"`
	GitRepo  bool   `json:"git_repo"`
	Branch   string `json:"branch,omitempty"`
	Dirty    bool   `json:"dirty,omitempty"`
	Ahead    int    `json:"ahead,omitempty"`
	Behind   int    `json:"behind,omitempty"`
	HEADTime *int64 `json:"head_unix,omitempty"`
	Graph    string `json:"graph,omitempty"` // fresh, stale, missing
	Detail   string `json:"detail,omitempty"`
}

func Status(p *Project, fetch bool) ([]RepoStatus, error) {
	graphPath, err := p.GraphJSON()
	if err != nil {
		return nil, err
	}
	var graphTime time.Time
	graphExists := false
	if st, err := os.Stat(graphPath); err == nil {
		graphExists = true
		graphTime = st.ModTime()
	}

	var out []RepoStatus
	for _, r := range p.Repos {
		s := RepoStatus{Name: r.Name, Path: ExpandPath(r.Path)}
		if s.Path == "" {
			s.Detail = "no local path"
			s.Graph = graphLabel(graphExists, false)
			out = append(out, s)
			continue
		}
		st, err := os.Stat(s.Path)
		if err != nil || !st.IsDir() {
			s.Exists = false
			s.Detail = "path missing"
			s.Graph = graphLabel(graphExists, false)
			out = append(out, s)
			continue
		}
		s.Exists = true
		if _, err := os.Stat(filepath.Join(s.Path, ".git")); err != nil {
			s.GitRepo = false
			s.Detail = "not a git repo"
			s.Graph = graphLabel(graphExists, false)
			out = append(out, s)
			continue
		}
		s.GitRepo = true
		if fetch {
			_, _ = GitRun(s.Path, "fetch", "--quiet")
		}
		if b, err := GitRun(s.Path, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
			s.Branch = string(b)
		}
		if b, err := GitRun(s.Path, "status", "--porcelain"); err == nil {
			s.Dirty = len(b) > 0
		}
		if b, err := GitRun(s.Path, "rev-list", "--left-right", "--count", "@{upstream}...HEAD"); err == nil {
			parts := strings.Fields(string(b))
			if len(parts) == 2 {
				s.Behind, _ = strconv.Atoi(parts[0])
				s.Ahead, _ = strconv.Atoi(parts[1])
			}
		}
		if b, err := GitRun(s.Path, "log", "-1", "--format=%ct"); err == nil {
			if n, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64); err == nil {
				s.HEADTime = &n
				if !graphExists {
					s.Graph = "missing"
				} else if time.Unix(n, 0).After(graphTime) {
					s.Graph = "stale"
				} else {
					s.Graph = "fresh"
				}
			}
		}
		if s.Graph == "" {
			s.Graph = graphLabel(graphExists, false)
		}
		out = append(out, s)
	}
	return out, nil
}

func graphLabel(exists, stale bool) string {
	if !exists {
		return "missing"
	}
	if stale {
		return "stale"
	}
	return "fresh"
}
