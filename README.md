# Fragmented Keys (composite)

**Fragmented Keys** is Noizu's cache-invalidation library. Cache keys are
composed from a key name, a group id, and an ordered set of independently
versioned "tags"; the composite key is the MD5 of that string. When any tag's
version is bumped, the composite hash changes, so every dependent cache entry
becomes unreachable instantly — with **no bulk deletes** and no tombstones
(stale entries simply age out under their normal TTL). The shared,
language-agnostic spec lives in [`spec/ALGORITHM.md`](spec/ALGORITHM.md).

The library exists as several per-language implementations scattered across
GitHub orgs. This repo is the **composite**: one directory per language,
assembled with `git subtree add --squash`, so each variant keeps its own
history behind its prefix while the whole family is browsable and releasable
in one place. Shared, composite-owned files (`README.md`, `.subtrees`,
`spec/`, `.gitignore`) live at the root and never inside a subtree prefix.

Destined for `noizu-labs/fragmented-keys` and a nested submodule slot at
`Portfolio/Libs/core/fragmented-keys`.

## Layout

| dir | upstream repo | language | upstream branch | imported SHA |
|-----|---------------|----------|-----------------|--------------|
| `php/` | `noizu/fragmented-keys` | PHP | `master` | `d2c9b01ddf53bb3e68df2f607584e90c61b96e2f` |
| `python/` | `noizu-labs/fragmented-keys-py` | Python | `main` | `362ec0500d567ee1e8528bebb9ce148080745617` |
| `python-django/` | `noizu-labs-scaffolding/fragmented_keys_django` | Python (Django) | `main` | `6a901f017aca6c164eeee79aef9d23f4515dc29d` |
| `swift/` | `noizu-labs/SwiftFragmentedKeys` | Swift | `main` | `cab5f6b9524e6436c9d58f7833b9e9088f69c8b3` |
| `java/` | `noizu-labs/fragmented-keys-4java` | Java | `master` | `d5f48b5c8f17172b172d863f7920e8133f111af6` |
| `go/` | `noizu-labs-scaffolding/fragmented-keys-go` | Go | `main` | `db54f7b5c0865b60b541a04baadfb8b1a2c7c3c8` |
| `elixir/` | *pending authorization* (`fragmented-keys-ex`, private) | Elixir | — | — |
| `nodejs/` | *pending authorization* (`fragmented-keys-nodejs`, private) | TypeScript | — | — |

Deliberately **excluded**: `lax-os/fragmented-keys-4java` — a stale fork of
the Java repo, not a separate implementation. Prefixes `elixir/` and `nodejs/`
are reserved for the two private variants once authorization clears.

The machine-readable counterpart of the table above is [`.subtrees`](.subtrees).

## Syncing

Pull upstream changes into a directory (run from the composite root):

    git subtree pull --prefix=<dir> <url> <branch> --squash

Push composite-side changes under a directory back upstream:

    git subtree push --prefix=<dir> <url> <branch>

> **Warning: EVERY pull MUST use `--squash`.** Every import in this repo was
> created with `--squash`, so the composite's subtree history has a single
> squashed commit per import rather than the upstream's full ancestry. An
> unsquashed pull would try to graft the complete upstream history onto that
> squashed line, producing spurious conflicts and duplicate content. The
> `.subtrees` manifest records the exact url, branch and squash flag per
> prefix so no future pull has to guess.
