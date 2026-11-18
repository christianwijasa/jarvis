package role

import (
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
)

func TestResolve(t *testing.T) {
	t.Setenv("JARVIS_HOME", t.TempDir())
	if err := project.Save(&project.Project{Name: "plain"}); err != nil {
		t.Fatal(err)
	}
	if err := project.Save(&project.Project{Name: "docs", Role: Product}); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, project, role, source, plan string
		missing, err                      bool
	}{
		{"no config", "", Engineer, SourceDefault, PlanTRD, false, false},
		{"no role", "plain", Engineer, SourceDefault, PlanTRD, true, false},
		{"product", "docs", Product, SourceProject, PlanPRD, false, false},
		{"missing project", "gone", Engineer, SourceDefault, PlanTRD, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var cfg *config.Config
			if c.project != "" {
				cfg = &config.Config{Project: c.project}
			}
			got := Resolve(cfg)
			if got.Role != c.role || got.Source != c.source || got.PlanType != c.plan || got.Missing != c.missing || (got.Err != nil) != c.err {
				t.Fatalf("%+v", got)
			}
		})
	}
	prod := Resolve(&config.Config{Project: "docs"})
	if strings.Join(prod.Skills, ",") != "domain-modeling,grilling,plan-review,prd-planning" {
		t.Fatalf("product skills %v", prod.Skills)
	}
	if m := MissingSkills(Product, []string{"grilling", "domain-modeling", "plan-review"}); len(m) != 1 || m[0] != "prd-planning" {
		t.Fatalf("missing %v", m)
	}
	if !strings.Contains(MissingWarning("plain"), "jarvis project update plain --role engineer|product") {
		t.Fatal(MissingWarning("plain"))
	}
}
