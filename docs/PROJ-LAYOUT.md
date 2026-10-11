# PROJ-LAYOUT.md — fragmented-keys-ts

## Directory Structure

```
fragmented-keys-ts/
├── package.json                        # Multi-export package config (npm/yarn/pnpm/bun)
├── tsconfig.json                       # TypeScript compiler config
├── tsup.config.ts                      # Dual ESM/CJS build, multiple entry points
├── vitest.config.ts                    # Unit test config
├── vitest.integration.config.ts        # Integration test config (Redis required)
├── biome.json                          # Linter/formatter config
├── docker-compose.yml                  # Redis + Sentinel test infrastructure
├── LICENSE                             # MIT
├── .gitignore                          # Excludes node_modules, dist, coverage, etc.
│
├── .envrc                              # direnv — loads environment
│
├── docs/
│   ├── PROJ-ARCH.md                    # Architecture overview
│   ├── PROJ-ARCH.summary.md            # Architecture quick reference
│   ├── PROJ-LAYOUT.md                  # This file
│   └── PROJ-LAYOUT.summary.md          # Layout quick reference
│
├── src/
│   ├── configuration.ts                # Static singleton: default handler, global prefix, cache prefix
│   │
│   ├── interfaces/
│   │   ├── cache-handler.ts            # CacheHandler — async cache backend contract
│   │   ├── cache-handler-sync.ts       # SyncCacheHandler — sync variant
│   │   ├── tag.ts                      # Tag — version management contract
│   │   ├── key.ts                      # Key — composite cache key contract
│   │   ├── key-ring.ts                 # KeyRing — template factory contract
│   │   ├── encryption.ts              # EncryptionStrategy — encrypt/decrypt contract
│   │   └── types.ts                    # Shared: TagType, TagOptions, KeyParam, CacheEntryStatus, etc.
│   │
│   ├── cache-entry/
│   │   └── cache-entry.ts              # CacheEntry<T> wrapper: status, meta, timestamps, type info
│   │
│   ├── cache-handler/
│   │   ├── memory-handler.ts           # In-memory Map (sync + async)
│   │   └── redis-handler.ts            # ioredis single-instance backend (async)
│   │
│   ├── tag/
│   │   ├── base-tag.ts                 # Abstract: lazy version loading, cache persistence
│   │   ├── standard-tag.ts             # Mutable: increment (+0.1), reset (Date.now())
│   │   └── constant-tag.ts             # Immutable: fixed version, no-op mutations
│   │
│   ├── key/
│   │   └── standard-key.ts             # Composite MD5 from tags, bulk-fetch optimization
│   │
│   ├── key-ring/
│   │   └── fragmented-key-ring.ts      # Factory, 3-tier merging, ES6 Proxy dynamic accessor
│   │
│   ├── adapters/                       # (planned) Callback/Observable adapters
│   │
│   └── encryption/                     # (planned) AES-256-GCM, noop strategies
│
└── tests/
    ├── setup.ts                        # Global beforeEach: Configuration.reset()
    ├── helpers/                         # (planned) Redis lifecycle, seed data
    ├── unit/                           # No external deps required
    │   ├── configuration.test.ts
    │   ├── memory-handler.test.ts
    │   ├── cache-entry.test.ts
    │   ├── standard-tag.test.ts
    │   ├── constant-tag.test.ts
    │   ├── standard-key.test.ts
    │   └── fragmented-key-ring.test.ts
    └── integration/                    # Requires Docker Redis
        └── redis-handler.test.ts
├── README.md                           # Project entry point
```

## Module Dependency Graph

```
configuration.ts ─────────► interfaces/cache-handler
cache-handler/* ──────────► interfaces/cache-handler, configuration
tag/* ────────────────────► interfaces/tag, cache-handler, configuration
key/* ────────────────────► interfaces/key, tag/*
key-ring/* ───────────────► interfaces/key-ring, tag/*, key/*, cache-handler/*
```

## Planned (Not Yet Implemented)

The following are defined in `package.json` exports but not yet created:

- **Entry points**: `index.ts`, `promise.ts`, `callback.ts`, `observable.ts`, `sync.ts`
- **Cache handlers**: `redis-sentinel-handler.ts`, `redis-shard-router.ts`, `encrypted-handler.ts`
- **Adapters**: `callback-adapter.ts`, `observable-adapter.ts`
- **Encryption**: `aes-256-gcm-strategy.ts`, `noop-strategy.ts`
- **Test helpers**: `redis-setup.ts`, `seed-data.ts`
- **Integration tests**: Sentinel, sharding, encryption tests

## Build Output

```
dist/
├── index.js / index.cjs / index.d.ts / index.d.cts
├── promise.js / promise.cjs / promise.d.ts / promise.d.cts
├── callback.js / callback.cjs / callback.d.ts / callback.d.cts
├── observable.js / observable.cjs / observable.d.ts / observable.d.cts
└── sync.js / sync.cjs / sync.d.ts / sync.d.cts
```
