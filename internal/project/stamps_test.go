package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupClone(t *testing.T, branch string) (checkout, bare, sha string) {
	t.Helper()
	src := t.TempDir()
	gitOk(t, src, "init", "-b", branch)
	if err := os.WriteFile(filepath.Join(src, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOk(t, src, "add", "README")
	gitOk(t, src, "commit", "-m", "init")
	bare = filepath.Join(t.TempDir(), "repo.git")
	gitOk(t, "", "clone", "--bare", src, bare)
	checkout = t.TempDir()
	gitOk(t, "", "clone", bare, checkout)
	b, err := GitRun(checkout, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return checkout, bare, string(b)
}

func TestOriginDefaultMainThenMaster(t *testing.T) {
	checkout, _, sha := setupClone(t, "master")
	ref, got, err := originDefault(checkout, "")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "origin/master" {
		t.Fatalf("ref %s", ref)
	}
	if got != sha {
		t.Fatalf("sha %s want %s", got, sha)
	}

	main, _, mainSHA := setupClone(t, "main")
	ref, got, err = originDefault(main, "")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "origin/main" {
		t.Fatalf("ref %s", ref)
	}
	if got != mainSHA {
		t.Fatalf("sha %s want %s", got, mainSHA)
	}
}

func TestCheckStaleMissingStampsIsStale(t *testing.T) {
	withHome(t)
	checkout, _, sha := setupClone(t, "main")
	p := &Project{Name: "shop", Repos: []Repo{{Name: "cart-service", Path: checkout}}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, _ := Load("shop")
	gjson, _ := loaded.GraphJSON()
	if err := os.MkdirAll(filepath.Dir(gjson), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gjson, []byte(`{"nodes":[{"id":"x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := CheckStale(loaded, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Graph != "stale" {
		t.Fatalf("%+v", rows)
	}
	if rows[0].OriginCommit != sha {
		t.Fatalf("origin %s want %s", rows[0].OriginCommit, sha)
	}
}

func TestStatusUsesSHANotMtime(t *testing.T) {
	withHome(t)
	checkout, _, sha := setupClone(t, "main")
	p := &Project{Name: "shop", Repos: []Repo{{Name: "cart-service", Path: checkout}}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, _ := Load("shop")
	gjson, _ := loaded.GraphJSON()
	if err := os.MkdirAll(filepath.Dir(gjson), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gjson, []byte(`{"nodes":[{"id":"x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Add(time.Hour)
	if err := os.Chtimes(gjson, now, now); err != nil {
		t.Fatal(err)
	}
	stamps := GraphStamps{Repos: map[string]RepoStamp{
		"cart-service": {Commit: "deadbeef", Ref: "origin/main"},
	}}
	if err := WriteStamps(loaded, stamps); err != nil {
		t.Fatal(err)
	}
	rows, err := Status(loaded, false)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Graph != "stale" {
		t.Fatalf("want stale despite newer graph mtime: %+v", rows[0])
	}
	if rows[0].GraphCommit != "deadbeef" || rows[0].OriginCommit != sha {
		t.Fatalf("%+v origin want %s", rows[0], sha)
	}
}

func TestCheckStaleFreshWhenStampMatchesOrigin(t *testing.T) {
	withHome(t)
	checkout, _, sha := setupClone(t, "main")
	p := &Project{Name: "shop", Repos: []Repo{{Name: "cart-service", Path: checkout}}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, _ := Load("shop")
	gjson, _ := loaded.GraphJSON()
	if err := os.MkdirAll(filepath.Dir(gjson), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gjson, []byte(`{"nodes":[{"id":"x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteStamps(loaded, GraphStamps{Repos: map[string]RepoStamp{
		"cart-service": {Commit: sha, Ref: "origin/main"},
	}}); err != nil {
		t.Fatal(err)
	}
	rows, err := CheckStale(loaded, false)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Graph != "fresh" {
		t.Fatalf("%+v", rows[0])
	}
}

func TestWriteReadStampsRoundTrip(t *testing.T) {
	withHome(t)
	p := &Project{Name: "shop"}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, _ := Load("shop")
	in := GraphStamps{Repos: map[string]RepoStamp{
		"cart-service": {Commit: "abc", Ref: "origin/main"},
	}}
	if err := WriteStamps(loaded, in); err != nil {
		t.Fatal(err)
	}
	got, err := ReadStamps(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if got.Repos["cart-service"].Commit != "abc" {
		t.Fatalf("%+v", got)
	}
	raw, err := os.ReadFile(mustStamps(t, loaded))
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["schema_version"] != StampSchema {
		t.Fatalf("%v", parsed)
	}
}

func mustStamps(t *testing.T, p *Project) string {
	t.Helper()
	s, err := p.StampsPath()
	if err != nil {
		t.Fatal(err)
	}
	return s
}
