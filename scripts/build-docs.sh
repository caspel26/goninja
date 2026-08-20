#!/usr/bin/env bash
#
# Build the documentation site once per documented version into a single output
# tree, ready to publish as one artifact:
#
#   <out>/         the newest released version (what visitors land on)
#   <out>/vX.Y/    one directory per released minor series, newest patch wins
#   <out>/dev/     the current working tree, i.e. unreleased documentation
#
# Every version is built with the layouts, assets and configuration of the
# *current* tree — only `content/` comes from the tag. That keeps the version
# selector, theme and chrome identical across versions, and means a release
# published before the selector existed still gets one. The tradeoff is that a
# released version's pages are re-rendered by today's templates rather than
# frozen exactly as they shipped.
#
# Usage:
#   scripts/build-docs.sh [output-dir]
#
# Environment:
#   BASE_URL   site root without a trailing slash (default https://goninja.dev)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SITE_DIR="docs-site"
BASE_URL="${BASE_URL:-https://goninja.dev}"
BASE_URL="${BASE_URL%/}"
OUT="${1:-$REPO_ROOT/public}"

# Content paths, relative to the site's content/ directory, that are taken from
# the working tree for every version instead of from that version's tag.
UNVERSIONED=("docs/changelog")

cd "$REPO_ROOT"

command -v hugo >/dev/null || { echo "build-docs: hugo not found in PATH" >&2; exit 1; }

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

rm -rf "$OUT"
mkdir -p "$OUT"

# ---------------------------------------------------------------------------
# Work out which versions to publish.
#
# One directory per minor series (v0.1, v0.2, ...) built from the highest patch
# tag in that series, so v0.1.3 replaces v0.1.2 at /v0.1/ rather than piling up
# a directory per patch release.
# ---------------------------------------------------------------------------

# Parallel indexed arrays rather than an associative one: macOS still ships
# bash 3.2, which has no `declare -A`.
minors=()
tags=()

while IFS= read -r tag; do
  [[ -n "$tag" ]] || continue
  [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || continue
  minor="${tag%.*}"
  # Tags arrive newest-first, so the first hit for a series is its top patch.
  case " ${minors[*]:-} " in
    *" $minor "*) continue ;;
  esac
  minors+=("$minor")
  tags+=("$tag")
done < <(git tag -l 'v*' --sort=-version:refname)

latest_minor="${minors[0]:-}"

if [[ -z "$latest_minor" ]]; then
  echo "build-docs: no release tags found, publishing the working tree only"
fi

# ---------------------------------------------------------------------------
# The version list the selector renders from. Written before any build so every
# version, including older ones, lists the full set.
# ---------------------------------------------------------------------------

{
  printf '{\n'
  printf '  "latest": "%s",\n' "$latest_minor"
  printf '  "versions": [\n'
  printf '    { "label": "dev", "path": "/dev/", "state": "dev" }'
  for minor in ${minors[@]+"${minors[@]}"}; do
    state="old"
    [[ "$minor" == "$latest_minor" ]] && state="latest"
    printf ',\n    { "label": "%s", "path": "/%s/", "state": "%s" }' \
      "$minor" "$minor" "$state"
  done
  printf '\n  ]\n}\n'
} > "$SITE_DIR/data/versions.json"

echo "build-docs: versions -> dev ${minors[*]:-(none)}"

# build_site <source-dir> <destination> <url-path> <label> <state>
#
# The theme's banner slot only renders when `params.banner` is set, so it is
# supplied here for the versions that need a warning and left unset for the
# latest one. Without that, an empty custom banner partial would make the theme
# fall back to its own default "Welcome" message.
build_site() {
  local src="$1" dest="$2" url_path="$3" label="$4" state="$5"
  local banner=()

  if [[ "$state" != "latest" ]]; then
    banner=(HUGO_PARAMS_BANNER_MESSAGE=" ")
  fi

  echo "build-docs: $label ($state) -> ${url_path:-/}"
  # ${a[@]+"${a[@]}"} expands an empty array safely under `set -u` on bash 3.2.
  env ${banner[@]+"${banner[@]}"} \
    HUGO_PARAMS_VERSIONLABEL="$label" \
    HUGO_PARAMS_VERSIONSTATE="$state" \
    hugo --gc --minify \
      --source "$src" \
      --destination "$dest" \
      --baseURL "$BASE_URL/$url_path" \
      --quiet
}

# ---------------------------------------------------------------------------
# The working tree, published as the development version.
# ---------------------------------------------------------------------------

build_site "$SITE_DIR" "$OUT/dev" "dev/" "dev" "dev"

# With no releases to serve there would be nothing at the site root at all, so
# the working tree stands in for it — a fresh clone still builds a usable site.
if [[ -z "$latest_minor" ]]; then
  build_site "$SITE_DIR" "$OUT" "" "dev" "dev"
  echo "build-docs: done -> $OUT"
  exit 0
fi

# ---------------------------------------------------------------------------
# Each released minor series, content taken from its tag.
# ---------------------------------------------------------------------------

for i in "${!minors[@]}"; do  # non-empty: the no-release case returned above
  minor="${minors[$i]}"
  tag="${tags[$i]}"
  stage="$WORK/$minor"

  # Current tree minus content and build artifacts, then the tag's content.
  mkdir -p "$stage"
  tar -cf - --exclude content --exclude public --exclude resources -C "$SITE_DIR" . \
    | tar -xf - -C "$stage"

  if ! git archive "$tag" "$SITE_DIR/content" 2>/dev/null \
    | tar -xf - -C "$WORK" --strip-components=1; then
    echo "build-docs: $tag has no $SITE_DIR/content, skipping" >&2
    continue
  fi
  mv "$WORK/content" "$stage/content"

  # Some pages describe the project across releases rather than a single one,
  # so they always come from the working tree. A changelog frozen at its tag
  # cannot mention anything released after it, and every version links to it.
  for unversioned in "${UNVERSIONED[@]}"; do
    [[ -e "$SITE_DIR/content/$unversioned" ]] || continue
    rm -rf "$stage/content/$unversioned"
    mkdir -p "$(dirname "$stage/content/$unversioned")"
    cp -R "$SITE_DIR/content/$unversioned" "$stage/content/$unversioned"
  done

  state="old"
  [[ "$minor" == "$latest_minor" ]] && state="latest"

  build_site "$stage" "$OUT/$minor" "$minor/" "$minor" "$state"

  # The newest release is also what the site root serves.
  if [[ "$minor" == "$latest_minor" ]]; then
    build_site "$stage" "$WORK/root" "" "$minor" "latest"
    # Copy into place rather than building straight into $OUT, which already
    # holds the per-version directories Hugo would clean out.
    (cd "$WORK/root" && tar -cf - .) | (cd "$OUT" && tar -xf -)
  fi
done

echo "build-docs: done -> $OUT"
