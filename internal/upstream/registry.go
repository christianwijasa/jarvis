package upstream

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed upstream.yaml
var raw []byte

const (
	ModeVendored = "vendored"
	ModeAdapted  = "adapted"
	ModeOriginal = "original"
	KindSkill    = "skill"

	// DefaultRole is the role a catalog entry serves when it has no roles list.
	DefaultRole = "engineer"
)

// Registry is the machine-readable provenance file.
type Registry struct {
	Source  Source  `yaml:"source"`
	Entries []Entry `yaml:"entries"`
}

type Source struct {
	HubRepo        string `yaml:"hub_repo"`
	HubCommit      string `yaml:"hub_commit"`
	SnapshotBranch string `yaml:"snapshot_branch"`
	SnapshotCommit string `yaml:"snapshot_commit"`
}

type Entry struct {
	Name         string   `yaml:"name"`
	Kind         string   `yaml:"kind"`
	Pack         string   `yaml:"pack"`
	Recommended  bool     `yaml:"recommended"`
	Roles        []string `yaml:"roles,omitempty"`
	Summary      string   `yaml:"summary"`
	LocalPath    string   `yaml:"local_path"`
	Mode         string   `yaml:"mode"`
	UpstreamRepo string   `yaml:"upstream_repo"`
	UpstreamPath string   `yaml:"upstream_path"`
	PinnedCommit string   `yaml:"pinned_commit"`
	PinnedTag    string   `yaml:"pinned_tag"`
	Licence      string   `yaml:"licence"`
	Note         string   `yaml:"note"`
}

// Load returns the embedded registry.
func Load() (*Registry, error) {
	var reg Registry
	if err := yaml.Unmarshal(raw, &reg); err != nil {
		return nil, fmt.Errorf("parse upstream.yaml: %w", err)
	}
	return &reg, nil
}

// ModeFor returns the registry mode for a skill or rule name.
func (r *Registry) ModeFor(name string) string {
	for _, e := range r.Entries {
		if e.Name == name {
			return e.Mode
		}
	}
	return ""
}

func (r *Registry) Skill(name string) (Entry, bool) {
	for _, e := range r.Entries {
		if e.Kind == KindSkill && e.Name == name {
			return e, true
		}
	}
	return Entry{}, false
}

func (r *Registry) Skills() []Entry {
	var out []Entry
	for _, e := range r.Entries {
		if e.Kind == KindSkill {
			out = append(out, e)
		}
	}
	return out
}

// RecommendedSkillNames is the recommended set for the engineer role. It is
// the set `jarvis init --yes` installs when no role is given.
func (r *Registry) RecommendedSkillNames() []string {
	return r.SkillsForRole(DefaultRole)
}

// SkillsForRole returns the sorted names of recommended skills whose roles
// list contains role.
func (r *Registry) SkillsForRole(role string) []string {
	var out []string
	for _, e := range r.Entries {
		if e.Kind == KindSkill && e.Recommended && e.HasRole(role) {
			out = append(out, e.Name)
		}
	}
	sort.Strings(out)
	return out
}

// EffectiveRoles returns the roles list, or [engineer] when the entry has none.
func (e Entry) EffectiveRoles() []string {
	if len(e.Roles) == 0 {
		return []string{DefaultRole}
	}
	return e.Roles
}

// HasRole reports whether the entry serves role.
func (e Entry) HasRole(role string) bool {
	for _, r := range e.EffectiveRoles() {
		if r == role {
			return true
		}
	}
	return false
}

func (e Entry) HasUpstream() bool {
	return e.Mode != ModeOriginal && strings.TrimSpace(e.UpstreamRepo) != ""
}

func (e Entry) PinnedKnown() bool {
	p := strings.TrimSpace(e.PinnedCommit)
	return p != "" && !strings.EqualFold(p, "unknown")
}

// Bundled is true when Jarvis ships the skill text in this repo (original or adapted).
func (e Entry) Bundled() bool {
	return e.Kind == KindSkill && e.Mode != ModeVendored && strings.TrimSpace(e.LocalPath) != ""
}
