package project

import (
	"os"
	"path/filepath"
	"testing"
)

func withHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("ATLAS_HOME", home)
	return home
}

func TestValidateAndSaveLoad(t *testing.T) {
	withHome(t)
	if err := ValidateName("Shop"); err == nil {
		t.Fatal("expected reject uppercase")
	}
	if err := ValidateName("shop"); err != nil {
		t.Fatal(err)
	}
	p := &Project{
		Name:        "shop",
		Description: "Example shop",
		Repos: []Repo{
			{Name: "cart-service", URL: "https://github.com/example/cart-service.git", Path: "~/src/cart-service"},
			{Name: "payment-service", URL: "https://github.com/example/payment-service.git"},
		},
	}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := Load("shop")
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion || got.Description != "Example shop" || len(got.Repos) != 2 {
		t.Fatalf("%+v", got)
	}
	if err := got.SetRepo(Repo{Name: "cart-service", Path: "/tmp/cart"}); err != nil {
		t.Fatal(err)
	}
	if err := got.SetRepo(Repo{Name: "cart-service", Path: "/tmp/other"}); err != nil {
		t.Fatal(err)
	}
	if n := len(got.Repos); n != 2 {
		t.Fatalf("duplicate add grew list: %d", n)
	}
	if err := Save(got); err != nil {
		t.Fatal(err)
	}
	if err := got.SetRepo(Repo{Name: "cart-service"}); err != nil { // still unique
		t.Fatal(err)
	}
	dup := &Project{Name: "shop", Repos: []Repo{{Name: "a"}, {Name: "a"}}}
	dup.Repos[0].Name = "cart-service"
	dup.Repos[1].Name = "cart-service"
	if err := dup.Validate(); err == nil {
		t.Fatal("expected duplicate repo names to fail")
	}
	if err := got.RemoveRepo("payment-service"); err != nil {
		t.Fatal(err)
	}
	list, err := List()
	if err != nil || len(list) != 1 {
		t.Fatalf("list %v %v", list, err)
	}
	if err := Delete("shop"); err != nil {
		t.Fatal(err)
	}
	if _, err := Load("shop"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestExpandAndGraphPaths(t *testing.T) {
	home := withHome(t)
	p := &Project{Name: "shop", Repos: []Repo{{Name: "cart-service", Path: "~/src/cart"}}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, _ := Load("shop")
	abs := loaded.AbsRepoPath(loaded.Repos[0])
	if abs == "~/src/cart" {
		t.Fatal("expected ~ expand")
	}
	g, err := loaded.GraphJSON()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "projects", "shop", "graphify-out", "graph.json")
	if g != want {
		t.Fatalf("graph %s want %s", g, want)
	}
	out, err := loaded.ExtractOutDir("cart-service")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(out) != "cart-service" {
		t.Fatalf("extract out %s", out)
	}
	_ = os.Getenv
}

func TestPrefixNodeIDs(t *testing.T) {
	in := []byte(`{"nodes":[{"id":"foo"}],"links":[{"source":"foo","target":"bar"}]}`)
	out := prefixNodeIDs(in, "cart-service")
	s := string(out)
	if !containsAll(s, "cart-service:foo", "cart-service:bar") {
		t.Fatalf("%s", s)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !stringContains(s, p) {
			return false
		}
	}
	return true
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || search(s, sub))
}

func search(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestTokenSecrets(t *testing.T) {
	withHome(t)
	p := &Project{Name: "shop"}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	if TokenSet("shop") {
		t.Fatal("no token yet")
	}
	if err := WriteToken("shop", "secret-token"); err != nil {
		t.Fatal(err)
	}
	path, err := SecretsPath("shop")
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("perm %o", st.Mode().Perm())
	}
	tok, err := ReadToken("shop")
	if err != nil || tok != "secret-token" {
		t.Fatalf("token %q err %v", tok, err)
	}
	if !TokenSet("shop") {
		t.Fatal("expected token set")
	}
	if err := DeleteToken("shop"); err != nil {
		t.Fatal(err)
	}
	if TokenSet("shop") {
		t.Fatal("token should be gone")
	}
}
