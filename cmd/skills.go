package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/install"
	"github.com/christianwijasa/jarvis/internal/role"
	"github.com/christianwijasa/jarvis/internal/upstream"
	"github.com/christianwijasa/jarvis/skills"
	"github.com/spf13/cobra"
)

var skillFetcher install.Fetcher = install.DefaultFetcher{}

func newSkillsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List installed vs available skills",
		Long: `List catalog skills that are installed in this project versus still
available to install. Shows global and local paths.

Each skill shows whether it applies to the role of the active project
(engineer or product). Does not prompt. Pass --json for a schema_version
object with role, role_source, and per-skill roles and applies.
Load one skill body with: jarvis skills show <name>`,
		Example: `  jarvis skills
  jarvis skills --json
  jarvis skills show grilling
  jarvis skills install archify
  jarvis skills remove archify`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rep, err := buildSkillsReport()
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), rep)
			}
			printSkillsText(cmd, rep)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	cmd.AddCommand(newSkillsShowCmd(), newSkillsInstallCmd(), newSkillsRemoveCmd(), newSkillsUpstreamCmd())
	return cmd
}

type skillsReport struct {
	SchemaVersion string      `json:"schema_version"`
	Kind          string      `json:"kind"`
	SkillPack     string      `json:"skill_pack"`
	Role          string      `json:"role"`
	RoleSource    string      `json:"role_source"`
	JarvisHome    string      `json:"jarvis_home,omitempty"`
	Scope         string      `json:"scope,omitempty"`
	LocalDir      string      `json:"local_dir,omitempty"`
	Installed     []skillItem `json:"installed"`
	Available     []skillItem `json:"available"`
}

type skillItem struct {
	Name        string         `json:"name"`
	Mode        string         `json:"mode"`
	Recommended bool           `json:"recommended,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	Roles       []string       `json:"roles"`
	Applies     bool           `json:"applies"`
	Place       *install.Place `json:"place,omitempty"`
}

func buildSkillsReport() (*skillsReport, error) {
	reg, err := upstream.Load()
	if err != nil {
		return nil, err
	}
	cfg, projectRoot, _ := loadProjectConfig()
	rr := role.Resolve(cfg)
	installed := map[string]bool{}
	if cfg != nil {
		for _, n := range cfg.Skills.Installed {
			installed[n] = true
		}
	}
	rep := &skillsReport{
		SchemaVersion: SchemaVersion,
		Kind:          "skills",
		SkillPack:     skills.Version(),
		Role:          rr.Role,
		RoleSource:    rr.Source,
		Installed:     []skillItem{},
		Available:     []skillItem{},
	}
	if home, err := config.HomeDir(cfg); err == nil {
		rep.JarvisHome = home
	}
	if cfg != nil {
		rep.Scope = cfg.Skills.Scope
		rep.LocalDir = cfg.Skills.LocalDir
	}
	for _, e := range reg.Skills() {
		item := skillItem{
			Name:        e.Name,
			Mode:        e.Mode,
			Recommended: e.Recommended,
			Summary:     e.Summary,
			Roles:       e.EffectiveRoles(),
			Applies:     e.HasRole(rr.Role),
		}
		if installed[e.Name] {
			p := install.PlaceOf(cfg, projectRoot, e.Name)
			item.Place = &p
			rep.Installed = append(rep.Installed, item)
			continue
		}
		rep.Available = append(rep.Available, item)
	}
	return rep, nil
}

func printSkillsText(cmd *cobra.Command, rep *skillsReport) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Skill pack %s (catalog in upstream.yaml)\n", rep.SkillPack)
	if rep.JarvisHome != "" {
		fmt.Fprintf(out, "JARVIS_HOME %s\n", rep.JarvisHome)
	}
	if rep.Scope != "" {
		fmt.Fprintf(out, "scope %s  local_dir %s\n", rep.Scope, rep.LocalDir)
	}
	fmt.Fprintf(out, "role %s (%s)\n", rep.Role, rep.RoleSource)
	fmt.Fprintln(out, "Installed:")
	if len(rep.Installed) == 0 {
		fmt.Fprintln(out, "  (none)")
	}
	for _, item := range rep.Installed {
		loc := "not installed"
		if item.Place != nil {
			loc = formatPlace(*item.Place)
		}
		fmt.Fprintf(out, "  %-24s mode:%s  applies:%s  %s\n", item.Name, item.Mode, yesNo(item.Applies), loc)
	}
	fmt.Fprintln(out, "Available:")
	for _, item := range rep.Available {
		rec := ""
		if item.Recommended {
			rec = "recommended, "
		}
		fmt.Fprintf(out, "  %-24s %smode:%s  applies:%s  %s\n", item.Name, rec, item.Mode, yesNo(item.Applies), item.Summary)
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func formatPlace(p install.Place) string {
	var parts []string
	if p.Global != "" {
		parts = append(parts, "global "+p.Global)
	}
	if p.Local != "" {
		if p.LocalKind == "copy" {
			parts = append(parts, "local "+p.Local+" (copy)")
		} else {
			parts = append(parts, "local "+p.Local)
		}
	}
	if len(parts) == 0 {
		return "not installed"
	}
	out := parts[0]
	if len(parts) > 1 {
		out += "; " + parts[1]
	}
	return out
}

func modeOf(reg *upstream.Registry, name string) string {
	m := reg.ModeFor(name)
	if m == "" {
		return "mode:unknown"
	}
	return "mode:" + m
}

func newSkillsShowCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Print an installed or bundled skill",
		Long: `Print SKILL.md for a named skill. Prefer the installed copy under
$JARVIS_HOME, else bundled text. Vendored skills must be installed first.

This is how an agent should load a skill. Does not prompt. --json wraps
the body with schema_version.`,
		Example: `  jarvis skills show grilling
  jarvis skills show grilling --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, _, _ := loadProjectConfig()
			source := ""
			body := ""
			if global, err := config.GlobalSkillsDir(cfg); err == nil {
				if b, err := install.ReadInstalledSkill(filepath.Join(global, name)); err == nil {
					source = "installed"
					body = b
				}
			}
			if body == "" {
				if b, pack, err := skills.Read(name); err == nil {
					source = "bundled:" + pack
					body = b
				}
			}
			if body == "" {
				reg, _ := upstream.Load()
				if e, ok := reg.Skill(name); ok && e.Mode == upstream.ModeVendored {
					return fmt.Errorf("skill %q is not installed; run: jarvis skills install %s", name, name)
				}
				return fmt.Errorf("unknown skill %q", name)
			}
			rr := role.Resolve(cfg)
			applies := true
			if reg, err := upstream.Load(); err == nil {
				if e, ok := reg.Skill(name); ok {
					applies = e.HasRole(rr.Role)
				}
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), skillShowJSON{
					SchemaVersion: SchemaVersion,
					Kind:          "skill",
					Name:          name,
					Source:        source,
					Mode:          modeOfMust(name),
					Applies:       applies,
					Body:          body,
				})
			}
			if !applies {
				fmt.Fprintf(cmd.ErrOrStderr(), "note: %s does not apply to role %s (active project).\n", name, rr.Role)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "# %s  %s\n\n", source, modeOfMust(name))
			fmt.Fprint(cmd.OutOrStdout(), body)
			ensureNL(cmd, body)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

type skillShowJSON struct {
	SchemaVersion string `json:"schema_version"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	Mode          string `json:"mode"`
	Applies       bool   `json:"applies"`
	Body          string `json:"body"`
}

func newSkillsInstallCmd() *cobra.Command {
	var copyLocal, force bool
	var roleFlag string
	cmd := &cobra.Command{
		Use:   "install [name...]",
		Short: "Install catalog skills at the pinned upstream commit (or bundled text)",
		Long: `Install one or more catalog skills. Bundled skills are copied from this
binary. Vendored skills are fetched at the pinned commit in upstream.yaml
(git or gh tarball). Never follows HEAD or main.

--role engineer|product installs that role's recommended set (no names
needed). Run it once per role you use; installs add up.

A local install never replaces a folder that Jarvis did not create, unless
--force is passed. A local skills dir inside history.path is refused.

Requires jarvis.yaml. Does not prompt. Use --copy to copy into the project
instead of symlink.`,
		Example: `  jarvis skills install grilling
  jarvis skills install archify --copy
  jarvis skills install --role product`,
		Args: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("role") {
				return nil
			}
			return cobra.MinimumNArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, projectRoot, cfgPath, err := requireProjectConfig()
			if err != nil {
				return err
			}
			if copyLocal {
				cfg.Skills.Copy = true
			}
			reg, err := upstream.Load()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			mig, err := install.MigrateRenamed(cfg, projectRoot)
			if err != nil {
				return err
			}
			reportMigration(out, mig)
			names := currentNames(out, args)
			names = append(names, mig.Install...)
			if cmd.Flags().Changed("role") {
				r := strings.ToLower(strings.TrimSpace(roleFlag))
				if !role.Valid(r) {
					return fmt.Errorf("--role must be engineer or product (got %q)", roleFlag)
				}
				names = append(names, reg.SkillsForRole(r)...)
			}
			for _, name := range uniqueNames(names) {
				entry, ok := reg.Skill(name)
				if !ok {
					return fmt.Errorf("unknown skill %q", name)
				}
				res, err := install.InstallWith(cfg, projectRoot, entry, skillFetcher, install.Options{Force: force})
				if err != nil {
					return err
				}
				cfg.Skills.Add(name)
				fmt.Fprintf(out, "Installed %s (%s) at %s\n", res.Name, res.Source, res.Global)
				if res.Local != "" {
					how := "linked"
					if res.Copied {
						how = "copied"
					}
					fmt.Fprintf(out, "  %s to %s\n", how, res.Local)
				}
			}
			return config.Write(cfgPath, cfg)
		},
	}
	cmd.Flags().BoolVar(&copyLocal, "copy", false, "copy into the project instead of symlink")
	cmd.Flags().StringVar(&roleFlag, "role", "", "install the recommended set for a role: engineer or product")
	cmd.Flags().BoolVar(&force, "force", false, "replace existing local skill folders that Jarvis did not create")
	return cmd
}

// currentNames maps renamed skills to their current names and says so.
func currentNames(out io.Writer, in []string) []string {
	res := make([]string, 0, len(in))
	for _, n := range in {
		cur, renamed := install.CurrentName(n)
		if renamed {
			fmt.Fprintf(out, "Skill %s was renamed to %s; installing %s\n", n, cur, cur)
		}
		res = append(res, cur)
	}
	return res
}

// reportMigration prints what MigrateRenamed removed or kept.
func reportMigration(out io.Writer, m install.Migration) {
	for _, p := range m.Removed {
		fmt.Fprintf(out, "Removed old skill folder %s (renamed skill)\n", p)
	}
	for _, p := range m.Kept {
		fmt.Fprintf(out, "Kept %s: not installed by Jarvis; remove it by hand if unused\n", p)
	}
}

func uniqueNames(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range in {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

func newSkillsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed skill from JARVIS_HOME and the project local dir",
		Long: `Delete $JARVIS_HOME/skills/<name> and the project-local link or copy.
Updates skills.installed in jarvis.yaml. Does not prompt.`,
		Example: `  jarvis skills remove archify`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, projectRoot, cfgPath, err := requireProjectConfig()
			if err != nil {
				return err
			}
			name := args[0]
			if err := install.Remove(cfg, projectRoot, name); err != nil {
				return err
			}
			cfg.Skills.Remove(name)
			if err := config.Write(cfgPath, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", name)
			return nil
		},
	}
}

func newSkillsUpstreamCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "upstream",
		Short: "Report whether pinned upstream paths have moved (read-only)",
		Long: `Compare each upstream.yaml row to the latest commit on that
upstream path (gh api, or git ls-remote if gh fails).

This command never copies, merges, or rewrites skills. Adopting an
update is a manual, reviewed pull request. Install always uses the
pin, never latest. Does not prompt.`,
		Example: `  jarvis skills upstream
  jarvis skills upstream --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := upstream.Load()
			if err != nil {
				return err
			}
			remote := upstreamRemote()
			results := upstream.Check(reg, remote)
			if asJSON {
				rows := make([]upstreamRow, 0, len(results))
				failed := 0
				for _, r := range results {
					rows = append(rows, upstreamRow{
						Name:    r.Entry.Name,
						Kind:    r.Entry.Kind,
						Pack:    r.Entry.Pack,
						Mode:    r.Entry.Mode,
						Status:  r.Status,
						Pinned:  r.Pinned,
						Latest:  r.Latest,
						Compare: r.Compare,
						Detail:  r.Detail,
					})
					if r.Status == "error" {
						failed++
					}
				}
				if err := writeJSON(cmd.OutOrStdout(), upstreamJSON{
					SchemaVersion: SchemaVersion,
					Kind:          "upstream",
					ReadOnly:      true,
					Results:       rows,
				}); err != nil {
					return err
				}
				if failed > 0 {
					return fmt.Errorf("%d upstream lookups failed", failed)
				}
				return nil
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Read-only check. Nothing was applied.")
			failed := 0
			for _, r := range results {
				e := r.Entry
				fmt.Fprintf(out, "%s (%s, %s, mode:%s)\n", e.Name, e.Kind, e.Pack, e.Mode)
				fmt.Fprintf(out, "  status:  %s\n", r.Status)
				fmt.Fprintf(out, "  pinned:  %s\n", r.Pinned)
				if r.Latest != "" {
					fmt.Fprintf(out, "  latest:  %s\n", r.Latest)
				}
				if r.Compare != "" {
					fmt.Fprintf(out, "  compare: %s\n", r.Compare)
				}
				if r.Detail != "" {
					fmt.Fprintf(out, "  note:    %s\n", r.Detail)
				}
				if r.Status == "error" {
					failed++
				}
			}
			if failed > 0 {
				return fmt.Errorf("%d upstream lookups failed", failed)
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

type upstreamJSON struct {
	SchemaVersion string        `json:"schema_version"`
	Kind          string        `json:"kind"`
	ReadOnly      bool          `json:"read_only"`
	Results       []upstreamRow `json:"results"`
}

type upstreamRow struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Pack    string `json:"pack"`
	Mode    string `json:"mode"`
	Status  string `json:"status"`
	Pinned  string `json:"pinned"`
	Latest  string `json:"latest,omitempty"`
	Compare string `json:"compare,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

var upstreamRemote = func() upstream.Remote {
	return upstream.GHRemote{}
}

func loadProjectConfig() (*config.Config, string, string) {
	path, err := config.Find("")
	if err != nil {
		return nil, "", ""
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", path
	}
	return cfg, filepath.Dir(path), path
}

func requireProjectConfig() (*config.Config, string, string, error) {
	path, err := config.Find("")
	if err != nil {
		return nil, "", "", fmt.Errorf("jarvis.yaml not found; run jarvis init")
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", "", err
	}
	return cfg, filepath.Dir(path), path, nil
}

func modeOfMust(name string) string {
	reg, err := upstream.Load()
	if err != nil {
		return "mode:unknown"
	}
	return modeOf(reg, name)
}

func ensureNL(cmd *cobra.Command, body string) {
	if len(body) > 0 && body[len(body)-1] != '\n' {
		fmt.Fprintln(cmd.OutOrStdout())
	}
}
