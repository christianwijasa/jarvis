package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
)

func writeOld(t *testing.T, dir, pin string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if pin != "" {
		if err := os.WriteFile(filepath.Join(dir, pinFile), []byte(pin), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMigrateRenamedRemovesOnlyJarvisFolders(t *testing.T) {
	old, cur := "trd-pr-planning", "trd-planning"
	if n, ok := CurrentName(old); !ok || n != cur {
		t.Fatalf("CurrentName(%s) = %s %v", old, n, ok)
	}
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	root := t.TempDir()
	cfg := &config.Config{Skills: config.Skills{Scope: config.ScopeBoth, Installed: []string{"grilling", old}}}
	cfg.Skills.Normalize()

	global := filepath.Join(home, "skills", old)
	writeOld(t, global, "name="+old+"\nmode=original\nsource=bundled\n")
	localRoot := filepath.Join(root, config.DefaultLocalDir)
	if err := os.MkdirAll(localRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(localRoot, old)
	if err := os.Symlink(global, link); err != nil {
		t.Fatal(err)
	}

	m, err := MigrateRenamed(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if pathExists(global) || pathExists(link) {
		t.Fatalf("old Jarvis folder or link left behind: %+v", m)
	}
	if len(m.Removed) != 2 || len(m.Kept) != 0 {
		t.Fatalf("migration %+v", m)
	}
	if len(m.Install) != 1 || m.Install[0] != cur {
		t.Fatalf("install %v", m.Install)
	}
	if cfg.Skills.Has(old) {
		t.Fatalf("old name still in skills.installed: %v", cfg.Skills.Installed)
	}

	// Folders Jarvis did not install are kept: no pin, a pin for another
	// skill, or a link outside JARVIS_HOME.
	writeOld(t, global, "")
	elsewhere := t.TempDir()
	if err := os.Symlink(elsewhere, link); err != nil {
		t.Fatal(err)
	}
	m, err = MigrateRenamed(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(global, "SKILL.md")) || !isSymlink(link) {
		t.Fatal("foreign folder or link was removed")
	}
	if len(m.Removed) != 0 || len(m.Kept) != 2 || len(m.Install) != 0 {
		t.Fatalf("migration %+v", m)
	}

	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	writeOld(t, link, "name=other\n")
	if _, err := MigrateRenamed(cfg, root); err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(link, "SKILL.md")) {
		t.Fatal("local copy pinned to another skill was removed")
	}

	// A Jarvis copy (pin present) in the local dir is removed.
	if err := os.WriteFile(filepath.Join(link, pinFile), []byte("name="+old+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err = MigrateRenamed(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if pathExists(link) || !strings.Contains(strings.Join(m.Removed, " "), link) {
		t.Fatalf("Jarvis local copy not removed: %+v", m)
	}
}

func pathExists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}
