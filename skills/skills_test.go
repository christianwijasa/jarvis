package skills

import (
	"strings"
	"testing"
)

func TestVersionAndBundledSkills(t *testing.T) {
	if Version() != "0.4.0" {
		t.Fatalf("skill pack version: %q", Version())
	}
	names, err := Names()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"code-review", "debugging", "development", "domain-modeling",
		"generate-pr-description", "graphify", "grill-with-docs", "grilling",
		"plan-review", "prd-planning", "test-case-design", "trd-pr-planning",
	}
	got := strings.Join(names, ",")
	for _, n := range want {
		if !contains(names, n) {
			t.Errorf("missing bundled skill %q in %s", n, got)
		}
	}
	if contains(names, "archify") {
		t.Fatal("archify must not be bundled in the repo")
	}
	body, pack, err := Read("graphify")
	if err != nil || pack != PackBundled || body == "" {
		t.Fatalf("graphify: pack=%s err=%v", pack, err)
	}
	_, _, err = Read("archify")
	if err == nil {
		t.Fatal("archify must not be readable from the embed")
	}
	_, _, err = Read("does-not-exist")
	if err == nil {
		t.Fatal("expected unknown skill")
	}
	tree, err := Tree("grilling")
	if err != nil || len(tree["SKILL.md"]) == 0 {
		t.Fatalf("grilling tree: %v", err)
	}
	if len(tree["references/product-questions.md"]) == 0 {
		t.Fatal("grilling product question set missing")
	}
	prd, err := Tree("prd-planning")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"SKILL.md", "references/acceptance-criteria.md", "references/prd-review-checklist.md"} {
		if len(prd[f]) == 0 {
			t.Errorf("prd-planning missing %s", f)
		}
	}
}

func TestPlanReviewBundled(t *testing.T) {
	tree, err := Tree("plan-review")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"SKILL.md", "references/lenses.md", "references/ready-to-build.md", "references/resolve.md"} {
		if !strings.Contains(string(tree[f]), "google/elephant-goldfish-model") {
			t.Errorf("plan-review %s must credit google/elephant-goldfish-model", f)
		}
	}
	for _, need := range []string{"GO", "NO-GO", "Explain it back", "Tier A"} {
		if !strings.Contains(string(tree["SKILL.md"])+string(tree["references/ready-to-build.md"]), need) {
			t.Errorf("plan-review missing %q", need)
		}
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
