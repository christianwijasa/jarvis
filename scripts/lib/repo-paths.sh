#!/usr/bin/env bash
# Resolve personal project repo paths from workstation/repos.manifest.
# Source this file; do not execute directly.

_PROJECTS_ROOT="${PERSONAL_PROJECTS_ROOT:-$HOME/projects}"
_MANIFEST="${REPOS_MANIFEST:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/workstation/repos.manifest}"

repo_local_path() {
  local name="$1"
  printf '%s/%s\n' "$_PROJECTS_ROOT" "$name"
}

# Optional third column in manifest is full remote URL.
_repo_remote_from_manifest="${REPO_REMOTE_URL:-}"

repo_remote_url() {
  local name="$1"
  if [[ -n "${_repo_remote_from_manifest}" ]]; then
    printf '%s\n' "$_repo_remote_from_manifest"
    return 0
  fi
  if [[ -n "${GIT_REMOTE_PREFIX:-}" ]]; then
    printf '%s%s.git\n' "$GIT_REMOTE_PREFIX" "$name"
    return 0
  fi
  local user="${GITHUB_USER:-YOU}"
  printf 'git@%s:%s/%s.git\n' "${GITHUB_HOST:-github.com}" "$user" "$name"
}

# Usage: while read -r name branch remote; do ...; done < <(list_repos)
list_repos() {
  if [[ ! -f "$_MANIFEST" ]]; then
    echo "repos manifest not found: $_MANIFEST" >&2
    return 1
  fi
  grep -Ev '^\s*(#|$)' "$_MANIFEST" | awk '{print $1, $2, $3}'
}

list_repo_paths() {
  local name branch remote path
  while read -r name branch remote; do
    path="$(repo_local_path "$name")" || return 1
    printf '%s\n' "$path"
  done < <(list_repos)
}

repo_default_branch() {
  local repo_path="$1"
  local base
  base="$(git -C "$repo_path" remote show origin 2>/dev/null | sed -n '/HEAD branch/s/.*: //p' || true)"
  printf '%s\n' "${base:-main}"
}

repo_origin_base_rev() {
  local repo_path="$1"
  local base rev
  base="$(repo_default_branch "$repo_path")"
  git -C "$repo_path" fetch origin --quiet
  rev="$(git -C "$repo_path" rev-parse "origin/$base")"
  printf '%s\n' "$rev"
}
