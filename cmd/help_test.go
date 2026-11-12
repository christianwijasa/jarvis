package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/christianwijasabfi/atlas/internal/buildinfo"
)

func TestHumanHelpHasPurposeUsageFlagsExamples(t *testing.T) {
	root, err := execute(t, "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, need := range []string{"Atlas is a small", "Usage:", "Available Commands:", "atlas help --agent", "Examples:"} {
		if !strings.Contains(root, need) {
			t.Errorf("root help missing %q\n%s", need, root)
		}
	}

	initHelp, err := execute(t, "help", "init")
	if err != nil {
		t.Fatal(err)
	}
	for _, need := range []string{"Usage:", "init", "--yes", "--recommended", "--skills", "--scope", "--agents-md", "Examples:"} {
		if !strings.Contains(initHelp, need) {
			t.Errorf("init help missing %q\n%s", need, initHelp)
		}
	}

	skillsHelp, err := execute(t, "skills", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, need := range []string{"Usage:", "show", "install", "remove", "--json", "Examples:"} {
		if !strings.Contains(skillsHelp, need) {
			t.Errorf("skills help missing %q\n%s", need, skillsHelp)
		}
	}
}

func TestAgentHelpGolden(t *testing.T) {
	want, err := os.ReadFile("testdata/agent-help.md")
	if err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, "help", "--agent")
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("help --agent != testdata/agent-help.md\ngot %d bytes want %d", len(out), len(want))
	}
	alias, err := execute(t, "agent-help")
	if err != nil {
		t.Fatal(err)
	}
	if alias != out {
		t.Fatal("agent-help should match help --agent")
	}
}

func TestAgentHelpJSONGolden(t *testing.T) {
	out, err := execute(t, "help", "--agent", "--json")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/agent-help.json")
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("JSON golden mismatch.\ngot:\n%s\nwant:\n%s", out, want)
	}
	var got agentHelpJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion || got.Kind != "agent_help" {
		t.Fatalf("schema %q kind %q", got.SchemaVersion, got.Kind)
	}
	if !strings.Contains(got.Markdown, "atlas skills show") {
		t.Fatal("markdown missing load-skill instruction")
	}
	if got.LoadSkill != "atlas skills show <name>" {
		t.Fatalf("load_skill %q", got.LoadSkill)
	}
}

func TestHelpJSONRequiresAgent(t *testing.T) {
	_, err := execute(t, "help", "--json")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVersionJSONSchema(t *testing.T) {
	out, err := execute(t, "version", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["schema_version"] != SchemaVersion || got["kind"] != "version" {
		t.Fatalf("%v", got)
	}
	if got["cli_version"] != buildinfo.Version || got["skill_pack"] != "0.2.0" {
		t.Fatalf("%v", got)
	}
	if _, ok := got["config_file"]; !ok {
		t.Fatal("missing config_file")
	}
}

func TestSkillsJSONSchema(t *testing.T) {
	t.Setenv("ATLAS_HOME", t.TempDir())
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	out, err := execute(t, "skills", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["schema_version"] != SchemaVersion || got["kind"] != "skills" {
		t.Fatalf("%v", got)
	}
	avail, ok := got["available"].([]any)
	if !ok || len(avail) < 10 {
		t.Fatalf("available: %v", got["available"])
	}
	show, err := execute(t, "skills", "show", "grilling", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var skill map[string]any
	if err := json.Unmarshal([]byte(show), &skill); err != nil {
		t.Fatal(err)
	}
	if skill["schema_version"] != SchemaVersion || skill["kind"] != "skill" || skill["name"] != "grilling" {
		t.Fatalf("%v", skill)
	}
	body, _ := skill["body"].(string)
	if body == "" {
		t.Fatal("empty skill body")
	}
}

func TestDoctorJSONSchema(t *testing.T) {
	out, _ := execute(t, "doctor", "--json")
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Kind          string `json:"kind"`
		OK            bool   `json:"ok"`
		Checks        []struct {
			Name string `json:"name"`
			OK   bool   `json:"ok"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion || got.Kind != "doctor" {
		t.Fatalf("%+v", got)
	}
	names := map[string]bool{}
	for _, c := range got.Checks {
		names[c.Name] = true
	}
	for _, n := range []string{"git", "gh", "graphify", "atlas.yaml"} {
		if !names[n] {
			t.Errorf("missing check %s in %+v", n, got.Checks)
		}
	}
}
