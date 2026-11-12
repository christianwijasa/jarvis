package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/upstream"
)

func TestAssertPinned(t *testing.T) {
	for _, ref := range []string{"HEAD", "main", "master", "", "unknown"} {
		if err := AssertPinned(ref); err == nil {
			t.Errorf("expected refusal of %q", ref)
		}
	}
	if err := AssertPinned("208acba8d412335185b2056fedf54b3405df43fb"); err != nil {
		t.Fatal(err)
	}
}

func TestGitFetchUsesPinNotMain(t *testing.T) {
	pin := "208acba8d412335185b2056fedf54b3405df43fb"
	f := DefaultFetcher{
		LookPath: func(string) (string, error) { return "", errors.New("no gh") },
		Run: func(dir, name string, args ...string) ([]byte, error) {
			joined := strings.Join(args, " ")
			for _, a := range args {
				switch a {
				case "HEAD", "main", "master", "origin/main":
					t.Fatalf("fetch used floating ref: git %s", joined)
				}
			}
			if name == "git" && len(args) > 0 && args[0] == "fetch" {
				if args[len(args)-1] != pin {
					t.Fatalf("fetch pin %v", args)
				}
			}
			if name == "git" && len(args) > 0 && args[0] == "checkout" {
				if err := os.MkdirAll(filepath.Join(dir, "archify"), 0o755); err != nil {
					return nil, err
				}
				return nil, os.WriteFile(filepath.Join(dir, "archify", "SKILL.md"), []byte("# Archify\n"), 0o644)
			}
			return nil, nil
		},
	}
	files, err := f.FetchTree("https://github.com/tt-a1i/archify", pin, "archify")
	if err != nil {
		t.Fatal(err)
	}
	if string(files["SKILL.md"]) != "# Archify\n" {
		t.Fatalf("files=%v", files)
	}
}

func TestInstallBundledAndVendored(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	project := t.TempDir()
	cfg := &config.Config{Skills: config.Skills{Scope: config.ScopeBoth, LocalDir: ".agents/skills"}}
	cfg.Skills.Normalize()

	reg, err := upstream.Load()
	if err != nil {
		t.Fatal(err)
	}
	grilling, ok := reg.Skill("grilling")
	if !ok {
		t.Fatal("missing grilling")
	}
	res, err := Install(cfg, project, grilling, FuncFetcher(func(string, string, string) (map[string][]byte, error) {
		t.Fatal("bundled skill should not fetch")
		return nil, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != "bundled" || !fileExists(filepath.Join(res.Global, "SKILL.md")) {
		t.Fatalf("bundled install %+v", res)
	}
	if res.Local == "" || !isSymlink(res.Local) {
		t.Fatalf("expected symlink, got %+v", res)
	}

	var fetchedCommit string
	archify, ok := reg.Skill("archify")
	if !ok {
		t.Fatal("missing archify")
	}
	res, err = Install(cfg, project, archify, FuncFetcher(func(repo, commit, path string) (map[string][]byte, error) {
		fetchedCommit = commit
		if repo != "https://github.com/tt-a1i/archify" || path != "archify/" {
			t.Fatalf("repo=%s path=%s", repo, path)
		}
		if err := AssertPinned(commit); err != nil {
			return nil, err
		}
		return map[string][]byte{"SKILL.md": []byte("# Archify fetched\n")}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if fetchedCommit != "208acba8d412335185b2056fedf54b3405df43fb" {
		t.Fatalf("pin %s", fetchedCommit)
	}
	body, _ := os.ReadFile(filepath.Join(home, "skills", "archify", "SKILL.md"))
	if string(body) != "# Archify fetched\n" {
		t.Fatalf("body %q", body)
	}

	cfg.Skills.Copy = true
	res, err = Install(cfg, project, grilling, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Copied || isSymlink(res.Local) {
		t.Fatalf("expected copy %+v", res)
	}

	if err := Remove(cfg, project, "grilling"); err != nil {
		t.Fatal(err)
	}
	if fileExists(filepath.Join(home, "skills", "grilling", "SKILL.md")) {
		t.Fatal("grilling still present")
	}
}

func TestCheckLinksBroken(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JARVIS_HOME", home)
	project := t.TempDir()
	cfg := &config.Config{Skills: config.Skills{
		Scope:     config.ScopeBoth,
		LocalDir:  ".agents/skills",
		Installed: []string{"grilling"},
	}}
	problems := CheckLinks(cfg, project)
	if len(problems) == 0 {
		t.Fatal("expected missing install")
	}
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func isSymlink(p string) bool {
	st, err := os.Lstat(p)
	return err == nil && st.Mode()&os.ModeSymlink != 0
}
