package doctor

import (
	"bytes"
	"os"
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
