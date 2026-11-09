package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	var calls [][]string
	oldRun := GraphifyRun
	t.Cleanup(func() { GraphifyRun = oldRun })
	GraphifyRun = func(args ...string) ([]byte, error) {
		cp := append([]string{}, args...)
		calls = append(calls, cp)
		if len(args) > 0 && args[0] == "extract" {
			var outDir string
			for i, a := range args {
				if a == "--out" && i+1 < len(args) {
					outDir = args[i+1]
				}
			}
			d := filepath.Join(outDir, "graphify-out")
			if err := os.MkdirAll(d, 0o755); err != nil {
				return nil, err
			}
			body, _ := json.Marshal(map[string]any{
				"nodes": []map[string]string{{"id": "svc"}},
				"links": []any{},
			})
			return nil, os.WriteFile(filepath.Join(d, "graph.json"), body, 0o644)
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
					return nil, err
				}
				tag := filepath.Base(filepath.Dir(filepath.Dir(f))) // extract/<repo>/graphify-out
				var data struct {
					Nodes []map[string]string `json:"nodes"`
				}
				if err := json.Unmarshal(raw, &data); err != nil {
					return nil, err
				}
				for _, n := range data.Nodes {
					n["id"] = tag + ":" + n["id"]
					nodes = append(nodes, n)
				}
			}
			if err := os.MkdirAll(filepath.Dir(outp), 0o755); err != nil {
				return nil, err
			}
			body, _ := json.Marshal(map[string]any{"nodes": nodes, "links": []any{}})
			return nil, os.WriteFile(outp, body, 0o644)
		}
		return nil, nil
	}

	loaded, _ := Load("shop")
	res, err := BuildGraph(loaded, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatalf("calls %v", calls)
	}
	if calls[0][0] != "extract" || !strings.Contains(calls[0][len(calls[0])-1], filepath.Join("extract", "cart-service")) && !containsArg(calls[0], "--out") {
		t.Fatalf("extract cart: %v", calls[0])
	}
	if calls[2][0] != "merge-graphs" {
		t.Fatalf("expected merge-graphs, got %v", calls[2])
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

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
