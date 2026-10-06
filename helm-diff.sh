#!/usr/bin/env bash
# Diff a chart's rendered manifests between BASE_REF and the current checkout.
set -euo pipefail
: "${CHART:?}" "${BASE_REF:?}"

cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d)
git fetch --quiet --no-tags --depth=1 origin "$BASE_REF"
git worktree add --quiet --detach "$tmp/base" FETCH_HEAD
trap 'git worktree remove --force "$tmp/base"' EXIT

# Render the chart in the given checkout; prints nothing if the chart doesn't exist there
render() {
  [ -f "$1/$CHART/Chart.yaml" ] || return 0
  local args=(--no-hooks)
  for f in ${VALUES:-}; do
    [ -f "$1/$f" ] && args+=(--values "$1/$f")
  done
  helm template "$1/$CHART" "${args[@]}" ${ARGS:-}
}

render "$tmp/base" > "$tmp/base.yaml"
render . > "$tmp/head.yaml"

changed=true
diff -u --label "$BASE_REF" --label HEAD "$tmp/base.yaml" "$tmp/head.yaml" && { echo "No changes."; changed=false; }
echo "changed=$changed" >> "${GITHUB_OUTPUT:-/dev/null}"
