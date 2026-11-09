package upstream

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type fakeRemote struct {
	sha map[string]string
	err error
}

func (f fakeRemote) LatestCommit(repoURL, path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	key := repoURL + "|" + path
	if sha, ok := f.sha[key]; ok {
		return sha, nil
	}
	if sha, ok := f.sha[repoURL]; ok {
		return sha, nil
	}
	return "", errors.New("no fake sha")
}

func TestCheckStatuses(t *testing.T) {
	reg := &Registry{Entries: []Entry{
		{Name: "local", Mode: ModeOriginal},
		{Name: "archify", Mode: ModeVendored, UpstreamRepo: "https://github.com/tt-a1i/archify", UpstreamPath: "SKILL.md", PinnedCommit: "unknown"},
		{Name: "grilling", Mode: ModeAdapted, UpstreamRepo: "https://github.com/mattpocock/skills", UpstreamPath: "skills/productivity/grilling/SKILL.md", PinnedCommit: "49dd158d1076134a641b33efb035946536778336"},
		{Name: "moved", Mode: ModeAdapted, UpstreamRepo: "https://github.com/mattpocock/skills", UpstreamPath: "skills/engineering/domain-modeling/SKILL.md", PinnedCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}}
	remote := fakeRemote{sha: map[string]string{
		"https://github.com/mattpocock/skills|skills/productivity/grilling/SKILL.md":       "49dd158d1076134a641b33efb035946536778336",
		"https://github.com/mattpocock/skills|skills/engineering/domain-modeling/SKILL.md": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}}
	got := Check(reg, remote)
	want := []string{"original", "unpinned", "unchanged", "changed"}
	if len(got) != len(want) {
		t.Fatalf("len=%d", len(got))
	}
	for i, s := range want {
		if got[i].Status != s {
			t.Errorf("%s: status %q want %q (%s)", got[i].Entry.Name, got[i].Status, s, got[i].Detail)
		}
	}
	if got[3].Compare == "" || !strings.Contains(got[3].Compare, "compare/") {
		t.Errorf("expected compare URL, got %q", got[3].Compare)
	}
}

func TestCheckRemoteError(t *testing.T) {
	reg := &Registry{Entries: []Entry{
		{Name: "grilling", Mode: ModeAdapted, UpstreamRepo: "https://github.com/mattpocock/skills", UpstreamPath: "x", PinnedCommit: "abc1234"},
	}}
	got := Check(reg, fakeRemote{err: errors.New("network down")})
	if got[0].Status != "error" {
		t.Fatalf("status %q", got[0].Status)
	}
}

func TestRootFileMatchesEmbed(t *testing.T) {
	root, err := os.ReadFile("../../upstream.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(root) != string(raw) {
		t.Fatal("upstream.yaml at repo root must match internal/upstream/upstream.yaml")
	}
}

func TestLoadRegistry(t *testing.T) {
	reg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Entries) < 10 {
		t.Fatalf("too few entries: %d", len(reg.Entries))
	}
	var graphify, archify Entry
	for _, e := range reg.Entries {
		switch e.Name {
		case "graphify":
			graphify = e
		case "archify":
			archify = e
		}
	}
	if !graphify.Recommended || graphify.Pack != "recommended" {
		t.Fatalf("graphify pack %q recommended=%v", graphify.Pack, graphify.Recommended)
	}
	if !archify.Recommended || archify.Pack != "recommended" {
		t.Fatalf("archify pack %q recommended=%v", archify.Pack, archify.Recommended)
	}
	if archify.LocalPath != "" {
		t.Fatalf("archify must not store a vendored local_path, got %q", archify.LocalPath)
	}
	if archify.PinnedTag != "v3.2.0" {
		t.Fatalf("archify tag %q", archify.PinnedTag)
	}
	if graphify.Mode != ModeAdapted {
		t.Fatalf("graphify mode %q", graphify.Mode)
	}
	if archify.PinnedCommit != "208acba8d412335185b2056fedf54b3405df43fb" {
		t.Fatalf("archify pin %q", archify.PinnedCommit)
	}
	if archify.Licence != "Apache-2.0" {
		t.Fatalf("archify licence %q", archify.Licence)
	}
}

func TestParseGHCommits(t *testing.T) {
	sha, err := parseGHCommits([]byte(`[{"sha":"abc1234deadbeef"}]`))
	if err != nil || sha != "abc1234deadbeef" {
		t.Fatalf("sha=%q err=%v", sha, err)
	}
}

func TestGHRemoteUsesMockRun(t *testing.T) {
	r := GHRemote{Run: func(name string, args ...string) ([]byte, error) {
		if name == "gh" {
			return []byte(`[{"sha":"from-gh"}]`), nil
		}
		t.Fatalf("unexpected %s %v", name, args)
		return nil, nil
	}}
	sha, err := r.LatestCommit("https://github.com/mattpocock/skills", "skills/productivity/grilling/SKILL.md")
	if err != nil || sha != "from-gh" {
		t.Fatalf("sha=%q err=%v", sha, err)
	}
}

func TestGHRemoteFallsBackToLsRemote(t *testing.T) {
	r := GHRemote{Run: func(name string, args ...string) ([]byte, error) {
		if name == "gh" {
			return nil, errors.New("no gh")
		}
		if name == "git" {
			return []byte("deadbeef\tHEAD\n"), nil
		}
		return nil, errors.New(name)
	}}
	sha, err := r.LatestCommit("https://github.com/mattpocock/skills", "x")
	if err != nil || sha != "deadbeef" {
		t.Fatalf("sha=%q err=%v", sha, err)
	}
}
