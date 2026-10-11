# PROJ-ARCH Summary — fragmented-keys-ts

Tag-versioned cache invalidation library. Increment a tag version to invalidate all composite keys that depend on it — no explicit deletion needed.

## Key Formula

`md5("{key}_{groupId}:t{tag1_name}:v{ver1}:t{tag2_name}:v{ver2}...")`

## Core Components (✓ = implemented, ○ = planned)

| Component | Status | Purpose |
|-----------|--------|---------|
| StandardTag | ✓ | Mutable versioned tag (+0.1 increment) |
| ConstantTag | ✓ | Immutable fixed-version tag |
| StandardKey | ✓ | Composite MD5 hash from tags |
| FragmentedKeyRing | ✓ | Template factory with 3-tier option merging + Proxy |
| CacheEntry\<T\> | ✓ | Rich metadata wrapper for cached data |
| Configuration | ✓ | Static singleton for defaults and prefixes |
| MemoryHandler | ✓ | In-memory Map backend (sync + async) |
| RedisHandler | ✓ | ioredis single-instance backend |
| RedisSentinelHandler | ○ | ioredis Sentinel HA |
| RedisShardRouter | ○ | Multi-shard routing |
| EncryptedHandler | ○ | Encrypt/decrypt decorator |
| Entry points | ○ | index, promise, callback, observable, sync |
| Adapters | ○ | Callback and Observable wrappers |

## Design Patterns

- 3-tier option merging: global → per-tag → per-call
- Cache prefix namespacing for multi-tenant environments
- Cross-language parity (6th implementation: PHP, Python, Java, Swift, Elixir)
