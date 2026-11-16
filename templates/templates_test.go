package templates

import (
	"io/fs"
	"strings"
	"testing"
	"text/template"
)

type sample struct {
	Issue       string
	TrackerKind string
	Tier        string
	TestSource  string
	ReposCSV    string
	SourceNotes string
}

// The documents land in a team's history repo. They must not name the tool.
func TestTemplatesRenderWithoutToolName(t *testing.T) {
	names, err := fs.Glob(FS, "*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 4 {
		t.Fatalf("templates: %v", names)
	}
	data := sample{Issue: "TEAM-1", TrackerKind: "jira", Tier: "A", TestSource: "tracker:TEAM-1", ReposCSV: "api", SourceNotes: "notes"}
	for _, n := range names {
		raw, err := FS.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(string(raw)), "jarvis") {
			t.Errorf("%s template text names the tool", n)
		}
		tmpl, err := template.New(n).Parse(string(raw))
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		var b strings.Builder
		if err := tmpl.Execute(&b, data); err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		if !strings.Contains(b.String(), "TEAM-1") {
			t.Errorf("%s does not render the issue key", n)
		}
	}
}

func TestPRDTemplateSections(t *testing.T) {
	raw, err := FS.ReadFile("PRD.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"issue: {{.Issue}}", "type: prd", "## 7. Acceptance criteria", "## 13. Engineering checks", "TRD link"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("PRD template missing %q", want)
		}
	}
}
