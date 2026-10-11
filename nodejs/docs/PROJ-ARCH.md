# PROJ-ARCH.md — fragmented-keys-ts

## Purpose

Tag-versioned cache invalidation library. Instead of deleting stale cache entries, increment a tag's version — all composite keys depending on that tag resolve to a new MD5 hash, producing cache misses. Old entries expire naturally via TTL.

## Key Formula

```
md5("{key}_{groupId}:t{tag1_name}:v{ver1}:t{tag2_name}:v{ver2}...")
```

## Core Concepts

### Tags
Versioned identifiers stored in cache. Each tag has a `name`, `instance`, and `version`. When a tag is incremented, all composite keys using it change.

- **StandardTag**: Mutable. Version increments by +0.1. Reset seeds with `Date.now()`.
- **ConstantTag**: Immutable. Fixed version (default 1.0). All mutations are no-ops.

### Keys
Composite MD5 hashes built from multiple tags. A `StandardKey` gathers all its tags' versions, composes a raw string, and hashes it.

### KeyRings
Template factories that define reusable key structures. Define once with `defineKey()`, instantiate with `getKeyObj()` passing tag instance values.

### CacheHandlers
Backend abstraction for reading/writing tag versions. All methods are async (Promise-based).

- **MemoryHandler**: In-memory Map for testing (also supports sync interface).
- **RedisHandler**: ioredis single-instance backend.
- *(planned)* **RedisSentinelHandler**: ioredis with Sentinel discovery for HA.
- *(planned)* **RedisShardRouter**: Routes tags/keys to different Redis connections.
- *(planned)* **EncryptedHandler**: Decorator that encrypts/decrypts values through any handler.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│                  FragmentedKeyRing                    │
│  (template factory, 3-tier option merging, Proxy)    │
└──────────┬──────────────────────────────┬────────────┘
           │ creates                      │ creates
           ▼                              ▼
    ┌──────────────┐              ┌──────────────┐
    │  StandardKey  │◄─────────── │   BaseTag     │
    │  (MD5 hash)   │  composed   │  ┌──────────┐ │
    └──────────────┘  from tags   │  │ Standard │ │
                                  │  │ Constant │ │
                                  │  └──────────┘ │
                                  └──────┬────────┘
                                         │ reads/writes via
                                         ▼
                               ┌──────────────────┐
                               │   CacheHandler    │
                               │  ┌──────────────┐ │
                               │  │ Memory     ✓ │ │
                               │  │ Redis      ✓ │ │
                               │  │ Sentinel   ○ │ │
                               │  │ ShardRouter○ │ │
                               │  │ Encrypted  ○ │ │
                               │  └──────────────┘ │
                               └──────────────────┘

┌──────────────────┐
│  Configuration   │  (static singleton: default handler, prefixes)
└──────────────────┘
```

## Multi-Style API (Planned)

The library will support 4 invocation styles via sub-entry points. Entry points are defined in `package.json` but not yet implemented:

| Entry Point | Style | Status | Notes |
|---|---|---|---|
| `fragmented-keys` | Core | ○ | Types, interfaces, Configuration, MemoryHandler |
| `fragmented-keys/promise` | Promise | ○ | Default async style, all classes |
| `fragmented-keys/callback` | Callback | ○ | Node-style `(err, result)` wrapper |
| `fragmented-keys/observable` | Observable | ○ | RxJS wrapper (optional peer dep) |
| `fragmented-keys/sync` | Synchronous | ○ | MemoryHandler only |

Core `CacheHandler` is Promise-based. Adapters wrap for other styles.

## 3-Tier Option Merging

When creating tags via KeyRing, options merge in priority order (highest wins):

1. **Global options** — `globalOptions` in KeyRing constructor
2. **Per-tag options** — `globalTagOptions[tagName]`
3. **Per-call overrides** — inline dict in `defineKey()` params

## Cache Prefix / Namespacing

`Configuration.setCachePrefix("myapp:staging:")` prepends to every key on every handler read/write. Enables multi-tenant or multi-environment sharing of a Redis cluster.

## Encrypted Cache (Planned)

`EncryptedHandler` is a decorator wrapping any `CacheHandler`. Values are encrypted before `set()` and decrypted after `get()`. Users provide an `EncryptionStrategy` (built-in: AES-256-GCM via `node:crypto`).

## CacheEntry\<T\> Wrapper

Rich metadata wrapper for cached application data:

```typescript
interface CacheEntry<T = unknown> {
  cached: number;                        // Unix epoch ms
  status: "green" | "empty" | "error" | "unknown";
  meta: Record<string, unknown>;         // Error details, telemetry
  dataType?: string;                     // Stored type name
  dataTypeVersion?: string;              // For up/downcasting
  entry: T;                              // The actual data
}
```

## Sharding (Planned)

`RedisShardRouter` routes operations to different Redis handlers based on `shardKey` in tag options. `getMulti` fans out by shard and merges results.

## Cross-Language Parity

This is the 6th implementation (PHP, Python, Java, Swift, Elixir). Core tag/key/ring semantics are identical. TypeScript additions: async API, Sentinel, sharding, encryption, CacheEntry wrapper, multi-style adapters.
