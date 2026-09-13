#!/usr/bin/env bash
set -euo pipefail

HUB="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib/repo-paths.sh
source "$HUB/scripts/lib/repo-paths.sh"

MODE="${1:-sync}"  # sync | clone | pull | status

usage() {
  cat <<EOF
Usage: $0 [sync|clone|pull|status]

  sync   clone missing repos, pull existing (default)
  clone  clone missing repos only
  pull   git pull existing repos only
  status show clone/pull state without changing anything

Manifest: workstation/repos.manifest
Projects root: \$PERSONAL_PROJECTS_ROOT (default: ~/projects)
Override clone URL: GIT_REMOTE_PREFIX=git@github.com:you/
Per-repo URL: third column in manifest
EOF
}

clone_repo() {
  local name="$1" branch="${2:-}" remote="${3:-}"
  local path url
  path="$(repo_local_path "$name")"
  if [[ -n "$remote" && "$remote" != "-" ]]; then
    url="$remote"
  else
    url="$(repo_remote_url "$name")"
  fi
  mkdir -p "$(dirname "$path")"
  if [[ -n "$branch" && "$branch" != "-" ]]; then
    echo "CLONE  $name -> $path (branch $branch)"
    git clone --branch "$branch" "$url" "$path"
  else
    echo "CLONE  $name -> $path (default branch)"
    git clone "$url" "$path"
  fi
}

pull_repo() {
  local path="$1" name="$2"
  echo "PULL   $name ($path)"
  git -C "$path" pull --ff-only
}

process_repo() {
  local name="$1" branch="$2" remote="$3"
  local path
  path="$(repo_local_path "$name")"

  if [[ ! -d "$path/.git" ]]; then
    case "$MODE" in
      sync|clone)
        clone_repo "$name" "$branch" "$remote"
        ;;
      status)
        echo "MISSING $name -> $path"
        ;;
      pull)
        echo "SKIP   $name (not cloned)"
        ;;
    esac
    return 0
  fi

  case "$MODE" in
    sync|pull)
      pull_repo "$path" "$name"
      ;;
    status)
      local ref
      ref="$(git -C "$path" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "?")"
      echo "OK     $name @ $path (branch $ref)"
      ;;
    clone)
      echo "SKIP   $name (already cloned)"
      ;;
  esac
}

case "$MODE" in
  sync|clone|pull|status) ;;
  -h|--help|help) usage; exit 0 ;;
  *) echo "Unknown mode: $MODE" >&2; usage >&2; exit 1 ;;
esac

echo "=== Personal workstation ($MODE) ==="
echo "Manifest: $_MANIFEST"
echo "Projects root: $_PROJECTS_ROOT"
echo

failed=0
while read -r name branch remote; do
  if ! process_repo "$name" "$branch" "$remote"; then
    failed=1
  fi
done < <(list_repos)

echo
if [[ "$failed" -ne 0 ]]; then
  echo "Completed with errors." >&2
  exit 1
fi
echo "Done."
