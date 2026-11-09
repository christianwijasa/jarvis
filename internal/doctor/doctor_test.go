package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasabfi/atlas/internal/config"
)

func TestCheckWithConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte("jira: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Check(&buf, dir)
	out := buf.String()
	if !strings.Contains(out, "git:") || !strings.Contains(out, "gh:") || !strings.Contains(out, "graphify:") || !strings.Contains(out, "atlas.yaml: pass") {
		t.Fatalf("unexpected doctor output:\n%s", out)
	}
}

func TestCheckWithoutConfig(t *testing.T) {
	var buf bytes.Buffer
	err := Check(&buf, t.TempDir())
	if err == nil {
		t.Fatal("expected error when atlas.yaml is missing")
	}
	if !strings.Contains(buf.String(), "atlas.yaml: fail") {
		t.Fatalf("unexpected doctor output:\n%s", buf.String())
	}
}

func TestCheckSkillLinks(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
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
