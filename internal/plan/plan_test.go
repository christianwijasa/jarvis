package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
)

func TestValidateKey(t *testing.T) {
	for _, k := range []string{"TEAM-123", "card-Ab3xYz", "ENG-9", "42"} {
		if err := ValidateKey(k); err != nil {
			t.Fatalf("%s: %v", k, err)
		}
	}
	for _, k := range []string{"", "../x", "a/b", "has space"} {
		if err := ValidateKey(k); err == nil {
			t.Fatalf("expected reject %q", k)
		}
	}
}

func TestScaffoldAndNoOverwrite(t *testing.T) {
	hist := t.TempDir()
	req := Request{
		Key:         "TEAM-123",
		HistoryRoot: hist,
		YAMLRepos:   []string{"api"},
		Tier:        "A",
		Tracker:     config.Tracker{Kind: "jira", Site: "https://example.atlassian.net"},
	}
	res, err := Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != ActionCreated || res.Source != SourceTracker {
		t.Fatalf("%+v", res)
	}
	body, err := os.ReadFile(filepath.Join(res.Dir, "TRD-Architecture.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "issue: TEAM-123") || !strings.Contains(string(body), "tracker: jira") {
		t.Fatalf("%s", body)
	}
	if _, err := os.Stat(filepath.Join(res.Dir, "TRD-Testing.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(res.Dir, "TRD-Architecture.md"), []byte("keep-me"), 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if again.Action != ActionFound {
		t.Fatalf("action %s", again.Action)
	}
	kept, _ := os.ReadFile(filepath.Join(res.Dir, "TRD-Architecture.md"))
	if string(kept) != "keep-me" {
		t.Fatalf("overwrote: %s", kept)
	}
}

func TestTierBNoTestingUserSource(t *testing.T) {
	hist := t.TempDir()
	res, err := Run(Request{Key: "42", HistoryRoot: hist, Tier: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != SourceUser || res.TestSource != "user-prompt" {
		t.Fatalf("%+v", res)
	}
	if _, err := os.Stat(filepath.Join(res.Dir, "TRD-Testing.md")); !os.IsNotExist(err) {
		t.Fatal("tier B should omit testing")
	}
}

func TestMultiRepoNeedsFlag(t *testing.T) {
	_, err := Run(Request{
		Key:         "TEAM-1",
		HistoryRoot: t.TempDir(),
		YAMLRepos:   []string{"api", "web"},
	})
	if err == nil || !strings.Contains(err.Error(), "--repo") {
		t.Fatalf("%v", err)
	}
}

func TestMultiRepoLayoutAndDryRun(t *testing.T) {
	hist := t.TempDir()
	req := Request{
		Key:         "TEAM-1",
		HistoryRoot: hist,
		YAMLRepos:   []string{"api", "web"},
		Selected:    []string{"api", "web"},
		Tier:        "A",
		DryRun:      true,
	}
	res, err := Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != ActionDryRun {
		t.Fatal(res.Action)
	}
	joined := strings.Join(res.Files, "\n")
	if !strings.Contains(joined, filepath.Join("api", "TRD-Implementation.md")) {
		t.Fatalf("%s", joined)
	}
	if _, err := os.Stat(filepath.Join(hist, "docs", "plans", "TEAM-1")); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote files")
	}
	req.DryRun = false
	res, err = Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(res.Dir, "web", "TRD-Implementation.md")); err != nil {
		t.Fatal(err)
	}
}

func TestFromFileAndUnknownRepo(t *testing.T) {
	hist := t.TempDir()
	src := filepath.Join(t.TempDir(), "req.md")
	if err := os.WriteFile(src, []byte("# user req\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(Request{Key: "N-1", HistoryRoot: hist, FromPath: src, YAMLRepos: []string{"api"}, Selected: []string{"api"}})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(res.Dir, "REQUIREMENT.md"))
	if string(b) != "# user req\n" {
		t.Fatalf("%s", b)
	}
	_, err = Run(Request{Key: "N-2", HistoryRoot: hist, YAMLRepos: []string{"api"}, Selected: []string{"nope"}})
	if err == nil {
		t.Fatal("expected unknown repo")
	}
}

func TestResolveTrackerPrefersProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	p := &project.Project{Name: "shop", Tracker: config.Tracker{Kind: "jira", Site: "https://proj.example"}}
	if err := project.Save(p); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Project: "shop",
		Tracker: config.Tracker{Kind: "trello", Site: "https://yaml.example"},
	}
	got, tok := ResolveTracker(cfg)
	if tok || got.Site != "https://proj.example" {
		t.Fatalf("%+v tok=%v", got, tok)
	}
}

func TestHistoryRoot(t *testing.T) {
	dir := t.TempDir()
	hist := filepath.Join(dir, "hist")
	if err := os.Mkdir(hist, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, config.FileName)
	root, err := HistoryRoot(cfgPath, &config.Config{History: config.History{Path: "hist"}})
	if err != nil {
		t.Fatal(err)
	}
	if root != hist {
		t.Fatalf("%s vs %s", root, hist)
	}
	if _, err := HistoryRoot(cfgPath, &config.Config{}); err == nil {
		t.Fatal("expected empty history.path error")
	}
}
