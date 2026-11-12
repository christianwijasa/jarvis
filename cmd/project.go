package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/christianwijasabfi/atlas/internal/config"
	"github.com/christianwijasabfi/atlas/internal/project"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Named groups of repositories and their combined graph",
		Long: `A project is a named group of git repositories. Atlas stores each
one under $ATLAS_HOME/projects/<name>/project.yaml (default ~/.atlas).

The point of a project is one knowledge graph spanning every member repo
so agents can trace cross-repo calls. Build it with atlas project graph.
Never hardcodes product names; examples in help use shop / cart-service.

Read commands accept --json (schema_version atlas.cli/v1).`,
		Example: `  atlas project create shop --description "Example shop" --repo cart-service=https://github.com/example/cart-service.git
  atlas project use shop
  atlas project graph shop
  atlas project show shop --json`,
	}
	cmd.AddCommand(
		newProjectCreateCmd(),
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
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a project yaml under $ATLAS_HOME/projects/<name>/",
		Long: `Create a named project. Names are lowercase kebab-case.

On a TTY, Atlas may ask for a description and repos. If stdin is not a TTY,
pass flags; the name argument is enough. --repo may be repeated:
  name=url
  name=url@/abs/or/~/path`,
		Example: `  atlas project create shop --description "Checkout and payments" \
    --repo cart-service=https://github.com/example/cart-service.git \
    --repo payment-service=https://github.com/example/payment-service.git@~/src/payment-service`,
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
			if err := project.Save(p); err != nil {
				return err
			}
			dir, _ := project.Dir(p.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Created project %s at %s/%s\n", p.Name, dir, project.FileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&desc, "description", "", "optional description")
	cmd.Flags().StringArrayVar(&repos, "repo", nil, "repo spec name=url[@path] (repeatable)")
	return cmd
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
		Example: `  atlas project add-repo shop cart-service --url https://github.com/example/cart-service.git --path ~/src/cart-service --branch main`,
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
		Example: `  atlas project remove-repo shop cart-service`,
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
		Short:   "List projects in $ATLAS_HOME/projects",
		Example: `  atlas project list --json`,
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
		Example: `  atlas project show shop --json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := project.Load(args[0])
			if err != nil {
				return err
			}
			gdir, _ := p.GraphDir()
			gjson, _ := p.GraphJSON()
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version": SchemaVersion,
					"kind":           "project",
					"project":        p,
					"graph_dir":      gdir,
					"graph_json":     gjson,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "project %s\n", p.Name)
			if p.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p.Description)
			}
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
		Long: `Remove $ATLAS_HOME/projects/<name>/project.yaml. Does not delete
cloned repositories or graphify-out. Requires --yes when stdin is not a TTY.`,
		Example: `  atlas project delete shop --yes`,
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
		Short:   "Set project: <name> in atlas.yaml so agents know the active project",
		Example: `  atlas project use shop`,
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
		Short:   "Print the active project from atlas.yaml",
		Example: `  atlas project current --json`,
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
		Example: `  atlas project clone shop --dir ~/src --depth 1`,
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
ahead/behind vs upstream (no fetch unless --fetch). Also compares
graphify-out/graph.json mtime to each repo HEAD commit (stale if HEAD is newer).`,
		Example: `  atlas project status shop --json`,
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
	var update bool
	var repos string
	cmd := &cobra.Command{
		Use:   "graph <name>",
		Short: "Build or refresh the combined graphify graph for all repos",
		Long: `Shells out to graphify v0.9.15:

  graphify extract <repo-path> --out $ATLAS_HOME/projects/<name>/extract/<repo> --code-only
  graphify merge-graphs <each extract/.../graphify-out/graph.json> --out <graph.json>

extract --out keeps graphify-out out of member checkouts. merge-graphs
prefixes node ids with the repo name (parent.parent of each graph.json).
--update re-runs extract (incremental if a previous extract exists).
--repos a,b limits which members are extracted and merged.

Progress (extract i/n <repo>, merge n graphs) and live graphify logs go
to stderr. Success is one stdout line with the combined graph path.`,
		Example: `  atlas project graph shop
  atlas project graph shop --update --repos cart-service,payment-service`,
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
			old := project.Progress
			project.Progress = cmd.ErrOrStderr()
			defer func() { project.Progress = old }()
			res, err := project.BuildGraph(p, filter, update)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project graph written to %s (%d repo(s))\n", res.GraphJSON, len(res.Repos))
			return nil
		},
	}
	cmd.Flags().BoolVar(&update, "update", false, "re-extract even if an extract already exists")
	cmd.Flags().StringVar(&repos, "repos", "", "comma-separated repo names to include")
	return cmd
}

func newProjectGraphPathCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "graph-path <name>",
		Short:   "Print the project graph.json path for graphify query",
		Example: `  atlas project graph-path shop --json`,
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
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"schema_version": SchemaVersion,
					"kind":           "project_graph_path",
					"project":        p.Name,
					"graph_dir":      gdir,
					"graph_json":     gjson,
				})
			}
			fmt.Fprintln(cmd.OutOrStdout(), gjson)
			return nil
		},
	}
	addJSONFlag(cmd, &asJSON)
	return cmd
}
