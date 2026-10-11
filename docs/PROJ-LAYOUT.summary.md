# PROJ-LAYOUT Summary — fragmented-keys-ts

```
fragmented-keys-ts/
├── docs/                       # Project documentation
│   ├── PROJ-ARCH.md            # Architecture overview
│   ├── PROJ-ARCH.summary.md    # Architecture quick reference
│   ├── PROJ-LAYOUT.md          # Directory structure guide
│   └── PROJ-LAYOUT.summary.md  # This file
├── src/
│   ├── configuration.ts        # Global config singleton
│   ├── interfaces/             # Contracts: CacheHandler, Tag, Key, KeyRing, Encryption, types
│   ├── cache-entry/            # CacheEntry<T> wrapper
│   ├── cache-handler/          # Memory + Redis handlers
│   ├── tag/                    # BaseTag, StandardTag, ConstantTag
│   ├── key/                    # StandardKey — composite MD5
│   ├── key-ring/               # FragmentedKeyRing — factory + Proxy accessor
│   ├── adapters/               # (planned)
│   └── encryption/             # (planned)
├── tests/
│   ├── setup.ts                # Test lifecycle
│   ├── helpers/                # (planned)
│   ├── unit/                   # 7 test files, no external deps
│   └── integration/            # Redis handler tests (Docker required)
├── .envrc                      # direnv — loads environment
├── package.json
├── tsconfig.json
├── tsup.config.ts
├── vitest.config.ts
├── vitest.integration.config.ts
├── biome.json
├── docker-compose.yml
├── LICENSE
└── README.md
```
