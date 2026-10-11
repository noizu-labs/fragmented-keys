# Fragmented Keys — Algorithm (language-agnostic spec)

Composite-owned. Consolidated from the Noizu framework reference doc
(`skills/infrastructure-systems/noizu-frameworks/08-fragmented-keys.md`,
which documents the Elixir package `fragmented_keys` 0.1.0); every language
directory in this repo implements the same behaviour. Cross-implementation
agreement points are listed in [`CONFORMANCE.md`](CONFORMANCE.md).

## Composite key construction

A cache key is the **MD5 hash** of a string built from a key name, a group
id, and an ordered list of versioned tags:

```
Key = MD5("KeyName_GroupID:tTag1:vVer1:tTag2:vVer2...")
```

- `KeyName` — the human-readable key name (e.g. `Dashboard`).
- `_` separates the key name from the group id.
- `:` separates every subsequent field.
- Each tag contributes two fields: `t<TagName>` then `v<Version>`, in the
  order the tags were supplied — field order is significant.

## Version-based invalidation

- Every tag carries an independently stored **version counter** (a fresh
  version is seeded from microtime on reset).
- **Bumping a tag's version** (`increment`) changes the composite hash, so
  every cache entry built from that tag instantly resolves to a *different*
  key. Invalidation therefore needs **no bulk deletes and no tombstones** —
  stale entries simply expire naturally under their normal TTL.
- **Constant tags** hold a fixed version; `increment`/`reset` are no-ops.
  They model dimensions that never change (site-wide settings, environment,
  global constants).
- Tag types expected: *Standard* (versioned, mutable) and *Constant*
  (pinned); a tag is addressable by `(tag_name, instance)`.

## Backend-agnostic expectations

- All storage sits behind a **cache-handler abstraction** with at minimum:
  `get`, `set`, and a bulk `get_multi`. Reference handlers are an in-memory
  one (dev/test) and a Redis one (bulk ops via MGET/MSET); any backend
  implementing the handler contract is valid.
- **Bulk fetch optimization**: tags are grouped by handler so all version
  lookups needed to build one composite key resolve in a single round trip.
- **KeyRing templates** (named key definitions binding an ordered tag list to
  handler configuration, with per-tag global options) are the recommended
  reuse/factory layer on top of raw key construction.
