package install

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/christianwijasa/jarvis/internal/config"
)

// Renamed maps an old skill name to its current name. It is a migration
// note: installs made before a rename may still hold the old folder.
var Renamed = map[string]string{
	"trd-pr-planning": "trd-planning",
}

// CurrentName returns the current name of a skill, following Renamed.
func CurrentName(name string) (string, bool) {
	if n, ok := Renamed[name]; ok {
		return n, true
	}
	return name, false
}

// Migration is what MigrateRenamed did.
type Migration struct {
	// Removed lists folders and links Jarvis deleted.
	Removed []string
	// Kept lists old-name folders left alone because Jarvis did not
	// install them.
	Kept []string
	// Install lists current names that replace old names found in
	// skills.installed. The caller installs them.
	Install []string
}

// MigrateRenamed removes skill folders left under an old name. It deletes
// only what Jarvis itself installed: a folder whose .jarvis-pin names the
// old skill, or a local symlink that points into $JARVIS_HOME. Anything
// else is kept and reported. Old names in skills.installed are replaced by
// the current names, which are returned in Install.
func MigrateRenamed(cfg *config.Config, projectRoot string) (Migration, error) {
	var m Migration
	globalRoot, err := config.GlobalSkillsDir(cfg)
	if err != nil {
		return m, err
	}
	olds := make([]string, 0, len(Renamed))
	for old := range Renamed {
		olds = append(olds, old)
	}
	sort.Strings(olds)
	for _, old := range olds {
		if cfg != nil && cfg.Skills.UsesLocal() && !LocalInsideHistory(cfg, projectRoot) {
			local := filepath.Join(config.LocalSkillsDir(cfg, projectRoot), old)
			if err := removeOwnedLocal(local, old, globalRoot, &m); err != nil {
				return m, err
			}
		}
		if err := removeOwnedDir(filepath.Join(globalRoot, old), old, &m); err != nil {
			return m, err
		}
		if cfg != nil && cfg.Skills.Has(old) {
			cfg.Skills.Remove(old)
			m.Install = append(m.Install, Renamed[old])
		}
	}
	return m, nil
}

func removeOwnedLocal(path, name, globalRoot string, m *Migration) error {
	st, err := os.Lstat(path)
	if err != nil {
		return nil
	}
	if st.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			m.Kept = append(m.Kept, path)
			return nil
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		home := filepath.Dir(globalRoot)
		if within(resolve(home), resolve(target)) || within(filepath.Clean(home), filepath.Clean(target)) {
			if err := os.Remove(path); err != nil {
				return err
			}
			m.Removed = append(m.Removed, path)
			return nil
		}
		m.Kept = append(m.Kept, path)
		return nil
	}
	return removeOwnedDir(path, name, m)
}

func removeOwnedDir(dir, name string, m *Migration) error {
	st, err := os.Lstat(dir)
	if err != nil {
		return nil
	}
	if !st.IsDir() || !pinNames(filepath.Join(dir, pinFile), name) {
		m.Kept = append(m.Kept, dir)
		return nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	m.Removed = append(m.Removed, dir)
	return nil
}

// pinNames reports whether a .jarvis-pin file says name=<name>.
func pinNames(path, name string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) == "name="+name {
			return true
		}
	}
	return false
}
