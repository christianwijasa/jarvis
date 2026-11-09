package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Fetcher downloads one tree from a git host at an exact commit.
// Implementations must not substitute HEAD, main, or any other floating ref.
type Fetcher interface {
	FetchTree(repo, commit, treePath string) (map[string][]byte, error)
}

// FuncFetcher adapts a function to Fetcher.
type FuncFetcher func(repo, commit, treePath string) (map[string][]byte, error)

func (f FuncFetcher) FetchTree(repo, commit, treePath string) (map[string][]byte, error) {
	return f(repo, commit, treePath)
}

// DefaultFetcher fetches a pinned tree with `gh api` tarball, then git.
type DefaultFetcher struct {
	Run      func(dir, name string, args ...string) ([]byte, error)
	LookPath func(string) (string, error)
}

func (f DefaultFetcher) run() func(dir, name string, args ...string) ([]byte, error) {
	if f.Run != nil {
		return f.Run
	}
	return func(dir, name string, args ...string) ([]byte, error) {
		cmd := exec.Command(name, args...)
		if dir != "" {
			cmd.Dir = dir
		}
		return cmd.CombinedOutput()
	}
}

func (f DefaultFetcher) lookPath() func(string) (string, error) {
	if f.LookPath != nil {
		return f.LookPath
	}
	return exec.LookPath
}

func (f DefaultFetcher) FetchTree(repo, commit, treePath string) (map[string][]byte, error) {
	if err := AssertPinned(commit); err != nil {
		return nil, err
	}
	run := f.run()
	if _, err := f.lookPath()("gh"); err == nil {
		files, err := fetchGHTarball(run, repo, commit, treePath)
		if err == nil {
			return files, nil
		}
	}
	return fetchGit(run, repo, commit, treePath)
}

func AssertPinned(commit string) error {
	c := strings.TrimSpace(commit)
	if c == "" || strings.EqualFold(c, "unknown") {
		return fmt.Errorf("refusing to fetch: no pinned commit")
	}
	switch strings.ToLower(c) {
	case "head", "main", "master", "origin/main", "origin/master", "origin/head":
		return fmt.Errorf("refusing to fetch floating ref %q; change the pin in upstream.yaml", c)
	}
	return nil
}

func fetchGHTarball(run func(dir, name string, args ...string) ([]byte, error), repo, commit, treePath string) (map[string][]byte, error) {
	owner, name, err := githubOwnerRepo(repo)
	if err != nil {
		return nil, err
	}
	raw, err := run("", "gh", "api", fmt.Sprintf("repos/%s/%s/tarball/%s", owner, name, commit))
	if err != nil {
		return nil, fmt.Errorf("gh tarball at %s: %w", commit, err)
	}
	files, err := extractTarballPrefix(raw, treePath)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("gh tarball at %s had no files under %s", commit, treePath)
	}
	return files, nil
}

func fetchGit(run func(dir, name string, args ...string) ([]byte, error), repo, commit, treePath string) (map[string][]byte, error) {
	tmp, err := os.MkdirTemp("", "atlas-skill-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	if out, err := run(tmp, "git", "init"); err != nil {
		return nil, fmt.Errorf("git init: %w\n%s", err, out)
	}
	if out, err := run(tmp, "git", "remote", "add", "origin", repo); err != nil {
		return nil, fmt.Errorf("git remote: %w\n%s", err, out)
	}
	if out, err := run(tmp, "git", "fetch", "--depth", "1", "origin", commit); err != nil {
		return nil, fmt.Errorf("git fetch pin %s: %w\n%s", commit, err, out)
	}
	if out, err := run(tmp, "git", "checkout", "FETCH_HEAD", "--", treePath); err != nil {
		return nil, fmt.Errorf("git checkout pin %s path %s: %w\n%s", commit, treePath, err, out)
	}
	root := filepath.Join(tmp, filepath.FromSlash(treePath))
	return readDirTree(root)
}

func githubOwnerRepo(repo string) (owner, name string, err error) {
	u := strings.TrimSuffix(strings.TrimSpace(repo), ".git")
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host == "" {
		return "", "", fmt.Errorf("not a github https url: %s", repo)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("not a github https url: %s", repo)
	}
	return parts[0], parts[1], nil
}

func extractTarballPrefix(raw []byte, treePath string) (map[string][]byte, error) {
	prefix := strings.Trim(treePath, "/")
	gr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("tarball gzip: %w", err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	files := map[string][]byte{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		name := hdr.Name
		if i := strings.Index(name, "/"); i >= 0 {
			name = name[i+1:]
		}
		name = path.Clean(name)
		if prefix != "" && prefix != "." {
			if name != prefix && !strings.HasPrefix(name, prefix+"/") {
				continue
			}
			name = strings.TrimPrefix(name, prefix+"/")
			if name == prefix {
				continue
			}
		}
		if name == "" || name == "." {
			continue
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[name] = b
	}
	return files, nil
}

func readDirTree(root string) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files under %s", root)
	}
	return files, nil
}
