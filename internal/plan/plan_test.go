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

func TestPRDScaffoldAndNoOverwrite(t *testing.T) {
	hist := t.TempDir()
	req := Request{Key: "TEAM-7", HistoryRoot: hist, Type: TypePRD, YAMLRepos: []string{"api", "web"}, Tracker: config.Tracker{Kind: "jira", MCP: "tracker-mcp"}}
	res, err := Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != ActionCreated || res.Type != TypePRD || len(res.Files) != 1 || res.Source != SourceTracker {
		t.Fatalf("%+v", res)
	}
	path := filepath.Join(hist, "docs", "plans", "TEAM-7", PRDFileName)
	if res.Files[0] != path || res.PRDFile != path {
		t.Fatalf("files %v prd %s", res.Files, res.PRDFile)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("PRD run wrote more than PRD.md: %v", entries)
	}
	body, _ := os.ReadFile(path)
	for _, want := range []string{"issue: TEAM-7", "type: prd", "tracker: jira"} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("missing %q in\n%s", want, body)
		}
	}
	if strings.Contains(string(body), "## Source notes") {
		t.Fatal("Source notes section must be removed without --from")
	}
	st1, _ := os.Stat(path)
	again, err := Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if again.Action != ActionFound || len(again.Files) != 1 || again.Files[0] != path {
		t.Fatalf("%+v", again)
	}
	st2, _ := os.Stat(path)
	body2, _ := os.ReadFile(path)
	if !st1.ModTime().Equal(st2.ModTime()) || string(body) != string(body2) {
		t.Fatal("second PRD run changed the file")
	}
}

func TestPRDAndTRDDoNotBlockEachOther(t *testing.T) {
	hist := t.TempDir()
	trd := Request{Key: "TEAM-8", HistoryRoot: hist, YAMLRepos: []string{"api"}, Tier: "B"}
	if _, err := Run(trd); err != nil {
		t.Fatal(err)
	}
	prd, err := Run(Request{Key: "TEAM-8", HistoryRoot: hist, Type: TypePRD})
	if err != nil {
		t.Fatal(err)
	}
	if prd.Action != ActionCreated {
		t.Fatalf("TRD files blocked the PRD: %+v", prd)
	}
	foundTRD, err := Run(trd)
	if err != nil {
		t.Fatal(err)
	}
	if foundTRD.Action != ActionFound || foundTRD.PRDFile == "" {
		t.Fatalf("TRD run should report PRD.md: %+v", foundTRD)
	}
	for _, f := range foundTRD.Files {
		if filepath.Base(f) == PRDFileName {
			t.Fatal("PRD.md listed as a TRD file")
		}
	}

	hist2 := t.TempDir()
	if _, err := Run(Request{Key: "TEAM-9", HistoryRoot: hist2, Type: TypePRD}); err != nil {
		t.Fatal(err)
	}
	res, err := Run(Request{Key: "TEAM-9", HistoryRoot: hist2, YAMLRepos: []string{"api"}, Tier: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != ActionCreated || res.PRDFile == "" {
		t.Fatalf("PRD.md blocked the TRD: %+v", res)
	}
}

func TestPRDFromAndFlagsAndDryRun(t *testing.T) {
	hist := t.TempDir()
	from := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(from, []byte("Customers want saved carts.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dry, err := Run(Request{Key: "TEAM-10", HistoryRoot: hist, Type: TypePRD, DryRun: true, FromPath: from})
	if err != nil {
		t.Fatal(err)
	}
	if dry.Action != ActionDryRun {
		t.Fatalf("%+v", dry)
	}
	if _, err := os.Stat(filepath.Join(hist, "docs")); !os.IsNotExist(err) {
		t.Fatal("dry run wrote files")
	}
	res, err := Run(Request{Key: "TEAM-10", HistoryRoot: hist, Type: TypePRD, FromPath: from})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(res.Files[0])
	if !strings.Contains(string(body), "## Source notes\n\nCustomers want saved carts.") {
		t.Fatalf("source notes missing:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(res.Dir, "REQUIREMENT.md")); !os.IsNotExist(err) {
		t.Fatal("PRD run must not write REQUIREMENT.md")
	}
	if _, err := Run(Request{Key: "TEAM-11", HistoryRoot: hist, Type: TypePRD, Selected: []string{"api"}}); err == nil {
		t.Fatal("expected --repo to fail for a PRD")
	}
	if _, err := Run(Request{Key: "TEAM-11", HistoryRoot: hist, Type: TypePRD, Tier: "A"}); err == nil {
		t.Fatal("expected --tier to fail for a PRD")
	}
	if _, err := Run(Request{Key: "TEAM-11", HistoryRoot: hist, Type: "rfc"}); err == nil {
		t.Fatal("expected unknown type to fail")
	}
}

// Scaffolds land in the team's history repo. No file may name the tool.
func TestScaffoldsContainNoToolName(t *testing.T) {
	from := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(from, []byte("Customers want saved carts.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reqs := []Request{
		{Key: "A-1", Tier: "A", YAMLRepos: []string{"api"}, Tracker: config.Tracker{Kind: "jira", Site: "https://example.atlassian.net"}},
		{Key: "A-2", Tier: "A", YAMLRepos: []string{"api", "web"}, Selected: []string{"api", "web"}},
		{Key: "B-1", Tier: "B", YAMLRepos: []string{"api"}, FromPath: from},
		{Key: "C-1", Tier: "C", YAMLRepos: []string{"api"}, Tracker: config.Tracker{Kind: "jira", MCP: "tracker-mcp"}},
		{Key: "P-1", Type: TypePRD, Tracker: config.Tracker{Kind: "jira", Site: "https://example.atlassian.net"}},
		{Key: "P-2", Type: TypePRD, FromPath: from},
	}
	for _, req := range reqs {
		req.HistoryRoot = t.TempDir()
		res, err := Run(req)
		if err != nil {
			t.Fatalf("%s: %v", req.Key, err)
		}
		if len(res.Files) == 0 {
			t.Fatalf("%s: no files", req.Key)
		}
		for _, f := range res.Files {
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(strings.ToLower(string(b)), "jarvis") {
				t.Errorf("%s: %s names the tool:\n%s", req.Key, f, b)
			}
		}
		if strings.EqualFold(req.Tier, "A") {
			b, _ := os.ReadFile(filepath.Join(res.Dir, "TRD-Architecture.md"))
			if !strings.Contains(string(b), "test_source: TRD-Testing\n") {
				t.Errorf("%s: tier A test_source:\n%s", req.Key, b)
			}
		}
	}
}

func TestNormalizeTestSourceLegacyAlias(t *testing.T) {
	cases := map[string]string{
		"jarvis:TRD-Testing": "TRD-Testing",
		"TRD-Testing":        "TRD-Testing",
		"tracker:TEAM-1":     "tracker:TEAM-1",
		"user-prompt":        "user-prompt",
	}
	for in, want := range cases {
		if got := NormalizeTestSource(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}
