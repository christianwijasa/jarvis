package skills

import (
	"strings"
	"testing"
)

func TestVersionAndBundledSkills(t *testing.T) {
	if Version() != "0.2.0" {
		t.Fatalf("skill pack version: %q", Version())
	}
	names, err := Names()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"code-review", "debugging", "development", "domain-modeling",
		"generate-pr-description", "graphify", "grill-with-docs", "grilling",
		"test-case-design", "trd-pr-planning",
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
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
