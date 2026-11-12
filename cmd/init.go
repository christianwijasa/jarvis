package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/christianwijasabfi/atlas/internal/config"
	"github.com/christianwijasabfi/atlas/internal/install"
	"github.com/christianwijasabfi/atlas/internal/project"
	"github.com/christianwijasabfi/atlas/internal/upstream"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newInitCmd() *cobra.Command {
	var (
		recommended bool
		skillsFlag  string
		yes         bool
		scopeFlag   string
		copyLocal   bool
		localDir    string
		agentsMD    bool
		projectName string
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write atlas.yaml and install chosen skills",
		Long: `Create atlas.yaml in the current directory and install the chosen
skills into $ATLAS_HOME/skills (default ~/.atlas/skills).

On a TTY, Atlas asks: Recommended vs Choose skills, then global / local / both.
If stdin is not a TTY, Atlas never prompts. Pass --yes, --recommended, or
--skills (and optionally --scope). Missing flags exit 1 with the flag names.

--agents-md appends a marked block to AGENTS.md telling other agents to run
atlas help --agent first. Existing AGENTS.md is not overwritten; the marked
block is inserted or refreshed idempotently.

--project <name> links or creates a named repo group after skills are
installed. Non-TTY skips this unless the flag is given.`,
		Example: `  atlas init --yes
  atlas init --recommended --scope both --agents-md
  atlas init --skills grilling,graphify --scope local --copy
  atlas init --yes --project shop`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := config.FileName
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%s already exists", path)
			} else if !os.IsNotExist(err) {
				return err
			}

			reg, err := upstream.Load()
			if err != nil {
				return err
			}

			in := bufio.NewReader(cmd.InOrStdin())
			interactive := stdinIsTerminal(cmd.InOrStdin()) && !yes
			if !interactive && !recommended && !cmd.Flags().Changed("skills") && !yes {
				return fmt.Errorf("stdin is not a TTY; pass --yes (recommended skills, scope both), or --recommended / --skills and optionally --scope, so Atlas does not prompt")
			}

			names, err := chooseSkillNames(cmd, in, reg, recommended, skillsFlag, cmd.Flags().Changed("skills"), interactive)
			if err != nil {
				return err
			}
			scope, err := chooseScope(cmd, in, scopeFlag, interactive)
			if err != nil {
				return err
			}

			var cfg config.Config
			if err := yaml.Unmarshal(config.ExampleYAML(), &cfg); err != nil {
				return err
			}
			cfg.Skills.Scope = scope
			if localDir != "" {
				cfg.Skills.LocalDir = localDir
			}
			cfg.Skills.Copy = copyLocal
			cfg.Skills.Installed = nil
			cfg.Skills.Normalize()

			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, name := range names {
				entry, ok := reg.Skill(name)
				if !ok {
					return fmt.Errorf("unknown skill %q", name)
				}
				res, err := install.Install(&cfg, wd, entry, skillFetcher)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "Installed %s (%s) at %s\n", res.Name, res.Source, res.Global)
				if res.Local != "" {
					how := "linked"
					if res.Copied {
						how = "copied"
					}
					fmt.Fprintf(out, "  %s to %s\n", how, res.Local)
				}
				cfg.Skills.Add(name)
			}
			if err := applyInitProject(cmd, in, interactive, projectName, &cfg); err != nil {
				return err
			}
			if err := config.Write(path, &cfg); err != nil {
				return err
			}
			fmt.Fprintf(out, "Wrote %s (scope %s, %d skill(s))\n", path, cfg.Skills.Scope, len(cfg.Skills.Installed))
			if cfg.Project != "" {
				fmt.Fprintf(out, "Active project: %s\n", cfg.Project)
			}
			fmt.Fprintln(out, "Edit tracker, history.path, and repos for this team.")
			if agentsMD {
				action, err := upsertAgentsMD(agentsMDFile)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "AGENTS.md: %s marked Atlas block\n", action)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&recommended, "recommended", false, "install all prepared skills (10 bundled plus archify)")
	cmd.Flags().StringVar(&skillsFlag, "skills", "", "comma-separated skill names to install")
	cmd.Flags().BoolVar(&yes, "yes", false, "non-interactive; recommended skills and scope both unless flags say otherwise")
	cmd.Flags().StringVar(&scopeFlag, "scope", "", "where agents load skills: global, local, or both")
	cmd.Flags().BoolVar(&copyLocal, "copy", false, "copy into the project instead of symlink from $ATLAS_HOME")
	cmd.Flags().StringVar(&localDir, "local-dir", "", "project-local skills directory (default .agents/skills)")
	cmd.Flags().BoolVar(&agentsMD, "agents-md", false, "append or refresh a marked Atlas block in AGENTS.md (never overwrite the whole file)")
	cmd.Flags().StringVar(&projectName, "project", "", "link or create a named project (non-TTY: skipped unless this flag is set)")
	return cmd
}

func applyInitProject(cmd *cobra.Command, in *bufio.Reader, interactive bool, flagName string, cfg *config.Config) error {
	name := strings.TrimSpace(flagName)
	if name == "" && interactive {
		fmt.Fprintln(cmd.OutOrStdout(), "Named project (kebab-case) to create or link, empty to skip:")
		fmt.Fprint(cmd.OutOrStdout(), "Project: ")
		line, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		name = strings.TrimSpace(line)
	}
	if name == "" {
		return nil
	}
	if err := project.ValidateName(name); err != nil {
		return err
	}
	if _, err := project.Load(name); err != nil {
		p := &project.Project{SchemaVersion: project.SchemaVersion, Name: name}
		if err := project.Save(p); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created project %s\n", name)
	}
	cfg.Project = name
	return nil
}

func chooseSkillNames(cmd *cobra.Command, in *bufio.Reader, reg *upstream.Registry, recommended bool, skillsFlag string, skillsSet, interactive bool) ([]string, error) {
	if recommended && skillsSet {
		return nil, fmt.Errorf("use either --recommended or --skills, not both")
	}
	if recommended {
		return reg.RecommendedSkillNames(), nil
	}
	if skillsSet {
		return parseSkillList(reg, skillsFlag)
	}
	if !interactive {
		return reg.RecommendedSkillNames(), nil
	}
	return promptSkills(cmd, in, reg)
}

func chooseScope(cmd *cobra.Command, in *bufio.Reader, scopeFlag string, interactive bool) (string, error) {
	if strings.TrimSpace(scopeFlag) != "" {
		return config.ParseScope(scopeFlag)
	}
	if !interactive {
		return config.ScopeBoth, nil
	}
	return promptScope(cmd, in)
}

func parseSkillList(reg *upstream.Registry, raw string) ([]string, error) {
	var names []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := reg.Skill(p); !ok {
			return nil, fmt.Errorf("unknown skill %q", p)
		}
		names = append(names, p)
	}
	return names, nil
}

func promptSkills(cmd *cobra.Command, in *bufio.Reader, reg *upstream.Registry) ([]string, error) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Skills to install:")
	fmt.Fprintln(out, "  [1] Recommended — all prepared skills (10 bundled + archify)  [default]")
	fmt.Fprintln(out, "  [2] Choose — pick individually")
	fmt.Fprint(out, "Choice [1]: ")
	line, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" || line == "1" {
		return reg.RecommendedSkillNames(), nil
	}
	if line != "2" {
		return nil, fmt.Errorf("expected 1 or 2")
	}
	catalog := reg.Skills()
	fmt.Fprintln(out, "Available skills (comma-separated numbers or names):")
	for i, e := range catalog {
		mark := " "
		if e.Recommended {
			mark = "*"
		}
		fmt.Fprintf(out, "  %2d%s %s — %s\n", i+1, mark, e.Name, e.Summary)
	}
	fmt.Fprint(out, "Skills: ")
	sel, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	return parseChoiceList(reg, catalog, sel)
}

func parseChoiceList(reg *upstream.Registry, catalog []upstream.Entry, sel string) ([]string, error) {
	var names []string
	for _, p := range strings.Split(sel, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			if n < 1 || n > len(catalog) {
				return nil, fmt.Errorf("skill number %d out of range", n)
			}
			names = append(names, catalog[n-1].Name)
			continue
		}
		if _, ok := reg.Skill(p); !ok {
			return nil, fmt.Errorf("unknown skill %q", p)
		}
		names = append(names, p)
	}
	return names, nil
}

func promptScope(cmd *cobra.Command, in *bufio.Reader) (string, error) {
	out := cmd.OutOrStdout()
	home, err := config.HomeDir(nil)
	if err != nil {
		home = "~/.atlas"
	}
	fmt.Fprintln(out, "Where should agents load skills from?")
	fmt.Fprintf(out, "  [1] Global only (%s/skills)\n", home)
	fmt.Fprintf(out, "  [2] Project-local (%s, linked from global)\n", config.DefaultLocalDir)
	fmt.Fprintln(out, "  [3] Both  [default]")
	fmt.Fprint(out, "Choice [3]: ")
	line, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	switch strings.TrimSpace(line) {
	case "", "3":
		return config.ScopeBoth, nil
	case "1":
		return config.ScopeGlobal, nil
	case "2":
		return config.ScopeLocal, nil
	default:
		return "", fmt.Errorf("expected 1, 2, or 3")
	}
}

func stdinIsTerminal(r io.Reader) bool {
	if terminalCheck != nil {
		return terminalCheck(r)
	}
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

var terminalCheck func(io.Reader) bool
