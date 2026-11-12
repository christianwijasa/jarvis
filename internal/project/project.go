package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	FileName        = "project.yaml"
	SchemaVersion   = "jarvis.project/v1"
	DefaultGraphRel = "graphify-out"
)

var (
	ErrNotFound = errors.New("project not found")
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
)

// Project is a named group of repositories. Files live under
// $JARVIS_HOME/projects/<name>/project.yaml.
type Project struct {
	SchemaVersion string         `yaml:"schema_version" json:"schema_version"`
	Name          string         `yaml:"name" json:"name"`
	Description   string         `yaml:"description,omitempty" json:"description,omitempty"`
	Tracker       config.Tracker `yaml:"tracker,omitempty" json:"tracker,omitempty"`
	Repos         []Repo         `yaml:"repos" json:"repos"`
	Graph         Graph          `yaml:"graph,omitempty" json:"graph"`
}

type Repo struct {
	Name          string `yaml:"name" json:"name"`
	URL           string `yaml:"url,omitempty" json:"url,omitempty"`
	Path          string `yaml:"path,omitempty" json:"path,omitempty"`
	DefaultBranch string `yaml:"default_branch,omitempty" json:"default_branch,omitempty"`
}

type Graph struct {
	OutDir string `yaml:"out_dir,omitempty" json:"out_dir,omitempty"`
}

func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("project name %q must be lowercase kebab-case (e.g. shop, cart-service)", name)
	}
	return nil
}

func (p *Project) Validate() error {
	if err := ValidateName(p.Name); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, r := range p.Repos {
		if err := ValidateName(r.Name); err != nil {
			return fmt.Errorf("repo %s: %w", r.Name, err)
		}
		if seen[r.Name] {
			return fmt.Errorf("duplicate repo name %q", r.Name)
		}
		seen[r.Name] = true
	}
	return nil
}

func (p *Project) Repo(name string) (Repo, bool) {
	for _, r := range p.Repos {
		if r.Name == name {
			return r, true
		}
	}
	return Repo{}, false
}

func (p *Project) SetRepo(r Repo) error {
	if err := ValidateName(r.Name); err != nil {
		return fmt.Errorf("repo %s: %w", r.Name, err)
	}
	for i, existing := range p.Repos {
		if existing.Name == r.Name {
			p.Repos[i] = r
			return nil
		}
	}
	p.Repos = append(p.Repos, r)
	return nil
}

func (p *Project) RemoveRepo(name string) error {
	var out []Repo
	for _, r := range p.Repos {
		if r.Name != name {
			out = append(out, r)
		}
	}
	if len(out) == len(p.Repos) {
		return fmt.Errorf("repo %q not in project %s", name, p.Name)
	}
	p.Repos = out
	return nil
}

func Dir(name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}
	home, err := config.HomeDir(nil)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "projects", name), nil
}

func Path(name string) (string, error) {
	dir, err := Dir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, FileName), nil
}

func Load(name string) (*Project, error) {
	path, err := Path(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return nil, err
	}
	var p Project
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if p.Name == "" {
		p.Name = name
	}
	if p.SchemaVersion == "" {
		p.SchemaVersion = SchemaVersion
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

func Save(p *Project) error {
	if p.SchemaVersion == "" {
		p.SchemaVersion = SchemaVersion
	}
	if err := p.Validate(); err != nil {
		return err
	}
	dir, err := Dir(p.Name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Jarvis project: a named group of repositories.\n")
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(p); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, FileName), []byte(b.String()), 0o644)
}

func Delete(name string) error {
	path, err := Path(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return err
	}
	return os.Remove(path)
}

func List() ([]Project, error) {
	home, err := config.HomeDir(nil)
	if err != nil {
		return nil, err
	}
	root := filepath.Join(home, "projects")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := Load(e.Name())
		if err != nil {
			continue
		}
		out = append(out, *p)
	}
	return out, nil
}

func ExpandPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		if p == "~" {
			return home
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func (p *Project) AbsRepoPath(r Repo) string {
	return ExpandPath(r.Path)
}

func (p *Project) GraphDir() (string, error) {
	dir, err := Dir(p.Name)
	if err != nil {
		return "", err
	}
	rel := strings.TrimSpace(p.Graph.OutDir)
	if rel == "" {
		rel = DefaultGraphRel
	}
	if filepath.IsAbs(rel) {
		return rel, nil
	}
	return filepath.Join(dir, rel), nil
}

func (p *Project) GraphJSON() (string, error) {
	d, err := p.GraphDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "graph.json"), nil
}

// ExtractOutDir is the --out directory for `graphify extract` of one repo.
// graphify writes <dir>/graphify-out/graph.json; merge-graphs prefixes
// node ids with parent.parent of that file, which is the repo name.
func (p *Project) ExtractOutDir(repoName string) (string, error) {
	dir, err := Dir(p.Name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "extract", repoName), nil
}
