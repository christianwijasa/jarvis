package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyAlreadyLatestSkipsDownload(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "atlas")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	c, calls := fakeGH(t, "v0.1.0", map[string][]byte{"atlas-darwin-arm64": []byte("newbin")})
	res, err := c.Apply(Options{
		Repo:    "example/atlas",
		Current: "v0.1.0",
		Dest:    dest,
		GOOS:    "darwin",
		GOARCH:  "arm64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated || res.Latest != "v0.1.0" {
		t.Fatalf("%+v", res)
	}
	body, _ := os.ReadFile(dest)
	if string(body) != "old" {
		t.Fatal("replaced binary when already latest")
	}
	for _, c := range *calls {
		if contains(c, "download") {
			t.Fatalf("download should not run: %v", *calls)
		}
	}
}

func TestApplyDevUpgradesEvenIfTagLooksSame(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "atlas")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("newbin")
	sum := sha256.Sum256(payload)
	sums := hex.EncodeToString(sum[:]) + "  atlas-darwin-arm64\n"
	c, _ := fakeGH(t, "v0.1.0", map[string][]byte{
		"atlas-darwin-arm64": payload,
		"SHA256SUMS":         []byte(sums),
	})
	res, err := c.Apply(Options{
		Current: "dev",
		Dest:    dest,
		GOOS:    "darwin",
		GOARCH:  "arm64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated || res.Latest != "v0.1.0" {
		t.Fatalf("%+v", res)
	}
	body, _ := os.ReadFile(dest)
	if string(body) != "newbin" {
		t.Fatalf("got %q", body)
	}
}

func TestApplyProgressOmitsResultLine(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "atlas")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	c, _ := fakeGH(t, "v0.1.3", map[string][]byte{"atlas-darwin-arm64": []byte("newbin")})
	var prog strings.Builder
	res, err := c.Apply(Options{
		Current:  "dev",
		Dest:     dest,
		GOOS:     "darwin",
		GOARCH:   "arm64",
		Progress: &prog,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated {
		t.Fatalf("%+v", res)
	}
	got := prog.String()
	if strings.Contains(got, "installed") {
		t.Fatalf("progress should not print installed (cmd owns that): %q", got)
	}
	if !strings.Contains(got, "downloading") {
		t.Fatalf("expected download progress, got %q", got)
	}
}

func TestApplyCheckDoesNotWrite(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "atlas")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	c, calls := fakeGH(t, "v0.2.0", map[string][]byte{"atlas-linux-amd64": []byte("x")})
	res, err := c.Apply(Options{
		Check:   true,
		Current: "dev",
		Dest:    dest,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated || res.Latest != "v0.2.0" || !res.Check {
		t.Fatalf("%+v", res)
	}
	body, _ := os.ReadFile(dest)
	if string(body) != "old" {
		t.Fatal("check wrote dest")
	}
	for _, c := range *calls {
		if contains(c, "download") {
			t.Fatalf("download on --check: %v", *calls)
		}
	}
}

func TestApplyMissingAsset(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "atlas")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	c, _ := fakeGH(t, "v0.1.0", map[string][]byte{"atlas-linux-amd64": []byte("x")})
	_, err := c.Apply(Options{Current: "dev", Dest: dest, GOOS: "darwin", GOARCH: "arm64"})
	if err == nil || !strings.Contains(err.Error(), "no asset atlas-darwin-arm64") {
		t.Fatalf("err %v", err)
	}
}

func TestApplyChecksumMismatch(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "atlas")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	c, _ := fakeGH(t, "v0.1.0", map[string][]byte{
		"atlas-darwin-arm64": []byte("newbin"),
		"SHA256SUMS":         []byte("0000000000000000000000000000000000000000000000000000000000000000  atlas-darwin-arm64\n"),
	})
	_, err := c.Apply(Options{Current: "dev", Dest: dest, GOOS: "darwin", GOARCH: "arm64"})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err %v", err)
	}
	body, _ := os.ReadFile(dest)
	if string(body) != "old" {
		t.Fatal("mismatch still replaced dest")
	}
}

func TestApplyNoReleases(t *testing.T) {
	c := Client{
		LookPath: func(string) (string, error) { return "/bin/gh", nil },
		Run: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("HTTP 404")
		},
	}
	_, err := c.Apply(Options{Current: "dev", Dest: filepath.Join(t.TempDir(), "atlas")})
	if err == nil || !strings.Contains(err.Error(), "no GitHub Releases") {
		t.Fatalf("err %v", err)
	}
}

func fakeGH(t *testing.T, tag string, files map[string][]byte) (Client, *[][]string) {
	t.Helper()
	var calls [][]string
	var assets []map[string]string
	for name := range files {
		assets = append(assets, map[string]string{"name": name})
	}
	view, err := json.Marshal(map[string]any{"tagName": tag, "assets": assets})
	if err != nil {
		t.Fatal(err)
	}
	c := Client{
		LookPath: func(string) (string, error) { return "/bin/gh", nil },
		Run: func(name string, args ...string) ([]byte, error) {
			cp := append([]string{}, args...)
			calls = append(calls, cp)
			if len(args) >= 2 && args[0] == "release" && args[1] == "view" {
				return view, nil
			}
			if len(args) >= 2 && args[0] == "release" && args[1] == "download" {
				dir := flagValue(args, "--dir")
				pat := flagValue(args, "--pattern")
				body, ok := files[pat]
				if !ok {
					return nil, fmt.Errorf("no asset %s", pat)
				}
				if err := os.WriteFile(filepath.Join(dir, pat), body, 0o644); err != nil {
					return nil, err
				}
				return nil, nil
			}
			return nil, fmt.Errorf("unexpected gh %v", args)
		},
	}
	return c, &calls
}

func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func contains(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
