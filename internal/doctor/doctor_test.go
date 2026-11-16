package doctor

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
)

func TestCheckWithConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte("jira: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Check(&buf, dir)
	out := buf.String()
	if !strings.Contains(out, "git:") || !strings.Contains(out, "gh:") || !strings.Contains(out, "graphify:") || !strings.Contains(out, "jarvis.yaml: pass") {
		t.Fatalf("unexpected doctor output:\n%s", out)
	}
}

func TestCheckWithoutConfig(t *testing.T) {
	var buf bytes.Buffer
	err := Check(&buf, t.TempDir())
	if err == nil {
		t.Fatal("expected error when jarvis.yaml is missing")
	}
	if !strings.Contains(buf.String(), "jarvis.yaml: fail") {
		t.Fatalf("unexpected doctor output:\n%s", buf.String())
	}
}

func TestCheckSkillLinks(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	cfg := &config.Config{Skills: config.Skills{
		Scope:     config.ScopeBoth,
		LocalDir:  ".agents/skills",
		Installed: []string{"grilling"},
	}}
	if err := config.Write(filepath.Join(dir, config.FileName), cfg); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := Check(&buf, dir)
	if err == nil {
		t.Fatal("expected fail for missing skill files")
	}
	if !strings.Contains(buf.String(), "skills: fail") {
		t.Fatalf("output:\n%s", buf.String())
	}

	empty := t.TempDir()
	cfg2 := &config.Config{Skills: config.Skills{Scope: config.ScopeGlobal}}
	if err := config.Write(filepath.Join(empty, config.FileName), cfg2); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	_ = Check(&buf, empty)
	if !strings.Contains(buf.String(), "skills: pass") {
		t.Fatalf("empty install should pass:\n%s", buf.String())
	}
}

func TestCheckWarnsMissingGraphStamps(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	p := &project.Project{Name: "shop"}
	if err := project.Save(p); err != nil {
		t.Fatal(err)
	}
	gjson, err := p.GraphJSON()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(gjson), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gjson, []byte(`{"nodes":[{"id":"x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Project: "shop", Skills: config.Skills{Scope: config.ScopeGlobal}}
	if err := config.Write(filepath.Join(dir, config.FileName), cfg); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Check(&buf, dir)
	if !strings.Contains(buf.String(), "stamps.json is missing") {
		t.Fatalf("output:\n%s", buf.String())
	}
}

// gitOnlyPATH puts only git on PATH, so gh and graphify are missing.
func gitOnlyPATH(t *testing.T) {
	t.Helper()
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not on PATH")
	}
	bin := t.TempDir()
	if err := os.Symlink(gitPath, filepath.Join(bin, "git")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
}

func fakeInstall(t *testing.T, home string, names ...string) {
	t.Helper()
	for _, n := range names {
		d := filepath.Join(home, "skills", n)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func productSetup(t *testing.T, installed []string, withHistory bool) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	fakeInstall(t, home, installed...)
	if err := project.Save(&project.Project{Name: "docs", Role: project.RoleProduct}); err != nil {
		t.Fatal(err)
	}
	ws := t.TempDir()
	hist := filepath.Join(ws, "history")
	if withHistory {
		gitInit(t, hist)
	}
	cfg := &config.Config{Project: "docs", History: config.History{Path: hist}, Skills: config.Skills{Scope: config.ScopeGlobal, Installed: installed}}
	if err := config.Write(filepath.Join(ws, config.FileName), cfg); err != nil {
		t.Fatal(err)
	}
	return ws
}

func row(r Report, name string) (Item, bool) {
	for _, c := range r.Checks {
		if c.Name == name {
			return c, true
		}
	}
	return Item{}, false
}

func TestProductPassesWithoutGraphify(t *testing.T) {
	gitOnlyPATH(t)
	ws := productSetup(t, []string{"domain-modeling", "grilling", "prd-planning"}, true)
	r := Collect(ws)
	if !r.OK || r.Role != "product" || r.RoleSource != "project" {
		var buf bytes.Buffer
		WriteText(&buf, r)
		t.Fatalf("product doctor should pass:\n%s", buf.String())
	}
	if _, ok := row(r, "graphify"); ok {
		t.Fatal("product role must not check graphify")
	}
	if gh, _ := row(r, "gh"); !gh.Warn {
		t.Fatalf("gh should warn: %+v", gh)
	}
	if rem, _ := row(r, "history remote"); !rem.Warn {
		t.Fatalf("history remote should warn: %+v", rem)
	}
	if tr, _ := row(r, "tracker"); !tr.Warn {
		t.Fatalf("tracker should warn: %+v", tr)
	}
}

func TestProductFailsMissingSkillOrHistory(t *testing.T) {
	gitOnlyPATH(t)
	ws := productSetup(t, []string{"domain-modeling", "grilling"}, true)
	r := Collect(ws)
	rs, _ := row(r, "role skills")
	if r.OK || rs.OK || !strings.Contains(strings.Join(rs.Problems, " "), "jarvis skills install --role product") {
		t.Fatalf("missing prd-planning should fail with the fix: %+v", rs)
	}

	ws = productSetup(t, []string{"domain-modeling", "grilling", "prd-planning"}, false)
	r = Collect(ws)
	h, _ := row(r, "history.path")
	if r.OK || h.OK {
		t.Fatalf("missing history.path should fail: %+v", h)
	}
}

func TestEngineerUnchangedAndNewRowsWarn(t *testing.T) {
	gitOnlyPATH(t)
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	ws := t.TempDir()
	cfg := &config.Config{Skills: config.Skills{Scope: config.ScopeGlobal}}
	if err := config.Write(filepath.Join(ws, config.FileName), cfg); err != nil {
		t.Fatal(err)
	}
	r := Collect(ws)
	if r.Role != "engineer" || r.RoleSource != "default" {
		t.Fatalf("role %s/%s", r.Role, r.RoleSource)
	}
	if g, _ := row(r, "graphify"); g.OK {
		t.Fatal("engineer must still fail on missing graphify")
	}
	for _, n := range []string{"role skills", "history.path"} {
		it, ok := row(r, n)
		if !ok || !it.OK || !it.Warn {
			t.Fatalf("%s should be warn only for engineer: %+v", n, it)
		}
	}
}

func TestMissingRoleIsWarning(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	if err := project.Save(&project.Project{Name: "shop"}); err != nil {
		t.Fatal(err)
	}
	ws := t.TempDir()
	cfg := &config.Config{Project: "shop", Skills: config.Skills{Scope: config.ScopeGlobal}}
	if err := config.Write(filepath.Join(ws, config.FileName), cfg); err != nil {
		t.Fatal(err)
	}
	r := Collect(ws)
	it, _ := row(r, "role")
	if !it.OK || !it.Warn || !strings.Contains(strings.Join(it.Problems, " "), "jarvis project update shop --role engineer|product") {
		t.Fatalf("missing role should warn with the fix: %+v", it)
	}
	if r.Checks[0].Name != "role" {
		t.Fatalf("role should be the first row, got %s", r.Checks[0].Name)
	}
}

func TestLocalDirInsideHistoryFails(t *testing.T) {
	for _, roleName := range []string{"", project.RoleProduct} {
		home := t.TempDir()
		t.Setenv("JARVIS_HOME", home)
		if err := project.Save(&project.Project{Name: "p", Role: roleName}); err != nil {
			t.Fatal(err)
		}
		ws := t.TempDir()
		cfg := &config.Config{Project: "p", History: config.History{Path: "."}, Skills: config.Skills{Scope: config.ScopeBoth}}
		if err := config.Write(filepath.Join(ws, config.FileName), cfg); err != nil {
			t.Fatal(err)
		}
		r := Collect(ws)
		it, _ := row(r, "local skills dir")
		if it.OK || r.OK {
			t.Fatalf("role %q: local dir inside history should fail: %+v", roleName, it)
		}
	}
}
