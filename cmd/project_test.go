package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
)

func TestProjectCreateAddRemoveUseCurrentJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	out, err := execute(t, "project", "create", "shop", "--description", "Example shop",
		"--repo", "cart-service=https://github.com/example/cart-service.git",
		"--repo", "payment-service=https://github.com/example/payment-service.git@~/src/payment-service")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	p, err := project.Load("shop")
	if err != nil {
		t.Fatal(err)
	}
	if p.Repos[1].Path != "~/src/payment-service" {
		t.Fatalf("path %q", p.Repos[1].Path)
	}

	if _, err := execute(t, "project", "add-repo", "shop", "search-service", "--url", "https://github.com/example/search-service.git"); err != nil {
		t.Fatal(err)
	}
	if _, err := execute(t, "project", "remove-repo", "shop", "search-service"); err != nil {
		t.Fatal(err)
	}

	list, err := execute(t, "project", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var lj map[string]any
	if err := json.Unmarshal([]byte(list), &lj); err != nil {
		t.Fatal(err)
	}
	if lj["schema_version"] != SchemaVersion || lj["kind"] != "project_list" {
		t.Fatalf("%v", lj)
	}

	show, err := execute(t, "project", "show", "shop", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var sj map[string]any
	if err := json.Unmarshal([]byte(show), &sj); err != nil {
		t.Fatal(err)
	}
	if sj["kind"] != "project" {
		t.Fatalf("%v", sj)
	}
	if _, ok := sj["graph_json"].(string); !ok {
		t.Fatal("show json missing graph_json")
	}
	if sj["tracker_token_set"] != false {
		t.Fatalf("token set %v", sj["tracker_token_set"])
	}

	cfg := &config.Config{Skills: config.Skills{Scope: config.ScopeGlobal}}
	if err := config.Write(config.FileName, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := execute(t, "project", "use", "shop"); err != nil {
		t.Fatal(err)
	}
	cur, err := execute(t, "project", "current", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cur, `"project": "shop"`) {
		t.Fatalf("current %s", cur)
	}

	_, err = execute(t, "project", "delete", "shop")
	if err == nil {
		t.Fatal("non-tty delete without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("%v", err)
	}
	if _, err := execute(t, "project", "delete", "shop", "--yes"); err != nil {
		t.Fatal(err)
	}
	_ = home
}

func TestProjectCreateUpdateTracker(t *testing.T) {
	t.Setenv("JARVIS_HOME", t.TempDir())
	out, err := execute(t, "project", "create", "shop",
		"--tracker-kind", "jira",
		"--tracker-site", "https://example.atlassian.net",
		"--tracker-mcp", "plugin-atlassian-atlassian",
		"--tracker-token", "super-secret")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	p, err := project.Load("shop")
	if err != nil {
		t.Fatal(err)
	}
	if p.Tracker.Site != "https://example.atlassian.net" || p.Tracker.MCP != "plugin-atlassian-atlassian" {
		t.Fatalf("%+v", p.Tracker)
	}
	tok, err := project.ReadToken("shop")
	if err != nil || tok != "super-secret" {
		t.Fatalf("token %q %v", tok, err)
	}
	show, err := execute(t, "project", "show", "shop", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(show, "super-secret") {
		t.Fatal("token leaked in show json")
	}
	if !strings.Contains(show, `"tracker_token_set": true`) {
		t.Fatalf("%s", show)
	}

	if _, err := execute(t, "project", "update", "shop", "--tracker-site", "https://other.example"); err != nil {
		t.Fatal(err)
	}
	p, _ = project.Load("shop")
	if p.Tracker.Site != "https://other.example" {
		t.Fatalf("site %s", p.Tracker.Site)
	}
	if _, err := execute(t, "project", "update", "shop", "--unset-tracker-token"); err != nil {
		t.Fatal(err)
	}
	if project.TokenSet("shop") {
		t.Fatal("token should be unset")
	}
}

func TestProjectCreateBadName(t *testing.T) {
	t.Setenv("JARVIS_HOME", t.TempDir())
	_, err := execute(t, "project", "create", "Not_Valid")
	if err == nil {
		t.Fatal("expected kebab-case error")
	}
}

func TestInitProjectFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	dir := t.TempDir()
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, "init", "--skills", "grilling", "--scope", "global", "--project", "shop")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	cfg, err := config.Load(config.FileName)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Project != "shop" {
		t.Fatalf("project %q", cfg.Project)
	}
	if _, err := project.Load("shop"); err != nil {
		t.Fatal(err)
	}
}

func TestProjectGraphPathJSON(t *testing.T) {
	t.Setenv("JARVIS_HOME", t.TempDir())
	if _, err := execute(t, "project", "create", "shop"); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, "project", "graph-path", "shop", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["kind"] != "project_graph_path" {
		t.Fatalf("%v", got)
	}
	gj, _ := got["graph_json"].(string)
	if !strings.Contains(gj, filepath.Join("projects", "shop", "graphify-out", "graph.json")) {
		t.Fatalf("path %s", gj)
	}
}
