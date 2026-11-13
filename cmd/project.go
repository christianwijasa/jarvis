package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/project"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Named groups of repositories and their combined graph",
		Long: `A project is a named group of git repositories. Jarvis stores each
one under $JARVIS_HOME/projects/<name>/project.yaml (default ~/.jarvis).

The point of a project is one knowledge graph spanning every member repo
so agents can trace cross-repo calls. Build it with jarvis project graph.
Never hardcodes product names; examples in help use shop / cart-service.

Read commands accept --json (schema_version jarvis.cli/v1).`,
		Example: `  jarvis project create shop --description "Example shop" --repo cart-service=https://github.com/example/cart-service.git
  jarvis project use shop
  jarvis project graph shop
  jarvis project show shop --json`,
	}
	cmd.AddCommand(
		newProjectCreateCmd(),
		newProjectUpdateCmd(),
		newProjectAddRepoCmd(),
		newProjectRemoveRepoCmd(),
		newProjectListCmd(),
		newProjectShowCmd(),
		newProjectDeleteCmd(),
		newProjectUseCmd(),
		newProjectCurrentCmd(),
		newProjectCloneCmd(),
		newProjectStatusCmd(),
		newProjectGraphCmd(),
		newProjectGraphPathCmd(),
	)
	return cmd
}

func newProjectCreateCmd() *cobra.Command {
	var desc string
	var repos []string
	var tf trackerFlags
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a project yaml under $JARVIS_HOME/projects/<name>/",
		Long: `Create a named project. Names are lowercase kebab-case.

On a TTY, Jarvis may ask for a description, repos, and optional tracker
(URL, token, MCP id). Enter skips tracker fields. If stdin is not a TTY,
pass flags; the name argument is enough. --repo may be repeated:
  name=url
  name=url@/abs/or/~/path

Tracker token is stored in $JARVIS_HOME/projects/<name>/secrets (mode 0600),
never in project.yaml.`,
		Example: `  jarvis project create shop --description "Checkout and payments" \
    --repo cart-service=https://github.com/example/cart-service.git \
    --tracker-kind jira --tracker-site https://example.atlassian.net`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := project.ValidateName(name); err != nil {
				return err
			}
			if _, err := project.Load(name); err == nil {
				return fmt.Errorf("project %q already exists", name)
			}
			p := &project.Project{SchemaVersion: project.SchemaVersion, Name: name, Description: desc}
			applyTrackerFlags(cmd, p, tf, true)
			token := strings.TrimSpace(tf.token)
			for _, spec := range repos {
				r, err := parseRepoSpec(spec)
				if err != nil {
					return err
				}
				if err := p.SetRepo(r); err != nil {
					return err
				}
			}
			interactive := stdinIsTerminal(cmd.InOrStdin())
			if interactive && desc == "" && len(repos) == 0 {
				in := bufio.NewReader(cmd.InOrStdin())
				if err := promptProject(cmd, in, p); err != nil {
					return err
				}
			}
			if interactive && !anyTrackerFlag(cmd) {
				in := bufio.NewReader(cmd.InOrStdin())
				tok, err := promptTracker(cmd, in, p, false)
				if err != nil {
					return err
				}
				if tok != "" {
					token = tok
				}
			}
			if p.Tracker.Kind == "" && p.Tracker.Set() {
				p.Tracker.Kind = "jira"
			}
			if err := project.Save(p); err != nil {
				return err
			}
			if token != "" {
				if err := project.WriteToken(p.Name, token); err != nil {
					return err
				}
			}
			dir, _ := project.Dir(p.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Created project %s at %s/%s\n", p.Name, dir, project.FileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&desc, "description", "", "optional description")
	cmd.Flags().StringArrayVar(&repos, "repo", nil, "repo spec name=url[@path] (repeatable)")
	addTrackerFlags(cmd, &tf)
	return cmd
}

func newProjectUpdateCmd() *cobra.Command {
	var desc string
	var unsetTracker, unsetToken bool
	var tf trackerFlags
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change description or tracker settings on an existing project",
		Long: `Update fields that you pass. Omitted flags keep their current values.

--unset-tracker clears kind/site/mcp/project in project.yaml.
--unset-tracker-token deletes the secrets file.
On a TTY with no flags, Jarvis asks optional tracker questions (enter keeps current).`,
		Example: `  jarvis project update shop --tracker-site https://example.atlassian.net --tracker-mcp plugin-atlassian-atlassian
  jarvis project update shop --unset-tracker-token`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			if unsetTracker && anyTrackerFlag(cmd) {
				return fmt.Errorf("do not combine --unset-tracker with --tracker-* flags")
			}
			if unsetToken && cmd.Flags().Changed("tracker-token") {
				return fmt.Errorf("do not combine --unset-tracker-token with --tracker-token")
			}
			flagsUsed := cmd.Flags().Changed("description") || anyTrackerFlag(cmd) || unsetTracker || unsetToken
			if cmd.Flags().Changed("description") {
				p.Description = desc
			}
			if unsetTracker {
				p.Tracker = config.Tracker{}
			} else {
				applyTrackerFlags(cmd, p, tf, false)
			}
			token := ""
			writeTok := false
			if cmd.Flags().Changed("tracker-token") {
				token = strings.TrimSpace(tf.token)
				writeTok = true
			}
			interactive := stdinIsTerminal(cmd.InOrStdin())
			if interactive && !flagsUsed {
				in := bufio.NewReader(cmd.InOrStdin())
				tok, err := promptTracker(cmd, in, p, true)
				if err != nil {
					return err
				}
				if tok != "" {
					token = tok
					writeTok = true
				}
			}
			if p.Tracker.Kind == "" && p.Tracker.Set() {
				p.Tracker.Kind = "jira"
			}
			if err := project.Save(p); err != nil {
				return err
			}
			if unsetToken {
				if err := project.DeleteToken(p.Name); err != nil {
					return err
				}
			} else if writeTok {
				if err := project.WriteToken(p.Name, token); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated project %s\n", p.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&desc, "description", "", "replace description")
	cmd.Flags().BoolVar(&unsetTracker, "unset-tracker", false, "clear tracker kind/site/mcp/project")
	cmd.Flags().BoolVar(&unsetToken, "unset-tracker-token", false, "delete stored tracker token")
	addTrackerFlags(cmd, &tf)
	return cmd
}

type trackerFlags struct {
	kind, site, project, mcp, token string
}

func addTrackerFlags(cmd *cobra.Command, f *trackerFlags) {
	cmd.Flags().StringVar(&f.kind, "tracker-kind", "", "tracker kind (jira, trello, github, linear, other)")
	cmd.Flags().StringVar(&f.site, "tracker-site", "", "tracker site URL")
	cmd.Flags().StringVar(&f.project, "tracker-project", "", "tracker project or board id")
	cmd.Flags().StringVar(&f.mcp, "tracker-mcp", "", "host MCP id")
	cmd.Flags().StringVar(&f.token, "tracker-token", "", "tracker API token (stored in secrets, never printed)")
}

func anyTrackerFlag(cmd *cobra.Command) bool {
	for _, n := range []string{"tracker-kind", "tracker-site", "tracker-project", "tracker-mcp", "tracker-token"} {
		if cmd.Flags().Changed(n) {
			return true
		}
	}
	return false
}

func applyTrackerFlags(cmd *cobra.Command, p *project.Project, f trackerFlags, create bool) {
	set := func(name, val string, dst *string) {
		if create && !cmd.Flags().Changed(name) {
			return
		}
		if create || cmd.Flags().Changed(name) {
			*dst = strings.TrimSpace(val)
		}
	}
	set("tracker-kind", f.kind, &p.Tracker.Kind)
	set("tracker-site", f.site, &p.Tracker.Site)
	set("tracker-project", f.project, &p.Tracker.Project)
	set("tracker-mcp", f.mcp, &p.Tracker.MCP)
}

func promptProject(cmd *cobra.Command, in *bufio.Reader, p *project.Project) error {
	out := cmd.OutOrStdout()
	fmt.Fprint(out, "Description (optional): ")
	line, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	p.Description = strings.TrimSpace(line)
	fmt.Fprintln(out, "Repos as name=url[@path], empty line to finish:")
	for {
		fmt.Fprint(out, "  repo: ")
		line, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return nil
		}
		r, err := parseRepoSpec(line)
		if err != nil {
			return err
		}
		if err := p.SetRepo(r); err != nil {
			return err
		}
	}
}

func promptTracker(cmd *cobra.Command, in *bufio.Reader, p *project.Project, keepCurrent bool) (string, error) {
	out := cmd.OutOrStdout()
	read := func(label, current string) (string, error) {
		if keepCurrent && current != "" {
			fmt.Fprintf(out, "%s [%s]: ", label, current)
		} else {
			fmt.Fprintf(out, "%s: ", label)
		}
		line, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line == "" && keepCurrent {
			return current, nil
		}
		return line, nil
	}
	site, err := read("Tracker URL (optional)", p.Tracker.Site)
	if err != nil {
		return "", err
	}
	p.Tracker.Site = site
	kind, err := read("Tracker kind (optional)", p.Tracker.Kind)
	if err != nil {
		return "", err
	}
	p.Tracker.Kind = kind
	mcp, err := read("MCP id (optional)", p.Tracker.MCP)
	if err != nil {
		return "", err
	}
	p.Tracker.MCP = mcp
	fmt.Fprint(out, "Tracker token (optional, not echoed in later commands): ")
	line, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func parseRepoSpec(spec string) (project.Repo, error) {
	spec = strings.TrimSpace(spec)
	name, rest, ok := strings.Cut(spec, "=")
	if !ok || strings.TrimSpace(name) == "" {
		return project.Repo{}, fmt.Errorf("repo spec %q must be name=url[@path]", spec)
	}
	r := project.Repo{Name: strings.TrimSpace(name)}
	url, path := splitURLPath(strings.TrimSpace(rest))
	r.URL = url
	r.Path = path
	return r, nil
}

func splitURLPath(rest string) (url, path string) {
	i := strings.LastIndex(rest, "@")
	if i <= 0 {
		return rest, ""
	}
	suffix := rest[i+1:]
	if suffix == "" {
		return rest, ""
	}
	if strings.HasPrefix(suffix, "/") || strings.HasPrefix(suffix, "~") || strings.HasPrefix(suffix, ".") {
		return rest[:i], suffix
	}
	return rest, ""
}

func newProjectAddRepoCmd() *cobra.Command {
	var url, path, branch string
	cmd := &cobra.Command{
		Use:     "add-repo <project> <name>",
		Short:   "Add or replace a repo in a project",
		Example: `  jarvis project add-repo shop cart-service --url https://github.com/example/cart-service.git --path ~/src/cart-service --branch main`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			r := project.Repo{Name: args[1], URL: url, Path: path, DefaultBranch: branch}
			if err := p.SetRepo(r); err != nil {
				return err
			}
			if err := project.Save(p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added repo %s to %s\n", r.Name, p.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "git URL")
	cmd.Flags().StringVar(&path, "path", "", "local checkout path")
	cmd.Flags().StringVar(&branch, "branch", "", "default branch")
	return cmd
}

func newProjectRemoveRepoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove-repo <project> <name>",
		Short:   "Remove a repo from a project yaml (does not delete the checkout)",
		Example: `  jarvis project remove-repo shop cart-service`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			if err := p.RemoveRepo(args[1]); err != nil {
				return err
			}
			if err := project.Save(p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed repo %s from %s\n", args[1], p.Name)
			return nil
		},
	}
}

func newProjectListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List projects in $JARVIS_HOME/projects",
		Example: `  jarvis project list --json`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			list, err := project.List()
			if err != nil {
				return err
			}
			if asJSON {
				summaries := make([]map[string]any, 0, len(list))
				for _, p := range list {
					summaries = append(summaries, map[string]any{
						"name":        p.Name,
						"description": p.Description,
						"repos":       len(p.Repos),
					})
				}
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version": SchemaVersion,
					"kind":           "project_list",
					"projects":       summaries,
				})
			}
			if len(list) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no projects)")
				return nil
			}
			for _, p := range list {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %d repo(s)  %s\n", p.Name, len(p.Repos), p.Description)
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

func newProjectShowCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "show <name>",
		Short:   "Show a project, its repos, and graph path",
		Example: `  jarvis project show shop --json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			gdir, _ := p.GraphDir()
			gjson, _ := p.GraphJSON()
			tokenSet := project.TokenSet(p.Name)
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version":    SchemaVersion,
					"kind":              "project",
					"project":           p,
					"graph_dir":         gdir,
					"graph_json":        gjson,
					"tracker_token_set": tokenSet,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "project %s\n", p.Name)
			if p.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p.Description)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "tracker kind=%s site=%s mcp=%s token_set=%v\n", p.Tracker.Kind, p.Tracker.Site, p.Tracker.MCP, tokenSet)
			fmt.Fprintf(cmd.OutOrStdout(), "graph: %s\n", gjson)
			for _, r := range p.Repos {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s url=%s path=%s branch=%s\n", r.Name, r.URL, r.Path, r.DefaultBranch)
			}
			if len(p.Repos) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "  (no repos)")
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

func newProjectDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete the project yaml only (never checkouts)",
		Long: `Remove $JARVIS_HOME/projects/<name>/project.yaml. Does not delete
cloned repositories or graphify-out. Requires --yes when stdin is not a TTY.`,
		Example: `  jarvis project delete shop --yes`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if _, err := project.Load(name); err != nil {
				return err
			}
			if !yes {
				if !stdinIsTerminal(cmd.InOrStdin()) {
					return fmt.Errorf("stdin is not a TTY; pass --yes to delete project %q yaml (checkouts are never removed)", name)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Delete project %s yaml only (not repo checkouts)? [y/N]: ", name)
				in := bufio.NewReader(cmd.InOrStdin())
				line, _ := in.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(line)) != "y" && strings.ToLower(strings.TrimSpace(line)) != "yes" {
					return fmt.Errorf("aborted")
				}
			}
			if err := project.Delete(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted project %s yaml\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "do not prompt")
	return cmd
}

func newProjectUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "use <name>",
		Short:   "Set project: <name> in jarvis.yaml so agents know the active project",
		Example: `  jarvis project use shop`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if _, err := project.Load(name); err != nil {
				return err
			}
			cfg, _, cfgPath, err := requireProjectConfig()
			if err != nil {
				return err
			}
			cfg.Project = name
			if err := config.Write(cfgPath, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Active project is %s (%s)\n", name, cfgPath)
			return nil
		},
	}
}

func newProjectCurrentCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "current",
		Short:   "Print the active project from jarvis.yaml",
		Example: `  jarvis project current --json`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, _, cfgPath := loadProjectConfig()
			name := ""
			if cfg != nil {
				name = cfg.Project
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version": SchemaVersion,
					"kind":           "project_current",
					"project":        name,
					"config_file":    cfgPath,
				})
			}
			if name == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "(no active project)")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), name)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}

func newProjectCloneCmd() *cobra.Command {
	var dir string
	var depth int
	cmd := &cobra.Command{
		Use:   "clone <name>",
		Short: "Clone missing repos that have a url; skip existing paths",
		Long: `Clone each repo that has a git URL and no existing checkout.
Never pulls, resets, or deletes. Updates path in the project yaml for
newly cloned repos. Existing paths are skipped.`,
		Example: `  jarvis project clone shop --dir ~/src --depth 1`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			results, err := project.Clone(p, dir, depth)
			out := cmd.OutOrStdout()
			for _, r := range results {
				fmt.Fprintf(out, "%s: %s %s %s\n", r.Name, r.Status, r.Path, r.Detail)
			}
			return err
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "parent directory for repos that have no path")
	cmd.Flags().IntVar(&depth, "depth", 0, "shallow clone depth (optional)")
	return cmd
}

func newProjectStatusCmd() *cobra.Command {
	var asJSON, fetch bool
	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Read-only git and graph-freshness status for each repo",
		Long: `For each repo: path exists, is a git repo, current branch, dirty,
ahead/behind vs upstream (no fetch unless --fetch). Graph freshness
compares stamps.json SHAs to origin/main or origin/master (stale if they
differ or stamps are missing). --fetch updates origin refs first.`,
		Example: `  jarvis project status shop --json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			rows, err := project.Status(p, fetch)
			if err != nil {
				return err
			}
			gjson, _ := p.GraphJSON()
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version": SchemaVersion,
					"kind":           "project_status",
					"project":        p.Name,
					"graph_json":     gjson,
					"repos":          rows,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "project %s  graph %s\n", p.Name, gjson)
			for _, r := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s exists=%v git=%v branch=%s dirty=%v ahead=%d behind=%d graph=%s %s\n",
					r.Name, r.Exists, r.GitRepo, r.Branch, r.Dirty, r.Ahead, r.Behind, r.Graph, r.Detail)
			}
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	cmd.Flags().BoolVar(&fetch, "fetch", false, "git fetch before ahead/behind (off by default)")
	return cmd
}

func newProjectGraphCmd() *cobra.Command {
	var update, ifStale bool
	var repos string
	cmd := &cobra.Command{
		Use:   "graph <name>",
		Short: "Build or refresh the combined graphify graph for all repos",
		Long: `Shells out to graphify v0.9.15:

  graphify extract <repo-path> --out $JARVIS_HOME/projects/<name>/extract/<repo> --code-only
  graphify merge-graphs <each extract/.../graphify-out/graph.json> --out <graph.json>

extract --out keeps graphify-out out of member checkouts. merge-graphs
prefixes node ids with the repo name (parent.parent of each graph.json).
Writes stamps.json with each repo HEAD SHA after a successful build.
--if-stale fetches origin, compares stamps to origin/main or origin/master,
and extracts only stale or missing members. Agents run this before query.
--update re-runs extract (incremental if a previous extract exists).
--repos a,b limits which members are extracted and merged.

If extract --out is empty or fails, Jarvis copies that repo's in-repo
graphify-out/graph.json when present, otherwise skips the repo and
continues. Combined output stays under $JARVIS_HOME/projects/<name>/.

Progress (extract i/n <repo>, merge n graphs) and live graphify logs go
to stderr. Success is one stdout line with the combined graph path.`,
		Example: `  jarvis project graph shop
  jarvis project graph shop --if-stale
  jarvis project graph shop --update --repos cart-service,payment-service`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			var filter []string
			if strings.TrimSpace(repos) != "" {
				for _, n := range strings.Split(repos, ",") {
					n = strings.TrimSpace(n)
					if n != "" {
						filter = append(filter, n)
					}
				}
			}
			_ = update
			old := project.Progress
			project.Progress = cmd.ErrOrStderr()
			defer func() { project.Progress = old }()
			res, err := project.BuildGraph(p, filter, ifStale)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project graph written to %s (%d repo(s))\n", res.GraphJSON, len(res.Repos))
			return nil
		},
	}
	cmd.Flags().BoolVar(&update, "update", false, "re-extract even if an extract already exists")
	cmd.Flags().BoolVar(&ifStale, "if-stale", false, "fetch origin and extract only repos whose stamp SHA differs from origin/main or origin/master")
	cmd.Flags().StringVar(&repos, "repos", "", "comma-separated repo names to include")
	return cmd
}

func newProjectGraphPathCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "graph-path <name>",
		Short:   "Print the project graph.json path for graphify query",
		Example: `  jarvis project graph-path shop --json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			gjson, err := p.GraphJSON()
			if err != nil {
				return err
			}
			gdir, _ := p.GraphDir()
			rows, _ := project.CheckStale(p, false)
			fresh := project.AllFresh(rows)
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version": SchemaVersion,
					"kind":           "project_graph_path",
					"project":        p.Name,
					"graph_dir":      gdir,
					"graph_json":     gjson,
					"fresh":          fresh,
					"repos":          rows,
				})
			}
			fmt.Fprintln(cmd.OutOrStdout(), gjson)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}
