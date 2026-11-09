package skills

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed VERSION
//go:embed all:trd-pr-planning all:grilling all:grill-with-docs all:domain-modeling
//go:embed all:test-case-design all:development all:debugging all:code-review
//go:embed all:generate-pr-description all:graphify
var bundledFS embed.FS

const PackBundled = "bundled"

// Info is one skill shipped in this binary.
type Info struct {
	Name string
	Pack string
}

// Version is the skill-pack version shipped in this binary.
func Version() string {
	b, err := bundledFS.ReadFile("VERSION")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(b))
}

// BundledFS is the embedded original/adapted skill trees.
func BundledFS() fs.FS {
	return bundledFS
}

// List returns skills whose text is embedded in the binary.
func List() ([]Info, error) {
	return namesIn(bundledFS, ".", PackBundled)
}

// Names is the sorted list of bundled skill names.
func Names() ([]string, error) {
	infos, err := List()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(infos))
	for i, info := range infos {
		out[i] = info.Name
	}
	return out, nil
}

// Read returns the SKILL.md body of a bundled skill.
func Read(name string) (body string, pack string, err error) {
	name = strings.TrimSuffix(name, ".md")
	name = strings.TrimSuffix(name, "/SKILL.md")
	b, e := fs.ReadFile(bundledFS, path.Join(name, "SKILL.md"))
	if e != nil {
		return "", "", fmt.Errorf("unknown bundled skill %q", name)
	}
	return string(b), PackBundled, nil
}

// Tree returns every file under a bundled skill directory.
func Tree(name string) (map[string][]byte, error) {
	name = strings.TrimSuffix(name, "/")
	files := map[string][]byte{}
	err := fs.WalkDir(bundledFS, name, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := fs.ReadFile(bundledFS, p)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, name+"/")
		if rel == p {
			return fmt.Errorf("skill %q is not bundled", name)
		}
		files[rel] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("unknown bundled skill %q", name)
	}
	return files, nil
}

func namesIn(fsys embed.FS, root, pack string) ([]Info, error) {
	var out []Info
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Base(p) != "SKILL.md" {
			return nil
		}
		name := path.Base(path.Dir(p))
		if name == "." {
			return nil
		}
		out = append(out, Info{Name: name, Pack: pack})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
