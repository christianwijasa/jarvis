package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
	"github.com/christianwijasa/jarvis/internal/upstream"
	"github.com/christianwijasa/jarvis/skills"
)

const pinFile = ".jarvis-pin"

type Result struct {
	Name   string
	Global string
	Local  string
	Linked bool
	Copied bool
	Source string // bundled or vendored
}

// Options change how Install places a skill.
type Options struct {
	// Force replaces an existing local folder that Jarvis did not create.
	Force bool
}

// Install writes one skill with default options (no --force).
func Install(cfg *config.Config, projectRoot string, entry upstream.Entry, fetcher Fetcher) (Result, error) {
	return InstallWith(cfg, projectRoot, entry, fetcher, Options{})
}

// InstallWith writes one skill into $JARVIS_HOME/skills and, for scope local
// or both, links or copies it into the project-local skills dir.
func InstallWith(cfg *config.Config, projectRoot string, entry upstream.Entry, fetcher Fetcher, opts Options) (Result, error) {
	var r Result
	r.Name = entry.Name
	if cfg != nil && cfg.Skills.UsesLocal() {
		if err := CheckLocalTarget(cfg, projectRoot, entry.Name, opts.Force); err != nil {
			return r, err
		}
	}
	globalRoot, err := config.GlobalSkillsDir(cfg)
	if err != nil {
		return r, err
	}
	dest := filepath.Join(globalRoot, entry.Name)
	if err := os.MkdirAll(globalRoot, 0o755); err != nil {
		return r, err
	}
	if err := os.RemoveAll(dest); err != nil {
		return r, err
	}

	var files map[string][]byte
	switch {
	case entry.Mode == upstream.ModeVendored:
		if fetcher == nil {
			return r, fmt.Errorf("no fetcher for vendored skill %s", entry.Name)
		}
		if err := AssertPinned(entry.PinnedCommit); err != nil {
			return r, fmt.Errorf("%s: %w", entry.Name, err)
		}
		files, err = fetcher.FetchTree(entry.UpstreamRepo, entry.PinnedCommit, entry.UpstreamPath)
		if err != nil {
			return r, fmt.Errorf("fetch %s at %s: %w", entry.Name, entry.PinnedCommit, err)
		}
		r.Source = upstream.ModeVendored
	case entry.Bundled():
		files, err = skills.Tree(entry.Name)
		if err != nil {
			return r, err
		}
		r.Source = "bundled"
	default:
		return r, fmt.Errorf("skill %s has no bundled text and is not vendored", entry.Name)
	}
	if err := writeTree(dest, files); err != nil {
		return r, err
	}
	if err := writePin(dest, entry); err != nil {
		return r, err
	}
	r.Global = dest

	if cfg != nil && cfg.Skills.UsesLocal() {
		localRoot := config.LocalSkillsDir(cfg, projectRoot)
		localDest := filepath.Join(localRoot, entry.Name)
		linked, copied, err := placeLocal(dest, localDest, cfg.Skills.Copy)
		if err != nil {
			return r, err
		}
		r.Local = localDest
		r.Linked = linked
		r.Copied = copied
	}
	return r, nil
}

// ErrLocalInsideHistory means skills.local_dir resolves inside history.path.
var ErrLocalInsideHistory = errors.New("skills.local_dir is inside history.path; use --scope global for this setup")

// LocalInsideHistory reports whether the project-local skills dir resolves to
// a path inside history.path. A history repo can keep its own skills, and a
// local install there would replace them.
func LocalInsideHistory(cfg *config.Config, projectRoot string) bool {
	if cfg == nil || strings.TrimSpace(cfg.History.Path) == "" {
		return false
	}
	hist, err := config.AbsPath(projectRoot, cfg.History.Path)
	if err != nil {
		return false
	}
	local, err := filepath.Abs(config.LocalSkillsDir(cfg, projectRoot))
	if err != nil {
		return false
	}
	return within(resolve(hist), resolve(local))
}

// CheckLocalTarget refuses a local placement that would delete files Jarvis
// does not own: a local dir inside history.path, or an existing folder that
// is not a link into $JARVIS_HOME and has no Jarvis pin file. force skips the
// second check only.
func CheckLocalTarget(cfg *config.Config, projectRoot, name string, force bool) error {
	if LocalInsideHistory(cfg, projectRoot) {
		return ErrLocalInsideHistory
	}
	if force {
		return nil
	}
	localDest := filepath.Join(config.LocalSkillsDir(cfg, projectRoot), name)
	st, err := os.Lstat(localDest)
	if err != nil {
		return nil
	}
	if st.Mode()&os.ModeSymlink != 0 {
		target, err := filepath.EvalSymlinks(localDest)
		globalRoot, gerr := config.GlobalSkillsDir(cfg)
		if err != nil || gerr != nil {
			// A broken link holds no files; replacing it loses nothing.
			return nil
		}
		home := filepath.Dir(globalRoot)
		if within(resolve(home), target) {
			return nil
		}
		return fmt.Errorf("%s is a link to %s, not into %s; pass --force to replace it", localDest, target, home)
	}
	if st.IsDir() {
		if _, err := os.Stat(filepath.Join(localDest, pinFile)); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%s exists and was not installed by Jarvis; pass --force to replace it", localDest)
}

// resolve follows symlinks in the longest existing prefix of p, so a path
// that does not exist yet still compares correctly.
func resolve(p string) string {
	p = filepath.Clean(p)
	rest := ""
	for {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			return filepath.Join(r, rest)
		}
		parent := filepath.Dir(p)
		if parent == p {
			return filepath.Join(p, rest)
		}
		rest = filepath.Join(filepath.Base(p), rest)
		p = parent
	}
}

func within(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func Remove(cfg *config.Config, projectRoot, name string) error {
	globalRoot, err := config.GlobalSkillsDir(cfg)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(globalRoot, name)); err != nil {
		return err
	}
	if cfg != nil && cfg.Skills.UsesLocal() {
		local := filepath.Join(config.LocalSkillsDir(cfg, projectRoot), name)
		if err := os.RemoveAll(local); err != nil {
			return err
		}
	}
	return nil
}

func placeLocal(globalDest, localDest string, forceCopy bool) (linked, copied bool, err error) {
	if err := os.MkdirAll(filepath.Dir(localDest), 0o755); err != nil {
		return false, false, err
	}
	if err := os.RemoveAll(localDest); err != nil {
		return false, false, err
	}
	if !forceCopy {
		if err := os.Symlink(globalDest, localDest); err == nil {
			return true, false, nil
		}
	}
	files, err := readDirTree(globalDest)
	if err != nil {
		return false, false, err
	}
	if err := writeTree(localDest, files); err != nil {
		return false, false, err
	}
	return false, true, nil
}

func writeTree(dest string, files map[string][]byte) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	for rel, body := range files {
		if strings.Contains(rel, "..") {
			return fmt.Errorf("refusing path %q", rel)
		}
		p := filepath.Join(dest, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writePin(dest string, entry upstream.Entry) error {
	var b strings.Builder
	fmt.Fprintf(&b, "name=%s\n", entry.Name)
	fmt.Fprintf(&b, "mode=%s\n", entry.Mode)
	if entry.PinnedCommit != "" {
		fmt.Fprintf(&b, "commit=%s\n", entry.PinnedCommit)
	}
	if entry.PinnedTag != "" {
		fmt.Fprintf(&b, "tag=%s\n", entry.PinnedTag)
	}
	if entry.UpstreamRepo != "" {
		fmt.Fprintf(&b, "repo=%s\n", entry.UpstreamRepo)
	}
	if entry.UpstreamPath != "" {
		fmt.Fprintf(&b, "path=%s\n", entry.UpstreamPath)
	}
	if entry.Mode != upstream.ModeVendored {
		fmt.Fprintf(&b, "source=bundled\n")
		fmt.Fprintf(&b, "pack=%s\n", skills.Version())
	}
	return os.WriteFile(filepath.Join(dest, pinFile), []byte(b.String()), 0o644)
}

// Place is where one skill currently exists on disk.
type Place struct {
	Global    string `json:"global,omitempty"`
	Local     string `json:"local,omitempty"`
	LocalKind string `json:"local_kind,omitempty"`
}

func PlaceOf(cfg *config.Config, projectRoot, name string) Place {
	var p Place
	globalRoot, err := config.GlobalSkillsDir(cfg)
	if err == nil {
		g := filepath.Join(globalRoot, name)
		if skillPresent(g) {
			p.Global = g
		}
	}
	if cfg != nil && cfg.Skills.UsesLocal() {
		l := filepath.Join(config.LocalSkillsDir(cfg, projectRoot), name)
		if st, err := os.Lstat(l); err == nil {
			p.Local = l
			if st.Mode()&os.ModeSymlink != 0 {
				p.LocalKind = "symlink"
				if target, err := os.Readlink(l); err == nil {
					p.Local = l + " -> " + target
				}
			} else {
				p.LocalKind = "copy"
			}
		}
	}
	return p
}

func LocationLine(cfg *config.Config, projectRoot, name string) string {
	p := PlaceOf(cfg, projectRoot, name)
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
	return strings.Join(parts, "; ")
}

func skillPresent(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "SKILL.md"))
	return err == nil && !st.IsDir()
}

func ReadInstalledSkill(dir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckLinks(cfg *config.Config, projectRoot string) []string {
	var problems []string
	if cfg == nil {
		return problems
	}
	globalRoot, err := config.GlobalSkillsDir(cfg)
	if err != nil {
		return []string{err.Error()}
	}
	for _, name := range cfg.Skills.Installed {
		g := filepath.Join(globalRoot, name)
		if !skillPresent(g) {
			problems = append(problems, fmt.Sprintf("%s: missing global install at %s", name, g))
			continue
		}
		if !cfg.Skills.UsesLocal() {
			continue
		}
		l := filepath.Join(config.LocalSkillsDir(cfg, projectRoot), name)
		st, err := os.Lstat(l)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: missing local path %s", name, l))
			continue
		}
		if st.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(l)
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s: unreadable symlink %s", name, l))
				continue
			}
			resolved := target
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(filepath.Dir(l), target)
			}
			rg, _ := filepath.EvalSymlinks(g)
			rl, _ := filepath.EvalSymlinks(l)
			if rg != "" && rl != "" && rg != rl {
				problems = append(problems, fmt.Sprintf("%s: local link %s -> %s does not match %s", name, l, target, g))
			}
		} else if !skillPresent(l) {
			problems = append(problems, fmt.Sprintf("%s: local copy %s has no SKILL.md", name, l))
		}
	}
	return problems
}
