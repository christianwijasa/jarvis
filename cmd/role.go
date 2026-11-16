package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
	"github.com/christianwijasa/jarvis/internal/role"
	"github.com/spf13/cobra"
)

// errNoRoleAnswer means stdin ended before any answer (for example a closed
// or null stdin that still looks like a terminal). Callers then behave as if
// there were no TTY: warn and keep the default.
var errNoRoleAnswer = errors.New("no role answer")

// roleMaxTries is how many answers the role prompt accepts before it gives up.
const roleMaxTries = 3

// canPromptRole is true on a human TTY without --json. Agents and scripts
// never see the role prompt.
func canPromptRole(cmd *cobra.Command) bool {
	return stdinIsTerminal(cmd.InOrStdin()) && !wantsJSON(cmd)
}

// ensureProjectRole makes sure p has a role. On a TTY it asks once and saves
// the answer to project.yaml. Without a TTY, or with --json, it never asks:
// it prints one stderr warning and the project behaves as engineer.
func ensureProjectRole(cmd *cobra.Command, p *project.Project) (role.Resolved, error) {
	if p == nil {
		return role.Resolve(nil), nil
	}
	if p.HasRole() {
		return role.ForProject(p), nil
	}
	if !canPromptRole(cmd) {
		warnMissingRole(cmd, p.Name)
		return role.ForProject(p), nil
	}
	r, err := askRole(cmd.InOrStdin(), cmd.ErrOrStderr(), p.Name)
	if errors.Is(err, errNoRoleAnswer) {
		warnMissingRole(cmd, p.Name)
		return role.ForProject(p), nil
	}
	if err != nil {
		return role.Resolved{}, err
	}
	p.Role = r
	if err := project.Save(p); err != nil {
		return role.Resolved{}, err
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Saved role %s for project %s.\n", r, p.Name)
	return role.ForProject(p), nil
}

// ensureActiveRole resolves the role of the active project in cfg and asks
// for it when the project has none (see ensureProjectRole).
func ensureActiveRole(cmd *cobra.Command, cfg *config.Config) (role.Resolved, error) {
	res := role.Resolve(cfg)
	if !res.Missing {
		return res, nil
	}
	p, err := project.Load(res.Project)
	if err != nil {
		return res, nil
	}
	return ensureProjectRole(cmd, p)
}

func warnMissingRole(cmd *cobra.Command, name string) {
	fmt.Fprintln(cmd.ErrOrStderr(), "[jarvis] "+role.MissingWarning(name))
}

// askRole asks "What is your role in project <name>?" and returns engineer on
// an empty line. It returns errNoRoleAnswer when stdin ends first. It reads one byte at a time so it never consumes input
// meant for a later prompt.
func askRole(in io.Reader, out io.Writer, name string) (string, error) {
	for i := 0; i < roleMaxTries; i++ {
		fmt.Fprintf(out, "What is your role in project %s? [engineer/product] ", name)
		line, err := readLine(in)
		if err != nil && err != io.EOF {
			return "", err
		}
		if err == io.EOF && line == "" {
			fmt.Fprintln(out)
			return "", errNoRoleAnswer
		}
		ans := strings.ToLower(strings.TrimSpace(line))
		if ans == "" {
			return role.Engineer, nil
		}
		if role.Valid(ans) {
			return ans, nil
		}
		fmt.Fprintln(out, "Please answer engineer or product.")
		if err == io.EOF {
			break
		}
	}
	return "", fmt.Errorf("no role given for project %s; run: %s", name, role.FixCommand(name))
}

func readLine(r io.Reader) (string, error) {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return b.String(), nil
			}
			b.WriteByte(buf[0])
		}
		if err != nil {
			return b.String(), err
		}
	}
}

// askMissingRoles asks for a role for every project in $JARVIS_HOME/projects
// that has none. It runs after `jarvis update` on a TTY.
func askMissingRoles(cmd *cobra.Command) error {
	list, err := project.List()
	if err != nil {
		return err
	}
	for i := range list {
		p := &list[i]
		if p.HasRole() {
			continue
		}
		if _, err := ensureProjectRole(cmd, p); err != nil {
			return err
		}
	}
	return nil
}

// roleSkillsHint prints one line when the role's skills are not all installed.
func roleSkillsHint(cmd *cobra.Command, cfg *config.Config, r role.Resolved) {
	if cfg == nil {
		return
	}
	missing := role.MissingSkills(r.Role, cfg.Skills.Installed)
	if len(missing) == 0 {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Role %s needs: %s. Run: jarvis skills install --role %s\n", r.Role, strings.Join(missing, ", "), r.Role)
}

func roleFields(r role.Resolved) map[string]any {
	return map[string]any{
		"role":        r.Role,
		"role_source": r.Source,
		"role_skills": r.Skills,
		"plan_type":   r.PlanType,
	}
}

// productGuard refuses clone and graph on a product project.
func productGuard(p *project.Project) error {
	if p.EffectiveRole() == role.Product {
		return fmt.Errorf("project %s has role product; it does not clone or graph. Use an engineer project or change the role", p.Name)
	}
	return nil
}
