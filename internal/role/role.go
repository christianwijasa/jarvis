// Package role resolves the effective role of the active project. The role
// lives in project.yaml. A project with no role, or no active project,
// behaves as engineer.
package role

import (
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
	"github.com/christianwijasa/jarvis/internal/upstream"
)

const (
	Engineer = project.RoleEngineer
	Product  = project.RoleProduct

	// SourceProject means the role comes from project.yaml.
	SourceProject = "project"
	// SourceDefault means no role is stored, so engineer is used.
	SourceDefault = "default"

	PlanPRD = "prd"
	PlanTRD = "trd"
)

// Valid reports whether r is a known role.
func Valid(r string) bool { return project.ValidRole(r) }

// PlanTypeFor returns the plan type `jarvis plan` scaffolds by default.
func PlanTypeFor(r string) string {
	if r == Product {
		return PlanPRD
	}
	return PlanTRD
}

// Resolved is the effective role for the active project.
type Resolved struct {
	Role     string
	Source   string
	Project  string
	Skills   []string
	PlanType string
	// Missing is true when a project is active and loads, but stores no role.
	Missing bool
	// Err is set when the active project is named but cannot be loaded.
	Err error
}

// Resolve returns the effective role for cfg.Project.
func Resolve(cfg *config.Config) Resolved {
	name := ""
	if cfg != nil {
		name = strings.TrimSpace(cfg.Project)
	}
	if name == "" {
		return finish(Resolved{Role: Engineer, Source: SourceDefault})
	}
	p, err := project.Load(name)
	if err != nil {
		return finish(Resolved{Role: Engineer, Source: SourceDefault, Project: name, Err: err})
	}
	return ForProject(p)
}

// ForProject returns the effective role for a loaded project.
func ForProject(p *project.Project) Resolved {
	if p.HasRole() {
		return finish(Resolved{Role: p.Role, Source: SourceProject, Project: p.Name})
	}
	return finish(Resolved{Role: Engineer, Source: SourceDefault, Project: p.Name, Missing: true})
}

func finish(r Resolved) Resolved {
	r.PlanType = PlanTypeFor(r.Role)
	r.Skills = SkillsFor(r.Role)
	if r.Skills == nil {
		r.Skills = []string{}
	}
	return r
}

// SkillsFor returns the catalog skill set for a role.
func SkillsFor(r string) []string {
	reg, err := upstream.Load()
	if err != nil {
		return nil
	}
	return reg.SkillsForRole(r)
}

// MissingSkills returns the role skills that are not in installed.
func MissingSkills(r string, installed []string) []string {
	have := map[string]bool{}
	for _, n := range installed {
		have[n] = true
	}
	var out []string
	for _, n := range SkillsFor(r) {
		if !have[n] {
			out = append(out, n)
		}
	}
	return out
}

// FixCommand is the command that sets a role on a project.
func FixCommand(name string) string {
	return "jarvis project update " + name + " --role engineer|product"
}

// MissingWarning is the one-line warning for a project with no role.
func MissingWarning(name string) string {
	return "project " + name + " has no role; run: " + FixCommand(name)
}
