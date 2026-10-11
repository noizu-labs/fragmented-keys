# lib/ — Application Source

```
lib/
├── fragmented_keys/
│   ├── cache_handler/
│   │   ├── memory.ex               # In-memory cache via Agent (testing/dev)
│   │   └── redis.ex                # Redis cache via Redix (optional dep)
│   ├── tag/
│   │   ├── constant.ex             # Fixed-version tag (never changes)
│   │   └── standard.ex             # Versioned tag (stored in cache, incrementable)
│   ├── cache_handler.ex            # CacheHandler protocol — get/set/get_multi/group_name
│   ├── configuration.ex            # Global defaults via :persistent_term (no supervision needed)
│   ├── key.ex                      # Composite key — MD5 hash of base name + tag versions
│   ├── key_ring.ex                 # Factory — define key templates, resolve tags from definitions
│   └── tag.ex                      # Unified interface dispatching across Standard/Constant tags
└── fragmented_keys.ex              # Top-level module with @moduledoc and usage examples
```

## Module Relationships

- **CacheHandler protocol** (`cache_handler.ex`) — implemented by `Memory` and `Redis` structs
- **Tag types** — `Standard` (cache-backed, incrementable) and `Constant` (fixed version, no-ops on mutation)
- **Tag module** (`tag.ex`) — dispatches operations based on tag struct type; handles version fetch, increment, reset
- **Key** (`key.ex`) — composes a cache key string from a base name + ordered tag versions, bulk-fetches versions per handler group
- **KeyRing** (`key_ring.ex`) — higher-level factory: define key templates with named tag slots, then instantiate with concrete values
- **Configuration** (`configuration.ex`) — stores global defaults (cache handler, prefix) in `:persistent_term`
