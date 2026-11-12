package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/christianwijasabfi/atlas/internal/config"
	"github.com/christianwijasabfi/atlas/internal/project"
	"github.com/christianwijasabfi/atlas/templates"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

const (
	ActionCreated = "created"
	ActionFound   = "found"
	ActionDryRun  = "dry_run"
	SourceTracker = "tracker"
	SourceUser    = "user"
)

func ValidateKey(key string) error {
	if !keyPattern.MatchString(key) {
		return fmt.Errorf("plan key %q must be one path segment (letters, digits, ., _, -; max 128)", key)
	}
	return nil
}

type Data struct {
	Issue       string
	TrackerKind string
	Tier        string
	TestSource  string
	ReposCSV    string
}

type Request struct {
	Key         string
	HistoryRoot string
	YAMLRepos   []string
	Selected    []string
	Tier        string
	Tracker     config.Tracker
	TokenSet    bool
	FromPath    string
	DryRun      bool
}

type Result struct {
	Action     string
	Source     string
	Dir        string
	Files      []string
	Tracker    config.Tracker
	TokenSet   bool
	Tier       string
	Repos      []string
	TestSource string
}

func SourceOf(t config.Tracker, tokenSet bool) string {
	if config.Configured(t, tokenSet) {
		return SourceTracker
	}
	return SourceUser
}

func ResolveTracker(cfg *config.Config) (config.Tracker, bool) {
	if cfg == nil {
		return config.Tracker{}, false
	}
	if name := strings.TrimSpace(cfg.Project); name != "" {
		if p, err := project.Load(name); err == nil {
			tok := project.TokenSet(name)
			if config.Configured(p.Tracker, tok) {
				return p.Tracker, tok
			}
		}
	}
	return cfg.Tracker, false
}

func Run(req Request) (*Result, error) {
	if err := ValidateKey(req.Key); err != nil {
		return nil, err
	}
	tier := strings.TrimSpace(req.Tier)
	if tier == "" {
		tier = "A"
	}
	selected, err := selectRepos(req.Selected, req.YAMLRepos)
	if err != nil {
		return nil, err
	}
	source := SourceOf(req.Tracker, req.TokenSet)
	testSource := "tracker:" + req.Key
	if strings.EqualFold(tier, "A") {
		testSource = "atlas:TRD-Testing"
	} else if source == SourceUser {
		testSource = "user-prompt"
	}
	dir := filepath.Join(req.HistoryRoot, "docs", "plans", req.Key)
	rels, err := fileRels(selected, strings.EqualFold(tier, "A"))
	if err != nil {
		return nil, err
	}
	existing := scanExisting(dir)
	res := &Result{
		Source:     source,
		Dir:        dir,
		Tracker:    req.Tracker,
		TokenSet:   req.TokenSet,
		Tier:       tier,
		Repos:      selected,
		TestSource: testSource,
	}
	if len(existing) > 0 {
		res.Action = ActionFound
		res.Files = existing
		return res, nil
	}
	abs := make([]string, 0, len(rels)+1)
	for _, rel := range rels {
		abs = append(abs, filepath.Join(dir, rel))
	}
	if strings.TrimSpace(req.FromPath) != "" {
		abs = append(abs, filepath.Join(dir, "REQUIREMENT.md"))
	}
	if req.DryRun {
		res.Action = ActionDryRun
		res.Files = abs
		return res, nil
	}
	data := Data{
		Issue:       req.Key,
		TrackerKind: strings.TrimSpace(req.Tracker.Kind),
		Tier:        tier,
		TestSource:  testSource,
		ReposCSV:    strings.Join(selected, ", "),
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	for _, rel := range rels {
		tmplName := templateName(rel)
		body, err := render(tmplName, data)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return nil, err
		}
	}
	if src := strings.TrimSpace(req.FromPath); src != "" {
		b, err := os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("read --from: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "REQUIREMENT.md"), b, 0o644); err != nil {
			return nil, err
		}
	}
	res.Action = ActionCreated
	res.Files = abs
	return res, nil
}

func selectRepos(selected, yamlRepos []string) ([]string, error) {
	yamlRepos = uniqueKeep(yamlRepos)
	if len(selected) == 0 {
		if len(yamlRepos) > 1 {
			return nil, fmt.Errorf("atlas.yaml has more than one repo; pass --repo for each target")
		}
		return yamlRepos, nil
	}
	yamlSet := map[string]bool{}
	for _, n := range yamlRepos {
		yamlSet[n] = true
	}
	var out []string
	for _, n := range selected {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if len(yamlSet) > 0 && !yamlSet[n] {
			return nil, fmt.Errorf("repo %q is not in atlas.yaml repos", n)
		}
		out = append(out, n)
	}
	return uniqueKeep(out), nil
}

func fileRels(repos []string, testing bool) ([]string, error) {
	var rels []string
	rels = append(rels, "TRD-Architecture.md")
	if len(repos) >= 2 {
		for _, r := range repos {
			rels = append(rels, filepath.Join(r, "TRD-Implementation.md"))
		}
	} else {
		rels = append(rels, "TRD-Implementation.md")
	}
	if testing {
		rels = append(rels, "TRD-Testing.md")
	}
	return rels, nil
}

func templateName(rel string) string {
	base := filepath.Base(rel)
	return base
}

func render(name string, data Data) (string, error) {
	raw, err := templates.FS.ReadFile(name)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

func scanExisting(dir string) []string {
	var out []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		base := info.Name()
		if (strings.HasPrefix(base, "TRD-") && strings.HasSuffix(base, ".md")) || base == "REQUIREMENT.md" {
			out = append(out, path)
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func uniqueKeep(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range in {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return out
}

// RepoKeys returns sorted atlas.yaml repo names.
func RepoKeys(repos map[string]string) []string {
	var keys []string
	for k := range repos {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func HistoryRoot(cfgPath string, cfg *config.Config) (string, error) {
	if cfg == nil || strings.TrimSpace(cfg.History.Path) == "" {
		return "", fmt.Errorf("history.path is empty in atlas.yaml")
	}
	root, err := config.AbsPath(filepath.Dir(cfgPath), cfg.History.Path)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("history.path %s: %w", root, err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("history.path %s is not a directory", root)
	}
	return root, nil
}
