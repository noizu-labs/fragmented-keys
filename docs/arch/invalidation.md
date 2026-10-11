# Key Invalidation Strategy

## Concept

Traditional cache invalidation deletes or overwrites stale entries. FragmentedKeys takes a different approach: cache keys are **derived from tag versions**, so changing a version produces a new key. Old entries are never explicitly deleted — they simply expire via TTL or LRU eviction.

## How It Works

1. A `Key` is composed of a base name, a group ID, and an ordered list of tags
2. Each tag contributes its name and current version to the key string
3. The raw string is hashed with MD5 to produce a fixed-length cache key

```
Raw:  "Dashboard_:tUser_42DefaultPrefix:v1733000000000.0:tCity_chicagoDefaultPrefix:v1733000000000.0"
Hash: "a3f2b8c1d4e5f6a7b8c9d0e1f2a3b4c5"
```

4. When `Tag.increment/1` is called on the User tag, its version changes
5. The same `Key` now produces a different MD5 — cache miss, fresh data populated

## Tag Types

### Standard Tags
- Version stored in cache backend (via `CacheHandler`)
- `increment/1` bumps version by 0.1
- `reset_tag_version/1` sets version to current system time in milliseconds
- First access auto-initializes version to system time if not present

### Constant Tags
- Version is fixed at creation (default 1.0)
- All mutation operations (`increment`, `reset`, `set_version`) are no-ops
- Useful for incorporating static dimensions into keys (e.g., API version, schema version)

## Bulk Version Resolution

When a `Key` resolves its string, it groups tags by their handler's `group_name` and issues one `get_multi` call per group. This avoids N individual cache lookups for a key with N tags.

```
Tags: [User:42 (memory), City:chicago (memory), ApiVersion (constant)]
→ Group "MemoryHandler": bulk fetch [User:42, City:chicago]
→ Constant: skipped (version is local)
```
