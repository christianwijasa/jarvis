package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	FileName        = "jarvis.yaml"
	DefaultLocalDir = ".agents/skills"
	ScopeGlobal     = "global"
	ScopeLocal      = "local"
	ScopeBoth       = "both"
)

// ErrNotFound means no jarvis.yaml was found in the current directory or parents.
var ErrNotFound = errors.New("jarvis.yaml not found")

// Config is the project-specific file Jarvis reads. The CLI itself stays
// tool-agnostic; tracker site, history path, and repo map belong here.
type Config struct {
	Tracker Tracker           `yaml:"tracker,omitempty"`
	Jira    Jira              `yaml:"jira,omitempty"`
	History History           `yaml:"history"`
	Repos   map[string]string `yaml:"repos"`
	Tiers   Tiers             `yaml:"tiers"`
	Gates   Gates             `yaml:"gates"`
	Skills  Skills            `yaml:"skills"`
	Project string            `yaml:"project,omitempty"`
}

// Tracker is optional policy for a work-item system (Jira, Trello, …).
// Jarvis never stores a token here and never calls the tracker API.
type Tracker struct {
	Kind    string `yaml:"kind,omitempty" json:"kind,omitempty"`
	Site    string `yaml:"site,omitempty" json:"site,omitempty"`
	Project string `yaml:"project,omitempty" json:"project,omitempty"`
	MCP     string `yaml:"mcp,omitempty" json:"mcp,omitempty"`
}

// Jira is a deprecated jarvis.yaml alias for Tracker (kind jira).
type Jira struct {
	Site    string `yaml:"site,omitempty"`
	Project string `yaml:"project,omitempty"`
}

// Set reports whether URL or MCP id is present (token is stored separately).
func (t Tracker) Set() bool {
	return strings.TrimSpace(t.Site) != "" || strings.TrimSpace(t.MCP) != ""
}

// Configured is true when a plan should use a tracker (URL, MCP, or token).
func Configured(t Tracker, tokenSet bool) bool {
	return t.Set() || tokenSet
}

type History struct {
	Path   string `yaml:"path"`
	GitHub string `yaml:"github"`
}

type Tiers struct {
	Default string           `yaml:"default"`
	Rules   []map[string]any `yaml:"rules"`
}

type Gates struct {
	TestCase               bool `yaml:"test_case"`
	TestData               bool `yaml:"test_data"`
	BackendCoveragePercent int  `yaml:"backend_coverage_percent"`
	ClientCoveragePercent  int  `yaml:"client_coverage_percent"`
}

// Skills records which catalog entries this project uses and where agents
// should load them from. Canonical files always live under Home/skills.
type Skills struct {
	Home      string   `yaml:"home,omitempty"`
	Scope     string   `yaml:"scope"`
	LocalDir  string   `yaml:"local_dir,omitempty"`
	Copy      bool     `yaml:"copy,omitempty"`
	Installed []string `yaml:"installed"`
}

// Find walks from start (or the current working directory) up toward
// the filesystem root looking for jarvis.yaml. It returns the absolute
// path of the first match.
func Find(start string) (string, error) {
	dir := start
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// Load reads and parses a config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.Skills.Normalize()
	cfg.mergeDeprecatedJira()
	return &cfg, nil
}

func (c *Config) mergeDeprecatedJira() {
	if c == nil {
		return
	}
	if !c.Tracker.Set() && c.Tracker.Kind == "" && c.Tracker.Project == "" {
		if c.Jira.Site != "" || c.Jira.Project != "" {
			c.Tracker.Kind = "jira"
			c.Tracker.Site = c.Jira.Site
			c.Tracker.Project = c.Jira.Project
		}
	}
	c.Jira = Jira{}
}

// Write marshals cfg to path.
func Write(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	cfg.Skills.Normalize()
	cfg.mergeDeprecatedJira()
	var b strings.Builder
	b.WriteString("# Project settings for Jarvis. TRDs belong in history.path, not here.\n")
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// ExampleYAML is the commented template copied by `jarvis init` before
// the chosen skill list is written back.
func ExampleYAML() []byte {
	return exampleYAML
}

func (s *Skills) Normalize() {
	if s.Scope == "" {
		s.Scope = ScopeBoth
	}
	if s.LocalDir == "" {
		s.LocalDir = DefaultLocalDir
	}
	s.Installed = uniqueSorted(s.Installed)
}

func ParseScope(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case ScopeGlobal:
		return ScopeGlobal, nil
	case ScopeLocal:
		return ScopeLocal, nil
	case ScopeBoth, "":
		return ScopeBoth, nil
	default:
		return "", fmt.Errorf("scope must be global, local, or both (got %q)", v)
	}
}

func (s Skills) UsesGlobal() bool {
	return s.Scope == ScopeGlobal || s.Scope == ScopeBoth
}

func (s Skills) UsesLocal() bool {
	return s.Scope == ScopeLocal || s.Scope == ScopeBoth
}

func (s Skills) Has(name string) bool {
	for _, n := range s.Installed {
		if n == name {
			return true
		}
	}
	return false
}

func (s *Skills) Add(names ...string) {
	s.Installed = uniqueSorted(append(s.Installed, names...))
}

func (s *Skills) Remove(name string) {
	var out []string
	for _, n := range s.Installed {
		if n != name {
			out = append(out, n)
		}
	}
	s.Installed = out
}

// HomeDir is $JARVIS_HOME, else skills.home from yaml, else ~/.jarvis.
func HomeDir(cfg *Config) (string, error) {
	if v := strings.TrimSpace(os.Getenv("JARVIS_HOME")); v != "" {
		return filepath.Abs(expandHome(v))
	}
	if cfg != nil {
		if v := strings.TrimSpace(cfg.Skills.Home); v != "" {
			return filepath.Abs(expandHome(v))
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jarvis"), nil
}

func GlobalSkillsDir(cfg *Config) (string, error) {
	home, err := HomeDir(cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "skills"), nil
}

func LocalSkillsDir(cfg *Config, projectRoot string) string {
	d := DefaultLocalDir
	if cfg != nil && strings.TrimSpace(cfg.Skills.LocalDir) != "" {
		d = cfg.Skills.LocalDir
	}
	if filepath.IsAbs(d) {
		return d
	}
	return filepath.Join(projectRoot, d)
}

// AbsPath expands ~ and joins a relative path against baseDir.
func AbsPath(baseDir, p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	p = expandHome(p)
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	if strings.TrimSpace(baseDir) == "" {
		return filepath.Abs(p)
	}
	return filepath.Abs(filepath.Join(baseDir, p))
}

func expandHome(p string) string {
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

func uniqueSorted(in []string) []string {
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
	sort.Strings(out)
	return out
}
