package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/install"
	"github.com/christianwijasa/jarvis/internal/project"
)

// Item is one doctor row.
type Item struct {
	Name     string   `json:"name"`
	OK       bool     `json:"ok"`
	Warn     bool     `json:"warn,omitempty"`
	Detail   string   `json:"detail"`
	Problems []string `json:"problems,omitempty"`
}

// Report is the structured doctor result (printed as text or JSON).
type Report struct {
	OK     bool   `json:"ok"`
	Checks []Item `json:"checks"`
}

// Collect runs the checks without printing.
func Collect(startDir string) Report {
	var r Report
	r.Checks = append(r.Checks, toolCheck("git"), toolCheck("gh"), toolCheck("graphify"))

	path, err := config.Find(startDir)
	if err != nil {
		r.Checks = append(r.Checks, Item{
			Name:   "jarvis.yaml",
			OK:     false,
			Detail: "not found in this directory or parents",
		})
		r.OK = allOK(r.Checks)
		return r
	}
	r.Checks = append(r.Checks, Item{Name: "jarvis.yaml", OK: true, Detail: path})

	cfg, err := config.Load(path)
	if err != nil {
		r.Checks = append(r.Checks, Item{Name: "skills", OK: false, Detail: err.Error()})
		r.OK = allOK(r.Checks)
		return r
	}
	home, err := config.HomeDir(cfg)
	if err != nil {
		r.Checks = append(r.Checks, Item{Name: "JARVIS_HOME", OK: false, Detail: err.Error()})
	} else {
		r.Checks = append(r.Checks, Item{Name: "JARVIS_HOME", OK: true, Detail: home})
	}
	problems := install.CheckLinks(cfg, filepath.Dir(path))
	sk := Item{
		Name:   "skills",
		OK:     len(problems) == 0,
		Detail: fmt.Sprintf("%d installed, scope %s", len(cfg.Skills.Installed), cfg.Skills.Scope),
	}
	if len(problems) > 0 {
		sk.Problems = problems
		sk.Detail = "links invalid"
	}
	r.Checks = append(r.Checks, sk)
	r.Checks = append(r.Checks, projectCheck(cfg))
	r.OK = allOK(r.Checks)
	return r
}

func projectCheck(cfg *config.Config) Item {
	name := strings.TrimSpace(cfg.Project)
	if name == "" {
		return Item{Name: "project", OK: true, Detail: "none set"}
	}
	p, err := project.Load(name)
	if err != nil {
		return Item{Name: "project", OK: false, Detail: err.Error()}
	}
	var missing []string
	for _, r := range p.Repos {
		path := p.AbsRepoPath(r)
		if path == "" {
			continue
		}
		if st, err := os.Stat(path); err != nil || !st.IsDir() {
			missing = append(missing, r.Name+" ("+path+")")
		}
	}
	item := Item{Name: "project", OK: true, Detail: name}
	if len(missing) > 0 {
		item.Warn = true
		item.Problems = missing
		item.Detail = name + "; missing paths (warn)"
	}
	return item
}

// Check prints a text report. It returns an error when any row failed.
func Check(out io.Writer, startDir string) error {
	r := Collect(startDir)
	WriteText(out, r)
	if !r.OK {
		return fmt.Errorf("doctor found problems")
	}
	return nil
}

func WriteText(out io.Writer, r Report) {
	for _, c := range r.Checks {
		status := "fail"
		if c.OK && c.Warn {
			status = "warn"
		} else if c.OK {
			status = "pass"
		}
		if c.Detail != "" {
			fmt.Fprintf(out, "%s: %s (%s)\n", c.Name, status, c.Detail)
		} else {
			fmt.Fprintf(out, "%s: %s\n", c.Name, status)
		}
		for _, p := range c.Problems {
			fmt.Fprintf(out, "  %s\n", p)
		}
	}
}

func toolCheck(name string) Item {
	path, err := exec.LookPath(name)
	if err != nil {
		return Item{Name: name, OK: false, Detail: "not on PATH"}
	}
	return Item{Name: name, OK: true, Detail: path}
}

func allOK(checks []Item) bool {
	for _, c := range checks {
		if !c.OK {
			return false
		}
	}
	return true
}
