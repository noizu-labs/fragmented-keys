#!/usr/bin/env bash
#
# sync-subtrees.sh — pull/push the vendored language prefixes via git-subtree.
#
# Source of truth for prefix -> url is the .subtrees manifest at the repo
# root (one line per prefix: <prefix> <url> <branch> squash). The prefix
# list is NEVER hardcoded here.
#
# Branch semantics (per the composite's convention):
#   pull  — pull FROM the upstream's `main`; if the upstream has no `main`
#           (php and java were imported from `master`), fall back to the
#           branch recorded in .subtrees. The branch actually used is
#           printed for every prefix.
#   push  — push TO the upstream's `develop`. If the upstream has no
#           `develop` branch, that prefix is SKIPPED (never created, never
#           falling back to main/master).
# Each remote is probed with `git ls-remote --heads <url>` first because
# the upstream default branches are NOT uniform across the family.
#
# Usage:
#   scripts/sync-subtrees.sh pull  [prefix ...]   # no prefix = all
#   scripts/sync-subtrees.sh push  [prefix ...]
#   scripts/sync-subtrees.sh status [prefix ...]  # probe + show plan
#   scripts/sync-subtrees.sh pull --dry-run ...   # print commands only
#
# Prefixes may be given with or without the trailing slash ("python" ==
# "python/"); they must exist in .subtrees or the script refuses.

set -euo pipefail

MANIFEST=".subtrees"
DRY_RUN=0
declare -a RESULTS_FAILED=()
declare -a RESULTS_SKIPPED=()
declare -a RESULTS_OK=()

die() { echo "sync-subtrees: $*" >&2; exit 2; }

usage() {
    sed -n '3,25p' "$0" | sed 's/^# \{0,1\}//'
    exit 2
}

# ---------------------------------------------------------------- args -----
[[ $# -ge 1 ]] || usage
ACTION=""
declare -a WANTED=()
for arg in "$@"; do
    case "$arg" in
        --dry-run) DRY_RUN=1 ;;
        -h|--help) usage ;;
        -*) die "unknown flag: $arg" ;;
        *)
            if [[ -z "$ACTION" ]]; then
                ACTION="$arg"
            else
                WANTED+=("${arg%/}/")   # normalize: ensure trailing slash
            fi
            ;;
    esac
done
case "$ACTION" in
    pull|push|status) ;;
    *) die "unknown subcommand '$ACTION' (want pull|push|status)" ;;
esac

cd "$(git rev-parse --show-toplevel)"
[[ -f "$MANIFEST" ]] || die "manifest $MANIFEST not found — run from the composite repo"

# A dirty tree makes any pull/push unsafe: the merge subtree creates could
# sweep unrelated staged/unstaged work into the commit. Refuse up front.
if [[ "$ACTION" != "status" ]] && [[ -n "$(git status --porcelain)" ]]; then
    die "working tree is dirty — commit or stash first (status subcommand still works)"
fi

# ------------------------------------------------------------ manifest -----
declare -a PREFIXES=() URLS=() BRANCHES=()
while read -r prefix url branch _flag; do
    case "$prefix" in ''|'#'*) continue ;; esac
    PREFIXES+=("$prefix"); URLS+=("$url"); BRANCHES+=("$branch")
done < "$MANIFEST"
[[ ${#PREFIXES[@]} -ge 1 ]] || die "no entries parsed from $MANIFEST"

# Resolve the requested prefix list against the manifest; refuse unknowns.
# SELECTED always holds manifest INDICES (so the main loop can look up
# prefix/url/branch together), whether "all" or an explicit list.
declare -a SELECTED=()
if [[ ${#WANTED[@]} -eq 0 ]]; then
    for idx in "${!PREFIXES[@]}"; do SELECTED+=("$idx"); done
else
    for w in "${WANTED[@]}"; do
        found=""
        for i in "${!PREFIXES[@]}"; do
            [[ "${PREFIXES[$i]}" == "$w" ]] && found=$i
        done
        [[ -n "$found" ]] || die "prefix '$w' is not in $MANIFEST (known: ${PREFIXES[*]})"
        SELECTED+=("$found")
    done
fi

# ------------------------------------------------------------- probing -----
# remote_has_branch <url> <branch> -> exit 0 if the remote has that head
remote_has_branch() {
    git ls-remote --heads "$1" "refs/heads/$2" | grep -q .
}

note_ok()      { RESULTS_OK+=("$1");      echo "  [ok]      $1"; }
note_skip()    { RESULTS_SKIPPED+=("$1"); echo "  [skip]    $1"; }
note_fail()    { RESULTS_FAILED+=("$1");  echo "  [FAILED]  $1" >&2; }

run_or_print() {
    if [[ $DRY_RUN -eq 1 ]]; then
        echo "  DRY-RUN: $*"
    else
        "$@"
    fi
}

# --------------------------------------------------------------- main ------
for i in "${SELECTED[@]}"; do
    prefix="${PREFIXES[$i]}"; url="${URLS[$i]}"; mbranch="${BRANCHES[$i]}"
    echo "== $prefix  ($url, manifest branch: $mbranch)"

    if ! probe_out="$(git ls-remote --heads "$url" 2>&1)"; then
        note_fail "$prefix — cannot reach remote: $url ($(echo "$probe_out" | head -1))"
        continue
    fi
    has_main=$(grep -c 'refs/heads/main$'    <<<"$probe_out" || true)
    has_dev=$( grep -c 'refs/heads/develop$' <<<"$probe_out" || true)
    has_mb=$(  grep -c "refs/heads/${mbranch}\$" <<<"$probe_out" || true)

    case "$ACTION" in
    pull)
        # ALWAYS --squash: every import in this composite was created with
        # `git subtree add --squash`, so the prefix's history here is a
        # single squashed commit. An unsquashed pull would try to graft the
        # upstream's FULL ancestry onto that squashed line, producing
        # spurious conflicts and duplicate content. Non-optional by design.
        if [[ "$has_main" -ge 1 ]]; then
            src="main"
        elif [[ "$has_mb" -ge 1 ]]; then
            src="$mbranch"
            echo "  note: no 'main' upstream; falling back to manifest branch '$mbranch'"
        else
            note_fail "$prefix — remote has neither 'main' nor manifest branch '$mbranch'; nothing to pull"
            continue
        fi
        if run_or_print git subtree pull --prefix="$prefix" "$url" "$src" --squash; then
            [[ $DRY_RUN -eq 1 ]] || note_ok "$prefix — pulled from $src (squash)"
        else
            note_fail "$prefix — git subtree pull from $src failed"
        fi
        ;;
    push)
        if [[ "$has_dev" -lt 1 ]]; then
            note_skip "$prefix — remote has no 'develop' branch (would have to be created; refusing. Needs a PR by hand or an owner to branch it.)"
            continue
        fi
        if run_or_print git subtree push --prefix="$prefix" "$url" develop; then
            [[ $DRY_RUN -eq 1 ]] || note_ok "$prefix — pushed to develop"
        else
            note_fail "$prefix — git subtree push to develop failed"
        fi
        ;;
    status)
        pull_src="-"; [[ "$has_main" -ge 1 ]] && pull_src="main" || { [[ "$has_mb" -ge 1 ]] && pull_src="$mbranch"; }
        push_to="develop (absent — push would skip)"
        [[ "$has_dev" -ge 1 ]] && push_to="develop"
        echo "  pull would use:   $pull_src"
        echo "  push would target: $push_to"
        RESULTS_OK+=("$prefix")
        ;;
    esac
done

# ------------------------------------------------------------- summary -----
echo
echo "================ summary ($ACTION) ================"
echo "ok:      ${#RESULTS_OK[@]}   ${RESULTS_OK[*]:-}"
echo "skipped: ${#RESULTS_SKIPPED[@]}   ${RESULTS_SKIPPED[*]:-}"
echo "failed:  ${#RESULTS_FAILED[@]}   ${RESULTS_FAILED[*]:-}"
[[ ${#RESULTS_FAILED[@]} -eq 0 ]] || exit 1
exit 0
