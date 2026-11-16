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
	"github.com/christianwijasa/jarvis/internal/role"
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
	OK         bool   `json:"ok"`
	Role       string `json:"role"`
	RoleSource string `json:"role_source"`
	Checks     []Item `json:"checks"`
}

// Collect runs the checks without printing. It loads jarvis.yaml and the
// active project's role first, then picks the checks for that role.
func Collect(startDir string) Report {
	var r Report
	path, findErr := config.Find(startDir)
	var cfg *config.Config
	var loadErr error
	if findErr == nil {
		cfg, loadErr = config.Load(path)
	}
	res := role.Resolve(cfg)
	r.Role = res.Role
	r.RoleSource = res.Source
	r.Checks = append(r.Checks, roleCheck(res))
	product := res.Role == role.Product

	r.Checks = append(r.Checks, toolCheck("git"))
	if product {
		gh := toolCheck("gh")
		if !gh.OK {
			gh.OK, gh.Warn = true, true
			gh.Detail = "not on PATH (warn; the product role does not need it)"
		}
		r.Checks = append(r.Checks, gh)
	} else {
		r.Checks = append(r.Checks, toolCheck("gh"), toolCheck("graphify"))
	}

	if findErr != nil {
		r.Checks = append(r.Checks, Item{
			Name:   "jarvis.yaml",
			OK:     false,
			Detail: "not found in this directory or parents",
		})
		r.OK = allOK(r.Checks)
		return r
	}
	r.Checks = append(r.Checks, Item{Name: "jarvis.yaml", OK: true, Detail: path})

	if loadErr != nil {
		r.Checks = append(r.Checks, Item{Name: "skills", OK: false, Detail: loadErr.Error()})
		r.OK = allOK(r.Checks)
		return r
	}
	root := filepath.Dir(path)
	home, err := config.HomeDir(cfg)
	if err != nil {
		r.Checks = append(r.Checks, Item{Name: "JARVIS_HOME", OK: false, Detail: err.Error()})
	} else {
		r.Checks = append(r.Checks, Item{Name: "JARVIS_HOME", OK: true, Detail: home})
	}
	problems := install.CheckLinks(cfg, root)
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
	r.Checks = append(r.Checks, roleSkillsCheck(cfg, res, product))
	r.Checks = append(r.Checks, historyChecks(cfg, root, product)...)
	if product {
		r.Checks = append(r.Checks, trackerCheck(cfg, res))
	}
	r.Checks = append(r.Checks, localDirCheck(cfg, root))
	r.Checks = append(r.Checks, projectCheck(cfg, product))
	r.OK = allOK(r.Checks)
	return r
}

func roleCheck(res role.Resolved) Item {
	switch {
	case res.Project == "":
		return Item{Name: "role", OK: true, Detail: res.Role + ", default (no active project)"}
	case res.Err != nil:
		return Item{Name: "role", OK: true, Detail: res.Role + ", default (project " + res.Project + " did not load)"}
	case res.Missing:
		return Item{
			Name:     "role",
			OK:       true,
			Warn:     true,
			Detail:   res.Role + ", default; project " + res.Project + " has no role (warn)",
			Problems: []string{"run: " + role.FixCommand(res.Project)},
		}
	default:
		return Item{Name: "role", OK: true, Detail: res.Role + ", from project " + res.Project}
	}
}

// roleSkillsCheck fails for product and warns for engineer when a skill of
// the active role is not installed.
func roleSkillsCheck(cfg *config.Config, res role.Resolved, product bool) Item {
	missing := role.MissingSkills(res.Role, cfg.Skills.Installed)
	if len(missing) == 0 {
		return Item{Name: "role skills", OK: true, Detail: fmt.Sprintf("%d of %d installed for role %s", len(res.Skills), len(res.Skills), res.Role)}
	}
	item := Item{
		Name:     "role skills",
		OK:       !product,
		Warn:     !product,
		Detail:   fmt.Sprintf("role %s needs: %s", res.Role, strings.Join(missing, ", ")),
		Problems: []string{"run: jarvis skills install --role " + res.Role},
	}
	if !product {
		item.Detail += " (warn)"
	}
	return item
}

func historyChecks(cfg *config.Config, root string, product bool) []Item {
	fail := func(detail string) []Item {
		it := Item{Name: "history.path", OK: !product, Warn: !product, Detail: detail}
		if !product {
			it.Detail += " (warn)"
		}
		return []Item{it}
	}
	if strings.TrimSpace(cfg.History.Path) == "" {
		return fail("empty in jarvis.yaml")
	}
	hist, err := config.AbsPath(root, cfg.History.Path)
	if err != nil {
		return fail(err.Error())
	}
	if st, err := os.Stat(hist); err != nil || !st.IsDir() {
		return fail(hist + " does not exist")
	}
	if _, err := runGit(hist, "rev-parse", "--is-inside-work-tree"); err != nil {
		return fail(hist + " is not a git checkout")
	}
	out := []Item{{Name: "history.path", OK: true, Detail: hist}}
	if product {
		remotes, err := runGit(hist, "remote")
		if err != nil || strings.TrimSpace(remotes) == "" {
			out = append(out, Item{Name: "history remote", OK: true, Warn: true, Detail: "no git remote; the PRD cannot be pushed (warn)"})
		} else {
			out = append(out, Item{Name: "history remote", OK: true, Detail: strings.Join(strings.Fields(remotes), ", ")})
		}
	}
	return out
}

func trackerCheck(cfg *config.Config, res role.Resolved) Item {
	if res.Project != "" {
		if p, err := project.Load(res.Project); err == nil && config.Configured(p.Tracker, project.TokenSet(p.Name)) {
			return Item{Name: "tracker", OK: true, Detail: "from project " + p.Name}
		}
	}
	if cfg.Tracker.Set() {
		return Item{Name: "tracker", OK: true, Detail: "from jarvis.yaml"}
	}
	return Item{Name: "tracker", OK: true, Warn: true, Detail: "not set; plan source will be user (warn)"}
}

func localDirCheck(cfg *config.Config, root string) Item {
	if !cfg.Skills.UsesLocal() {
		return Item{Name: "local skills dir", OK: true, Detail: "not used (scope " + cfg.Skills.Scope + ")"}
	}
	if install.LocalInsideHistory(cfg, root) {
		return Item{Name: "local skills dir", OK: false, Detail: install.ErrLocalInsideHistory.Error()}
	}
	return Item{Name: "local skills dir", OK: true, Detail: config.LocalSkillsDir(cfg, root)}
}

func projectCheck(cfg *config.Config, product bool) Item {
	name := strings.TrimSpace(cfg.Project)
	if name == "" {
		return Item{Name: "project", OK: true, Detail: "none set"}
	}
	p, err := project.Load(name)
	if err != nil {
		return Item{Name: "project", OK: false, Detail: err.Error()}
	}
	if product {
		return Item{Name: "project", OK: true, Detail: name + " (product role: repo paths and graph not checked)"}
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
	gjson, err := p.GraphJSON()
	if err == nil {
		if st, e := os.Stat(gjson); e == nil && !st.IsDir() {
			sp, _ := p.StampsPath()
			if _, e := os.Stat(sp); e != nil {
				item.Warn = true
				item.Problems = append(item.Problems, "graph.json exists but stamps.json is missing; run jarvis project graph "+name)
				if item.Detail == name {
					item.Detail = name + "; graph stamps missing (warn)"
				}
			}
		}
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

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	b, err := cmd.Output()
	return string(b), err
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
