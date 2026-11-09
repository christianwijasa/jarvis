package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWalksParents(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, FileName), []byte("jira:\n  site: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Find(child)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	want := filepath.Join(root, FileName)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFindMissing(t *testing.T) {
	_, err := Find(t.TempDir())
	if err != ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestLoadExample(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, ExampleYAML(), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Jira.Site == "" {
		t.Fatal("expected jira.site in example config")
	}
	if cfg.History.Path == "" {
		t.Fatal("expected history.path in example config")
	}
	if len(cfg.Repos) == 0 {
		t.Fatal("expected repos map in example config")
	}
	if !cfg.Gates.TestCase {
		t.Fatal("expected gates.test_case in example config")
	}
	if cfg.Skills.Scope != ScopeBoth {
		t.Fatalf("skills.scope %q", cfg.Skills.Scope)
	}
}

func TestHomeDirEnvAndWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
	got, err := HomeDir(nil)
	if err != nil || got != home {
		t.Fatalf("HomeDir=%q err=%v", got, err)
	}
	path := filepath.Join(t.TempDir(), FileName)
	cfg := &Config{Skills: Skills{Scope: ScopeLocal, Installed: []string{"grilling", "grilling"}}}
	if err := Write(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Skills.Scope != ScopeLocal || len(loaded.Skills.Installed) != 1 {
		t.Fatalf("%+v", loaded.Skills)
	}
}
