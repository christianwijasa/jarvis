package cmd

import (
	"os"
	"strings"
)

const (
	agentsMDFile  = "AGENTS.md"
	agentsMarkBeg = "<!-- atlas-agents-md:start -->"
	agentsMarkEnd = "<!-- atlas-agents-md:end -->"
)

const agentsMDSnippet = `## Atlas

Before planning or editing in this repository, run:

    atlas help --agent

Use ` + "`atlas help --agent --json`" + ` for machine-readable help (` + "`schema_version`" + ` ` + SchemaVersion + `).
Prefer non-interactive flags (` + "`--yes`" + `, ` + "`--skills`" + `, ` + "`--scope`" + `). Do not wait on Atlas prompts.
Load a skill with ` + "`atlas skills show <name>`" + `.
`

func upsertAgentsMD(path string) (string, error) {
	block := agentsMarkBeg + "\n" + strings.TrimSpace(agentsMDSnippet) + "\n" + agentsMarkEnd + "\n"
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(block), 0o644); err != nil {
				return "", err
			}
			return "wrote", nil
		}
		return "", err
	}
	s := string(existing)
	i := strings.Index(s, agentsMarkBeg)
	j := strings.Index(s, agentsMarkEnd)
	if i >= 0 && j > i {
		rest := ""
		after := j + len(agentsMarkEnd)
		if after < len(s) {
			rest = strings.TrimLeft(s[after:], "\n")
		}
		prefix := s[:i]
		next := prefix + block
		if rest != "" {
			if !strings.HasSuffix(next, "\n") {
				next += "\n"
			}
			next += rest
			if !strings.HasSuffix(next, "\n") {
				next += "\n"
			}
		}
		if next == s {
			return "unchanged", nil
		}
		if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
			return "", err
		}
		return "updated", nil
	}
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	if err := os.WriteFile(path, []byte(s+"\n"+block), 0o644); err != nil {
		return "", err
	}
	return "appended", nil
}
