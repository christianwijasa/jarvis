package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasabfi/atlas/internal/buildinfo"
	"github.com/christianwijasabfi/atlas/internal/config"
	"github.com/christianwijasabfi/atlas/internal/install"
	"github.com/christianwijasabfi/atlas/internal/upstream"
)

func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return executeIn(t, nil, args...)
}

func executeIn(t *testing.T, stdin io.Reader, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if stdin != nil {
		cmd.SetIn(stdin)
	} else {
		cmd.SetIn(bytes.NewReader(nil))
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestRootWiresExpectedCommands(t *testing.T) {
	root := NewRootCmd()
	want := []string{"version", "update", "plan", "skills", "doctor", "init", "agent-help", "project"}
	got := map[string]bool{}
	for _, c := range root.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing command %q", name)
		}
	}
	for _, c := range root.Commands() {
		if c.Name() != "skills" {
			continue
		}
		sub := map[string]bool{}
		for _, s := range c.Commands() {
			sub[s.Name()] = true
		}
		for _, name := range []string{"show", "upstream", "install", "remove"} {
			if !sub[name] {
				t.Errorf("missing skills %s", name)
			}
		}
	}
}

func TestVersionOutput(t *testing.T) {
	out, err := execute(t, "version")
	if err != nil {
		t.Fatal(err)
	}
	for _, need := range []string{"CLI version:", buildinfo.Version, "Skill pack:", "0.2.0", "Config file:"} {
		if !strings.Contains(out, need) {
			t.Errorf("version output missing %q\n%s", need, out)
		}
	}
}

func TestUpdatePlaceholder(t *testing.T) {
	out, err := execute(t, "update")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "tagged release") {
		t.Fatalf("expected tagged-release note, got:\n%s", out)
	}
}

func TestPlanRequiresKey(t *testing.T) {
	_, err := execute(t, "plan")
	if err == nil {
		t.Fatal("expected error when JIRA key is missing")
	}
}

func TestPlanPlaceholder(t *testing.T) {
	out, err := execute(t, "plan", "TEAM-123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "TEAM-123") || !strings.Contains(out, "trd-pr-planning") {
		t.Fatalf("expected issue key and skill name:\n%s", out)
	}
}

func TestSkillsListWithoutProject(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ATLAS_HOME", filepath.Join(dir, "home"))
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	out, err := execute(t, "skills")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Installed:") || !strings.Contains(out, "(none)") {
		t.Fatalf("expected empty installed:\n%s", out)
	}
	if !strings.Contains(out, "Available:") || !strings.Contains(out, "archify") || !strings.Contains(out, "graphify") {
		t.Fatalf("expected catalog available:\n%s", out)
	}

	shown, err := execute(t, "skills", "show", "grilling")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(shown, "bundled") || !strings.Contains(shown, "mode:adapted") {
		t.Fatalf("skills show header:\n%s", shown)
	}

	_, err = execute(t, "skills", "show", "archify")
	if err == nil {
		t.Fatal("expected error for uninstalled vendored skill")
	}

	_, err = execute(t, "skills", "show", "does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestSkillsUpstreamMocked(t *testing.T) {
	old := upstreamRemote
	t.Cleanup(func() { upstreamRemote = old })
	upstreamRemote = func() upstream.Remote {
		return fakeRemote{
			sha: map[string]string{
				"https://github.com/mattpocock/skills":                 "49dd158d1076134a641b33efb035946536778336",
				"https://github.com/obra/superpowers":                  "bb92a77741419a4ab5f06e711a283343f1ada0c3",
				"https://github.com/DietrichGebert/ponytail":           "09c434ab73ca43761c5d8b0c5d341926c2fba710",
				"https://github.com/multica-ai/andrej-karpathy-skills": "2c606141936f1eeef17fa3043a72095b4765b9c2",
				"https://github.com/Graphify-Labs/graphify":            "961b78e57a10e9c5bb98421ff3e45b40be73542b",
				"https://github.com/tt-a1i/archify":                    "208acba8d412335185b2056fedf54b3405df43fb",
			},
		}
	}
	out, err := execute(t, "skills", "upstream")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Read-only check. Nothing was applied.") {
		t.Fatalf("missing read-only banner:\n%s", out)
	}
	if !strings.Contains(out, "graphify (skill, recommended, mode:adapted)") {
		t.Fatalf("graphify should be recommended:\n%s", out)
	}
	if !strings.Contains(out, "archify (skill, recommended, mode:vendored)") {
		t.Fatalf("archify should be recommended vendored:\n%s", out)
	}
	if strings.Contains(out, "status:  unpinned") {
		t.Fatalf("did not expect unpinned entries:\n%s", out)
	}
}

func TestSkillsUpstreamRemoteError(t *testing.T) {
	old := upstreamRemote
	t.Cleanup(func() { upstreamRemote = old })
	upstreamRemote = func() upstream.Remote {
		return fakeRemote{err: errors.New("offline")}
	}
	out, err := execute(t, "skills", "upstream")
	if err == nil {
		t.Fatal("expected lookup failures")
	}
	if !strings.Contains(out, "status:  error") {
		t.Fatalf("expected error status:\n%s", out)
	}
}

type fakeRemote struct {
	sha map[string]string
	err error
}

func (f fakeRemote) LatestCommit(repoURL, path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if sha, ok := f.sha[repoURL+"|"+path]; ok {
		return sha, nil
	}
	if sha, ok := f.sha[repoURL]; ok {
		return sha, nil
	}
	return "ffffffffffffffffffffffffffffffffffffffff", nil
}

func TestDoctorPrintsPassFailLines(t *testing.T) {
	out, _ := execute(t, "doctor")
	for _, tool := range []string{"git:", "gh:", "graphify:", "atlas.yaml:"} {
		if !strings.Contains(out, tool) {
			t.Errorf("doctor missing line for %s\n%s", tool, out)
		}
	}
}

func TestInitRequiresFlagsWhenNotTTY(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	_, err = execute(t, "init")
	if err == nil {
		t.Fatal("expected error without flags on non-tty")
	}
	if !strings.Contains(err.Error(), "--yes") || !strings.Contains(err.Error(), "--skills") {
		t.Fatalf("error should name flags, got %v", err)
	}
}

func TestInitRecommendedAndSkillsCommands(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	old := skillFetcher
	t.Cleanup(func() { skillFetcher = old })
	skillFetcher = install.FuncFetcher(func(repo, commit, path string) (map[string][]byte, error) {
		if err := install.AssertPinned(commit); err != nil {
			return nil, err
		}
		if commit != "208acba8d412335185b2056fedf54b3405df43fb" {
			t.Fatalf("install fetched %s not pin", commit)
		}
		return map[string][]byte{"SKILL.md": []byte("# Archify pin\n")}, nil
	})

	out, err := execute(t, "init", "--recommended", "--yes", "--scope", "both")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Wrote") {
		t.Fatalf("init output: %s", out)
	}
	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Skills.Scope != config.ScopeBoth {
		t.Fatalf("scope %s", cfg.Skills.Scope)
	}
	if len(cfg.Skills.Installed) != 11 {
		t.Fatalf("installed %v", cfg.Skills.Installed)
	}
	if _, err := os.Stat(filepath.Join(home, "skills", "grilling", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(dir, ".agents", "skills", "archify")
	st, err := os.Lstat(local)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected local symlink for archify")
	}
	body, _ := os.ReadFile(filepath.Join(home, "skills", "archify", "SKILL.md"))
	if string(body) != "# Archify pin\n" {
		t.Fatalf("archify body %q", body)
	}

	list, err := execute(t, "skills")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(list, "Installed:") || !strings.Contains(list, "grilling") || !strings.Contains(list, "archify") {
		t.Fatalf("list:\n%s", list)
	}
	if !strings.Contains(list, "global") || !strings.Contains(list, "local") {
		t.Fatalf("expected locations:\n%s", list)
	}

	_, err = execute(t, "init", "--yes")
	if err == nil {
		t.Fatal("second init should fail when file exists")
	}

	if _, err := execute(t, "skills", "remove", "archify"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = config.Load(filepath.Join(dir, config.FileName))
	if cfg.Skills.Has("archify") {
		t.Fatal("archify still listed")
	}
	if _, err := os.Stat(filepath.Join(home, "skills", "archify", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("archify files remain: %v", err)
	}

	out, err = execute(t, "skills", "install", "archify")
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Installed archify") {
		t.Fatalf("install output %s", out)
	}
}

func TestInitSkillsFlagAndInteractive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	out, err := execute(t, "init", "--skills", "grilling", "--scope", "global")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Skills.Scope != config.ScopeGlobal || !cfg.Skills.Has("grilling") || cfg.Skills.Has("archify") {
		t.Fatalf("cfg %+v", cfg.Skills)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); !os.IsNotExist(err) {
		t.Fatal("global scope should not create local dir")
	}

	dir2 := t.TempDir()
	if err := os.Chdir(dir2); err != nil {
		t.Fatal(err)
	}
	oldTerm := terminalCheck
	t.Cleanup(func() { terminalCheck = oldTerm })
	terminalCheck = func(io.Reader) bool { return true }
	out, err = executeIn(t, strings.NewReader("2\ngrilling\n2\n"), "init")
	if err != nil {
		t.Fatalf("interactive: %v\n%s", err, out)
	}
	cfg, err = config.Load(filepath.Join(dir2, config.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Skills.Scope != config.ScopeLocal || !cfg.Skills.Has("grilling") {
		t.Fatalf("interactive cfg %+v\n%s", cfg.Skills, out)
	}
}

func TestInitCopyFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, "init", "--skills", "grilling", "--scope", "local", "--copy")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	local := filepath.Join(dir, ".agents", "skills", "grilling")
	st, err := os.Lstat(local)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode()&os.ModeSymlink != 0 {
		t.Fatal("expected copy, got symlink")
	}
}

func TestInitAgentsMDIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	pre := "# Existing agents file\n\nFollow repo rules.\n"
	if err := os.WriteFile("AGENTS.md", []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, "init", "--skills", "grilling", "--scope", "global", "--agents-md")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	first, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	s := string(first)
	if !strings.HasPrefix(s, "# Existing agents file") {
		t.Fatalf("overwrote AGENTS.md:\n%s", s)
	}
	if !strings.Contains(s, agentsMarkBeg) || !strings.Contains(s, "atlas help --agent") {
		t.Fatalf("missing marked block:\n%s", s)
	}
	// atlas.yaml exists; refresh block via upsert only
	action, err := upsertAgentsMD("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if action != "unchanged" && action != "updated" {
		t.Fatalf("action %s", action)
	}
	second, _ := os.ReadFile("AGENTS.md")
	if strings.Count(string(second), agentsMarkBeg) != 1 {
		t.Fatalf("duplicate blocks:\n%s", second)
	}
}
