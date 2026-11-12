package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stubGraphify(t *testing.T, failRepo string) *[][]string {
	t.Helper()
	var calls [][]string
	oldRun := GraphifyRun
	oldProg := Progress
	t.Cleanup(func() {
		GraphifyRun = oldRun
		Progress = oldProg
	})
	GraphifyRun = func(args ...string) error {
		cp := append([]string{}, args...)
		calls = append(calls, cp)
		if len(args) > 0 && args[0] == "extract" {
			var outDir string
			for i, a := range args {
				if a == "--out" && i+1 < len(args) {
					outDir = args[i+1]
				}
			}
			if failRepo != "" && strings.Contains(outDir, filepath.Join("extract", failRepo)) {
				return errors.New("extract failed")
			}
			d := filepath.Join(outDir, "graphify-out")
			if err := os.MkdirAll(d, 0o755); err != nil {
				return err
			}
			body, _ := json.Marshal(map[string]any{
				"nodes": []map[string]string{{"id": "svc"}},
				"links": []any{},
			})
			return os.WriteFile(filepath.Join(d, "graph.json"), body, 0o644)
		}
		if len(args) > 0 && args[0] == "merge-graphs" {
			var files []string
			outp := ""
			for i := 1; i < len(args); i++ {
				if args[i] == "--out" && i+1 < len(args) {
					outp = args[i+1]
					i++
					continue
				}
				files = append(files, args[i])
			}
			var nodes []map[string]string
			for _, f := range files {
				raw, err := os.ReadFile(f)
				if err != nil {
					return err
				}
				tag := filepath.Base(filepath.Dir(filepath.Dir(f))) // extract/<repo>/graphify-out
				var data struct {
					Nodes []map[string]string `json:"nodes"`
				}
				if err := json.Unmarshal(raw, &data); err != nil {
					return err
				}
				for _, n := range data.Nodes {
					n["id"] = tag + ":" + n["id"]
					nodes = append(nodes, n)
				}
			}
			if err := os.MkdirAll(filepath.Dir(outp), 0o755); err != nil {
				return err
			}
			body, _ := json.Marshal(map[string]any{"nodes": nodes, "links": []any{}})
			return os.WriteFile(outp, body, 0o644)
		}
		return nil
	}
	return &calls
}

func TestBuildGraphUsesExtractOutAndMerge(t *testing.T) {
	home := withHome(t)
	cart := t.TempDir()
	pay := t.TempDir()
	p := &Project{Name: "shop", Repos: []Repo{
		{Name: "cart-service", Path: cart},
		{Name: "payment-service", Path: pay},
	}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}

	calls := stubGraphify(t, "")
	var prog bytes.Buffer
	Progress = &prog

	loaded, _ := Load("shop")
	res, err := BuildGraph(loaded, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 3 {
		t.Fatalf("calls %v", *calls)
	}
	if (*calls)[0][0] != "extract" || !strings.Contains((*calls)[0][len((*calls)[0])-1], filepath.Join("extract", "cart-service")) && !containsArg((*calls)[0], "--out") {
		t.Fatalf("extract cart: %v", (*calls)[0])
	}
	if (*calls)[2][0] != "merge-graphs" {
		t.Fatalf("expected merge-graphs, got %v", (*calls)[2])
	}
	got := prog.String()
	for _, line := range []string{"extract 1/2 cart-service", "extract 2/2 payment-service", "merge 2 graphs"} {
		if !strings.Contains(got, line) {
			t.Fatalf("progress missing %q:\n%s", line, got)
		}
	}
	if !strings.HasSuffix(res.GraphJSON, filepath.Join("projects", "shop", "graphify-out", "graph.json")) {
		t.Fatalf("graph json %s home %s", res.GraphJSON, home)
	}
	raw, err := os.ReadFile(res.GraphJSON)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "cart-service:svc") || !strings.Contains(s, "payment-service:svc") {
		t.Fatalf("prefixed merge: %s", s)
	}
	// member repos must not gain graphify-out
	if _, err := os.Stat(filepath.Join(cart, "graphify-out")); !os.IsNotExist(err) {
		t.Fatal("wrote graphify-out inside member repo")
	}
}

func TestBuildGraphExtractFailureSkipsMerge(t *testing.T) {
	_ = withHome(t)
	cart := t.TempDir()
	pay := t.TempDir()
	p := &Project{Name: "shop", Repos: []Repo{
		{Name: "cart-service", Path: cart},
		{Name: "payment-service", Path: pay},
	}}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}

	calls := stubGraphify(t, "payment-service")
	Progress = io.Discard

	loaded, _ := Load("shop")
	_, err := BuildGraph(loaded, nil, false)
	if err == nil {
		t.Fatal("expected extract error")
	}
	if !strings.Contains(err.Error(), "extract payment-service (2/2)") {
		t.Fatalf("error %v", err)
	}
	for _, c := range *calls {
		if len(c) > 0 && c[0] == "merge-graphs" {
			t.Fatalf("merge should not run: %v", *calls)
		}
	}
	if len(*calls) != 2 {
		t.Fatalf("want 2 extracts, got %v", *calls)
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
