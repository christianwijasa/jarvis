package project

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Progress receives extract i/n and merge lines plus streamed graphify logs.
// Defaults to stderr. Tests replace this.
var Progress io.Writer = os.Stderr

func progressWriter() io.Writer {
	if Progress != nil {
		return Progress
	}
	return os.Stderr
}

func writeProgress(format string, args ...any) {
	fmt.Fprintf(progressWriter(), format, args...)
}

// GraphifyRun shells out to the graphify binary. Tests replace this.
// stdout and stderr are streamed to Progress so a hang is visible.
var GraphifyRun = func(args ...string) error {
	path, err := exec.LookPath("graphify")
	if err != nil {
		return fmt.Errorf("graphify is not on PATH; install the graphify CLI (jarvis doctor checks this)")
	}
	cmd := exec.Command(path, args...)
	out := progressWriter()
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("graphify %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

type GraphBuildResult struct {
	Project   string   `json:"project"`
	GraphJSON string   `json:"graph_json"`
	Repos     []string `json:"repos"`
	Extracts  []string `json:"extract_invocations,omitempty"`
	MergeArgs []string `json:"merge_args,omitempty"`
	Skipped   []string `json:"skipped_fresh,omitempty"`
}

// BuildGraph extracts each selected repo with `graphify extract PATH --out DIR`
// (v0.9.15) so graphify-out is NOT written inside member checkouts, then
// `graphify merge-graphs` which prefixes node ids with the repo name
// (parent.parent of each graph.json).
//
// If ifStale is true, Jarvis fetches origin, compares stamps.json to
// origin/<default-branch>, and extracts only stale or missing members.
func BuildGraph(p *Project, repoFilter []string, ifStale bool) (GraphBuildResult, error) {
	var res GraphBuildResult
	res.Project = p.Name
	graphJSON, err := p.GraphJSON()
	if err != nil {
		return res, err
	}
	res.GraphJSON = graphJSON

	selected, err := selectRepos(p, repoFilter)
	if err != nil {
		return res, err
	}
	if len(selected) == 0 {
		return res, fmt.Errorf("project %s has no repos with a local path", p.Name)
	}

	extractSet := map[string]bool{}
	if ifStale {
		rows, err := CheckStale(p, true)
		if err != nil {
			return res, err
		}
		byName := map[string]RepoFreshness{}
		for _, row := range rows {
			byName[row.Name] = row
		}
		allFresh := true
		for _, r := range selected {
			row := byName[r.Name]
			if needsExtract(row.Graph) {
				extractSet[r.Name] = true
				allFresh = false
				continue
			}
			if row.Graph == "fresh" {
				res.Skipped = append(res.Skipped, r.Name)
			}
		}
		if allFresh && graphJSONUsable(graphJSON) {
			for _, r := range selected {
				res.Repos = append(res.Repos, r.Name)
			}
			writeProgress("graph fresh; skip extract\n")
			return res, nil
		}
	} else {
		for _, r := range selected {
			extractSet[r.Name] = true
		}
	}

	stamps, err := ReadStamps(p)
	if err != nil {
		return res, err
	}
	if stamps.Repos == nil {
		stamps.Repos = map[string]RepoStamp{}
	}

	n := len(selected)
	for i, r := range selected {
		if !extractSet[r.Name] {
			continue
		}
		abs := p.AbsRepoPath(r)
		if abs == "" {
			return res, fmt.Errorf("repo %s has no local path; clone or set --path", r.Name)
		}
		if st, err := os.Stat(abs); err != nil || !st.IsDir() {
			return res, fmt.Errorf("repo %s path %s is missing", r.Name, abs)
		}
		outDir, err := p.ExtractOutDir(r.Name)
		if err != nil {
			return res, err
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return res, err
		}
		args := []string{"extract", abs, "--out", outDir, "--code-only"}
		writeProgress("extract %d/%d %s\n", i+1, n, r.Name)
		extractErr := GraphifyRun(args...)
		destGraph := filepath.Join(outDir, "graphify-out", "graph.json")
		ok := true
		if extractErr != nil || !graphJSONUsable(destGraph) {
			inRepo := filepath.Join(abs, "graphify-out", "graph.json")
			if graphJSONUsable(inRepo) {
				writeProgress("extract %d/%d %s: empty --out extract; using in-repo graphify-out\n", i+1, n, r.Name)
				if err := copyGraphJSON(inRepo, destGraph); err != nil {
					return res, fmt.Errorf("extract %s (%d/%d): copy in-repo graph: %w", r.Name, i+1, n, err)
				}
			} else {
				if extractErr != nil {
					writeProgress("extract %d/%d %s: extract failed; skipping (no in-repo graphify-out)\n", i+1, n, r.Name)
				} else {
					writeProgress("extract %d/%d %s: empty --out extract; skipping (no in-repo graphify-out)\n", i+1, n, r.Name)
				}
				ok = false
			}
		}
		if !ok {
			continue
		}
		res.Extracts = append(res.Extracts, strings.Join(args, " "))
		if st, err := stampRepo(abs, r.DefaultBranch); err == nil {
			stamps.Repos[r.Name] = st
		}
	}

	var graphFiles []string
	var mergeRepos []string
	for _, r := range selected {
		outDir, err := p.ExtractOutDir(r.Name)
		if err != nil {
			return res, err
		}
		destGraph := filepath.Join(outDir, "graphify-out", "graph.json")
		if !graphJSONUsable(destGraph) {
			continue
		}
		graphFiles = append(graphFiles, destGraph)
		mergeRepos = append(mergeRepos, r.Name)
	}
	res.Repos = mergeRepos
	if len(graphFiles) == 0 {
		return res, fmt.Errorf("project %s: no usable graphs after extract", p.Name)
	}
	if len(graphFiles) == 1 {
		if err := prefixGraphFile(graphFiles[0], graphJSON, mergeRepos[0]); err != nil {
			return res, err
		}
	} else {
		graphDir := filepath.Dir(graphJSON)
		if err := os.MkdirAll(graphDir, 0o755); err != nil {
			return res, err
		}
		args := append([]string{"merge-graphs"}, graphFiles...)
		args = append(args, "--out", graphJSON)
		writeProgress("merge %d graphs\n", len(graphFiles))
		if err := GraphifyRun(args...); err != nil {
			return res, fmt.Errorf("merge %d graphs: %w", n, err)
		}
		res.MergeArgs = args
	}
	stamps.BuiltAt = time.Now().UTC().Format(time.RFC3339)
	if err := WriteStamps(p, stamps); err != nil {
		return res, err
	}
	if err := writeGraphSidecar(p, graphJSON, mergeRepos); err != nil {
		return res, err
	}
	return res, nil
}

func selectRepos(p *Project, filter []string) ([]Repo, error) {
	if len(filter) == 0 {
		var out []Repo
		for _, r := range p.Repos {
			if ExpandPath(r.Path) != "" {
				out = append(out, r)
			}
		}
		return out, nil
	}
	var out []Repo
	for _, name := range filter {
		r, ok := p.Repo(name)
		if !ok {
			return nil, fmt.Errorf("repo %q is not in project %s", name, p.Name)
		}
		out = append(out, r)
	}
	return out, nil
}

func graphJSONUsable(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil || len(strings.TrimSpace(string(raw))) == 0 {
		return false
	}
	var data struct {
		Nodes []any `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return true
	}
	return len(data.Nodes) > 0
}

func copyGraphJSON(src, dest string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, raw, 0o644)
}

func prefixGraphFile(src, dest, tag string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, prefixNodeIDs(raw, tag), 0o644)
}

func prefixNodeIDs(raw []byte, tag string) []byte {
	if tag == "" {
		return raw
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return raw
	}
	rewrite := func(v any) any {
		s, ok := v.(string)
		if !ok || s == "" {
			return v
		}
		if strings.HasPrefix(s, tag+":") {
			return s
		}
		return tag + ":" + s
	}
	if nodes, ok := data["nodes"].([]any); ok {
		for _, n := range nodes {
			m, ok := n.(map[string]any)
			if !ok {
				continue
			}
			if id, ok := m["id"]; ok {
				m["id"] = rewrite(id)
			}
		}
	}
	for _, key := range []string{"links", "edges"} {
		list, ok := data[key].([]any)
		if !ok {
			continue
		}
		for _, e := range list {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}
			for _, f := range []string{"source", "target"} {
				if _, ok := m[f]; ok {
					m[f] = rewrite(m[f])
				}
			}
		}
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return raw
	}
	return append(out, '\n')
}

func writeGraphSidecar(p *Project, graphJSON string, repos []string) error {
	dir := filepath.Dir(graphJSON)
	body := fmt.Sprintf("# Project graph: %s\n\nBuilt %s\n\nRepos:\n", p.Name, time.Now().UTC().Format(time.RFC3339))
	for _, n := range repos {
		body += "- " + n + "\n"
	}
	body += "\nQuery with: graphify query \"...\" (cwd or --graph " + graphJSON + ")\n"
	return os.WriteFile(filepath.Join(dir, "GRAPH_REPORT.md"), []byte(body), 0o644)
}
