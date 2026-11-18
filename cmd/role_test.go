package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
	"github.com/christianwijasa/jarvis/internal/update"
)

// roleWorkspace makes a JARVIS_HOME, a working dir with jarvis.yaml and a
// history dir, and chdirs into it.
func roleWorkspace(t *testing.T, active string) (home, dir string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	dir = t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "hist"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Project: active,
		History: config.History{Path: "hist"},
		Repos:   map[string]string{"api": "../api", "web": "../web"},
		Tiers:   config.Tiers{Default: "B"},
		Skills:  config.Skills{Scope: config.ScopeGlobal},
	}
	if err := config.Write(config.FileName, cfg); err != nil {
		t.Fatal(err)
	}
	return home, dir
}

func saveProject(t *testing.T, name, role string) {
	t.Helper()
	if err := project.Save(&project.Project{Name: name, Role: role}); err != nil {
		t.Fatal(err)
	}
}

func roleOf(t *testing.T, name string) string {
	t.Helper()
	p, err := project.Load(name)
	if err != nil {
		t.Fatal(err)
	}
	return p.Role
}

func fakeTTY(t *testing.T) {
	t.Helper()
	old := terminalCheck
	t.Cleanup(func() { terminalCheck = old })
	terminalCheck = func(io.Reader) bool { return true }
}

const roleQuestion = "What is your role in project shop? [engineer/product]"

func TestRolePromptOnTTYSavesAnswer(t *testing.T) {
	roleWorkspace(t, "")
	saveProject(t, "shop", "")
	fakeTTY(t)

	out, errOut, err := executeSplit(t, strings.NewReader("product\n"), "project", "use", "shop")
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out, errOut)
	}
	if !strings.Contains(errOut, roleQuestion) {
		t.Fatalf("expected the role question on stderr: %q", errOut)
	}
	if got := roleOf(t, "shop"); got != "product" {
		t.Fatalf("role saved %q", got)
	}
	if !strings.Contains(out, "role product") {
		t.Fatalf("use output: %s", out)
	}
	if !strings.Contains(errOut, "Run: jarvis skills install --role product") {
		t.Fatalf("expected skills hint: %q", errOut)
	}

	// Asked once: the next command does not ask again.
	_, errOut, err = executeSplit(t, strings.NewReader(""), "project", "current")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errOut, "What is your role") {
		t.Fatalf("asked twice: %q", errOut)
	}
}

func TestRolePromptEnterMeansEngineerAndRetries(t *testing.T) {
	roleWorkspace(t, "")
	saveProject(t, "shop", "")
	saveProject(t, "web", "")
	fakeTTY(t)

	if _, errOut, err := executeSplit(t, strings.NewReader("\n"), "project", "show", "shop"); err != nil {
		t.Fatalf("%v %s", err, errOut)
	}
	if got := roleOf(t, "shop"); got != "engineer" {
		t.Fatalf("Enter should save engineer, got %q", got)
	}

	_, errOut, err := executeSplit(t, strings.NewReader("qa\nproduct\n"), "project", "graph", "web")
	if err == nil || !strings.Contains(err.Error(), "role product") {
		t.Fatalf("graph on a product project should refuse after the answer: %v", err)
	}
	if !strings.Contains(errOut, "Please answer engineer or product.") || roleOf(t, "web") != "product" {
		t.Fatalf("retry failed: %q role=%q", errOut, roleOf(t, "web"))
	}
}

func TestRolePromptSkippedWithJSONOnTTY(t *testing.T) {
	roleWorkspace(t, "shop")
	saveProject(t, "shop", "")
	fakeTTY(t)
	out, errOut, err := executeSplit(t, strings.NewReader("product\n"), "project", "current", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errOut, "What is your role") {
		t.Fatal("--json must never prompt")
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got["role"] != "engineer" || got["role_source"] != "default" || got["plan_type"] != "trd" {
		t.Fatalf("%v", got)
	}
	if roleOf(t, "shop") != "" {
		t.Fatal("--json must not save a role")
	}
}

func TestRoleMissingWarnsWithoutTTY(t *testing.T) {
	roleWorkspace(t, "shop")
	saveProject(t, "shop", "")
	warn := "project shop has no role; run: jarvis project update shop --role engineer|product"

	out, errOut, err := executeSplit(t, nil, "plan", "TEAM-1", "--repo", "api", "--json")
	if err != nil {
		t.Fatalf("%v\n%s", err, errOut)
	}
	if strings.Count(errOut, warn) != 1 {
		t.Fatalf("expected one warning, got %q", errOut)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["type"] != "trd" || got["role"] != "engineer" || got["role_source"] != "default" {
		t.Fatalf("%v", got)
	}

	for _, args := range [][]string{{"project", "current"}, {"project", "show", "shop"}, {"project", "use", "shop"}, {"doctor"}} {
		_, errOut, _ := executeSplit(t, nil, args...)
		if !strings.Contains(errOut, warn) || strings.Contains(errOut, "What is your role") {
			t.Fatalf("%v: stderr %q", args, errOut)
		}
	}
	if roleOf(t, "shop") != "" {
		t.Fatal("non-TTY must not save a role")
	}

	_, errOut, err = executeSplit(t, nil, "project", "create", "cart")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut, "project cart has no role") {
		t.Fatalf("create without --role off TTY should warn: %q", errOut)
	}
}

func TestProjectCreateAsksRoleOnTTY(t *testing.T) {
	roleWorkspace(t, "")
	fakeTTY(t)
	// role, description, empty repo list, then four empty tracker answers
	in := strings.NewReader("product\nDocs\n\n\n\n\n\n")
	if _, errOut, err := executeSplit(t, in, "project", "create", "shop"); err != nil {
		t.Fatalf("%v %s", err, errOut)
	}
	p, err := project.Load("shop")
	if err != nil {
		t.Fatal(err)
	}
	if p.Role != "product" || p.Description != "Docs" {
		t.Fatalf("%+v", p)
	}
}

func TestDoctorCmdRole(t *testing.T) {
	roleWorkspace(t, "shop")
	saveProject(t, "shop", "")
	out, _, _ := executeSplit(t, nil, "doctor", "--json")
	var got struct {
		Role       string `json:"role"`
		RoleSource string `json:"role_source"`
		Checks     []struct {
			Name     string   `json:"name"`
			OK       bool     `json:"ok"`
			Warn     bool     `json:"warn"`
			Problems []string `json:"problems"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got.Role != "engineer" || got.RoleSource != "default" || len(got.Checks) == 0 {
		t.Fatalf("%+v", got)
	}
	r := got.Checks[0]
	if r.Name != "role" || !r.OK || !r.Warn || !strings.Contains(strings.Join(r.Problems, " "), "jarvis project update shop --role engineer|product") {
		t.Fatalf("role row %+v", r)
	}

	fakeTTY(t)
	_, errOut, _ := executeSplit(t, strings.NewReader("product\n"), "doctor")
	if !strings.Contains(errOut, roleQuestion) || roleOf(t, "shop") != "product" {
		t.Fatalf("doctor on TTY should ask and save: %q", errOut)
	}
}

func TestUpdateAsksMissingRoles(t *testing.T) {
	roleWorkspace(t, "")
	saveProject(t, "alpha", "")
	saveProject(t, "beta", "")
	saveProject(t, "gamma", "product")
	old := applyUpdate
	t.Cleanup(func() { applyUpdate = old })
	applyUpdate = func(opts update.Options) (update.Result, error) {
		return update.Result{Current: "v0.0.5", Latest: "v0.0.5", Check: opts.Check}, nil
	}

	// Without a TTY: no prompt, nothing saved.
	if _, errOut, err := executeSplit(t, nil, "update"); err != nil || strings.Contains(errOut, "What is your role") {
		t.Fatalf("non-TTY update prompted: %v %q", err, errOut)
	}
	if roleOf(t, "alpha") != "" {
		t.Fatal("non-TTY update saved a role")
	}

	fakeTTY(t)
	// --check never asks.
	if _, errOut, _ := executeSplit(t, strings.NewReader("product\n"), "update", "--check"); strings.Contains(errOut, "What is your role") {
		t.Fatal("update --check prompted")
	}
	_, errOut, err := executeSplit(t, strings.NewReader("product\n\n"), "update")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"alpha", "beta"} {
		if !strings.Contains(errOut, "What is your role in project "+n+"?") {
			t.Fatalf("not asked for %s: %q", n, errOut)
		}
	}
	if strings.Contains(errOut, "project gamma?") {
		t.Fatal("asked for a project that has a role")
	}
	if roleOf(t, "alpha") != "product" || roleOf(t, "beta") != "engineer" || roleOf(t, "gamma") != "product" {
		t.Fatalf("roles %q %q %q", roleOf(t, "alpha"), roleOf(t, "beta"), roleOf(t, "gamma"))
	}
}

func TestPlanTypeFollowsRole(t *testing.T) {
	_, dir := roleWorkspace(t, "docs")
	saveProject(t, "docs", "product")
	out, errOut, err := executeSplit(t, nil, "plan", "TEAM-2", "--json")
	if err != nil {
		t.Fatalf("product plan with a multi-repo map: %v %s", err, errOut)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	prd := filepath.Join(dir, "hist", "docs", "plans", "TEAM-2", "PRD.md")
	if got["type"] != "prd" || got["role"] != "product" || got["role_source"] != "project" || got["prd_file"] != prd {
		t.Fatalf("%v", got)
	}
	if _, err := os.Stat(prd); err != nil {
		t.Fatal(err)
	}
	text, _, err := executeSplit(t, nil, "plan", "TEAM-2")
	if err != nil || !strings.Contains(text, "found TEAM-2") || !strings.Contains(text, "Load prd-planning and grilling") {
		t.Fatalf("%v %s", err, text)
	}
	if _, _, err := executeSplit(t, nil, "plan", "TEAM-3", "--repo", "api"); err == nil || !strings.Contains(err.Error(), "--repo") {
		t.Fatalf("product --repo: %v", err)
	}
	if _, _, err := executeSplit(t, nil, "plan", "TEAM-3", "--type", "prd", "--tier", "A"); err == nil || !strings.Contains(err.Error(), "--tier") {
		t.Fatalf("prd --tier: %v", err)
	}
	if _, _, err := executeSplit(t, nil, "plan", "TEAM-3", "--type", "rfc"); err == nil {
		t.Fatal("bad --type should fail")
	}
	// A product owner may still ask for a TRD on purpose; it reports the PRD.
	text, _, err = executeSplit(t, nil, "plan", "TEAM-2", "--type", "trd", "--repo", "api")
	if err != nil || !strings.Contains(text, "PRD.md found: read it first") {
		t.Fatalf("%v %s", err, text)
	}

	if _, _, err := executeSplit(t, nil, "project", "update", "docs", "--role", "engineer"); err != nil {
		t.Fatal(err)
	}
	out, _, err = executeSplit(t, nil, "plan", "TEAM-4", "--repo", "api", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"type": "trd"`) {
		t.Fatalf("engineer should get a TRD: %s", out)
	}
}

func TestProjectRoleFlags(t *testing.T) {
	roleWorkspace(t, "")
	if _, _, err := executeSplit(t, nil, "project", "create", "shop", "--role", "product"); err != nil {
		t.Fatal(err)
	}
	if roleOf(t, "shop") != "product" {
		t.Fatal("create --role")
	}
	if _, _, err := executeSplit(t, nil, "project", "create", "bad", "--role", "qa"); err == nil {
		t.Fatal("create --role qa should fail")
	}
	if _, _, err := executeSplit(t, nil, "project", "update", "shop", "--role", "engineer"); err != nil || roleOf(t, "shop") != "engineer" {
		t.Fatalf("update --role: %v", err)
	}
	if _, _, err := executeSplit(t, nil, "project", "update", "shop", "--role", "bad"); err == nil {
		t.Fatal("update --role bad should fail")
	}
	if _, _, err := executeSplit(t, nil, "project", "update", "shop", "--role", "product", "--unset-role"); err == nil {
		t.Fatal("--role with --unset-role should fail")
	}
	if _, _, err := executeSplit(t, nil, "project", "update", "shop", "--unset-role"); err != nil || roleOf(t, "shop") != "" {
		t.Fatalf("unset-role: %v", err)
	}

	saveProject(t, "docs", "product")
	out, _, err := executeSplit(t, nil, "project", "show", "docs", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["role"] != "product" || got["role_source"] != "project" {
		t.Fatalf("%v", got)
	}
	for _, args := range [][]string{{"project", "graph", "docs"}, {"project", "clone", "docs"}} {
		_, _, err := executeSplit(t, nil, args...)
		if err == nil || !strings.Contains(err.Error(), "project docs has role product; it does not clone or graph") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

func TestProjectCurrentJSONWithoutProject(t *testing.T) {
	roleWorkspace(t, "")
	out, errOut, err := executeSplit(t, nil, "project", "current", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if errOut != "" {
		t.Fatalf("no project should not warn: %q", errOut)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["role"] != "engineer" || got["role_source"] != "default" || got["plan_type"] != "trd" {
		t.Fatalf("%v", got)
	}
}

func TestInitRoleProduct(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if _, _, err := executeSplit(t, nil, "init", "--role", "product", "--yes"); err == nil || !strings.Contains(err.Error(), "--project") {
		t.Fatalf("--role without --project: %v", err)
	}
	saveProject(t, "eng", "engineer")
	if _, _, err := executeSplit(t, nil, "init", "--role", "product", "--project", "eng", "--yes"); err == nil || !strings.Contains(err.Error(), "already has role engineer") {
		t.Fatalf("role conflict: %v", err)
	}
	if _, err := os.Stat(config.FileName); !os.IsNotExist(err) {
		t.Fatal("failed init wrote jarvis.yaml")
	}
	out, errOut, err := executeSplit(t, nil, "init", "--role", "product", "--project", "docs", "--scope", "global", "--yes")
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out, errOut)
	}
	cfg, err := config.Load(config.FileName)
	if err != nil {
		t.Fatal(err)
	}
	got := append([]string{}, cfg.Skills.Installed...)
	sort.Strings(got)
	if strings.Join(got, ",") != "domain-modeling,grilling,plan-review,prd-planning" {
		t.Fatalf("installed %v", got)
	}
	raw, _ := os.ReadFile(config.FileName)
	if !strings.Contains(string(raw), "repos: {}") {
		t.Fatalf("product init should write repos: {}\n%s", raw)
	}
	if cfg.Project != "docs" || roleOf(t, "docs") != "product" {
		t.Fatalf("project %q role %q", cfg.Project, roleOf(t, "docs"))
	}
}

func TestSkillsInstallRoleAndApplies(t *testing.T) {
	roleWorkspace(t, "docs")
	saveProject(t, "docs", "product")
	out, errOut, err := executeSplit(t, nil, "skills", "install", "--role", "product")
	if err != nil {
		t.Fatalf("%v %s", err, errOut)
	}
	for _, n := range []string{"prd-planning", "plan-review", "grilling", "domain-modeling"} {
		if !strings.Contains(out, "Installed "+n) {
			t.Fatalf("missing %s: %s", n, out)
		}
	}
	if strings.Contains(out, "Installed graphify") {
		t.Fatal("product set must not include graphify")
	}
	if _, _, err := executeSplit(t, nil, "skills", "install"); err == nil {
		t.Fatal("install with no names and no --role should fail")
	}
	js, _, err := executeSplit(t, nil, "skills", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var rep skillsReport
	if err := json.Unmarshal([]byte(js), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Role != "product" || rep.RoleSource != "project" || len(rep.Installed) != 4 {
		t.Fatalf("%+v", rep)
	}
	for _, it := range rep.Available {
		if it.Name == "graphify" && it.Applies {
			t.Fatal("graphify should not apply to product")
		}
	}
	_, note, err := executeSplit(t, nil, "skills", "show", "graphify")
	if err != nil || !strings.Contains(note, "does not apply to role product") {
		t.Fatalf("show note: %v %q", err, note)
	}
}

func TestAgentHelpMentionsRoles(t *testing.T) {
	out, err := execute(t, "help", "--agent")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"jarvis project current --json", "--type prd", "role_source: default", "Never guess the role", "## Roles (per project)", "prd-planning"} {
		if !strings.Contains(out, want) {
			t.Errorf("agent guide missing %q", want)
		}
	}
}

// A stdin that looks like a terminal but is empty (for example /dev/null)
// must not block or fail: Jarvis warns and keeps the default.
func TestRolePromptEmptyStdinFallsBack(t *testing.T) {
	roleWorkspace(t, "")
	fakeTTY(t)
	_, errOut, err := executeSplit(t, strings.NewReader(""), "project", "create", "shop")
	if err != nil {
		t.Fatalf("create with empty stdin: %v", err)
	}
	if roleOf(t, "shop") != "" || !strings.Contains(errOut, "project shop has no role") {
		t.Fatalf("role %q stderr %q", roleOf(t, "shop"), errOut)
	}
	_, errOut, err = executeSplit(t, strings.NewReader(""), "project", "show", "shop")
	if err != nil || !strings.Contains(errOut, "project shop has no role") {
		t.Fatalf("show: %v %q", err, errOut)
	}
}
