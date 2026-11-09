package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitOk(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCloneLocalBareAndStatus(t *testing.T) {
	withHome(t)
	src := t.TempDir()
	gitOk(t, src, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(src, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOk(t, src, "add", "README")
	gitOk(t, src, "commit", "-m", "init")
	bare := filepath.Join(t.TempDir(), "cart-service.git")
	gitOk(t, "", "clone", "--bare", src, bare)

	destParent := t.TempDir()
	p := &Project{Name: "shop", Repos: []Repo{
		{Name: "cart-service", URL: bare},
		{Name: "payment-service"},
	}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load("shop")
	if err != nil {
		t.Fatal(err)
	}
	results, err := Clone(loaded, destParent, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("%+v", results)
	}
	if results[0].Status != "cloned" {
		t.Fatalf("clone: %+v", results[0])
	}
	if results[1].Status != "skipped" {
		t.Fatalf("no url should skip: %+v", results[1])
	}
	again, err := Load("shop")
	if err != nil {
		t.Fatal(err)
	}
	results, err = Clone(again, destParent, 1)
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Status != "skipped" {
		t.Fatalf("second clone should skip: %+v", results[0])
	}

	st, err := Status(again, false)
	if err != nil {
		t.Fatal(err)
	}
	if !st[0].Exists || !st[0].GitRepo || st[0].Branch != "main" {
		t.Fatalf("status %+v", st[0])
	}
	if st[0].Graph != "missing" {
		t.Fatalf("no graph yet: %+v", st[0])
	}
}
