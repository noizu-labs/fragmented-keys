# Project Layout

Fragmented Keys — Elixir library for composite cache key management and invalidation via independently versioned tags.

```
fragmented-keys-ex/
├── lib/                            # Application source → [lib.md](lib.md)
│   ├── fragmented_keys/            #   Core modules
│   │   ├── cache_handler/          #     Backend implementations (Memory, Redis)
│   │   ├── tag/                    #     Tag types (Standard, Constant)
│   │   ├── cache_handler.ex        #     CacheHandler protocol definition
│   │   ├── configuration.ex        #     Global config via :persistent_term
│   │   ├── key.ex                  #     Composite cache key builder
│   │   ├── key_ring.ex             #     Key template factory
│   │   └── tag.ex                  #     Unified tag operations interface
│   └── fragmented_keys.ex          #   Top-level module and docs
├── test/                           # Test suites
│   ├── fragmented_keys_test.exs    #   Main test file
│   └── test_helper.exs             #   Test bootstrap
├── docs/                           # Documentation
│   ├── arch/                       #   Architecture detail docs
│   │   ├── cache-handler.md        #     Cache handler protocol details
│   │   └── invalidation.md         #     Key invalidation strategy details
│   ├── layout/                     #   Layout detail docs
│   │   └── lib.md                  #     lib/ detailed breakdown
│   ├── PROJ-ARCH.md                #   Architecture overview
│   ├── PROJ-ARCH.summary.md        #   Architecture quick reference
│   ├── PROJ-LAYOUT.md              #   This file
│   └── PROJ-LAYOUT.summary.md      #   Layout quick reference
├── .claude/                        # Claude Code configuration
│   └── settings.local.json
├── .formatter.exs                  # Mix formatter config
├── .gitignore                      # Git ignore rules
├── .tool-versions                  # asdf — erlang 28.4, elixir 1.19.5, nodejs 25.8.1
├── LICENSE                         # MIT license
├── mix.exs                         # Project definition and dependencies
├── mix.lock                        # Locked dependency versions
└── README.md                       # Project entry point
```

## Key Files Requiring Setup

| File | Action |
|------|--------|
| `.tool-versions` | Install runtimes via `asdf install` |
| `mix.exs` | `:redix` is optional — add if Redis backend needed |

## External Links

- [fragmented-keys](https://github.com/noizu-labs/fragmented-keys) — PHP (original)
- [fragmented-keys-py](https://github.com/noizu-labs/fragmented-keys-py) — Python sibling library
- [fragmented-keys-4java](https://github.com/noizu-labs/fragmented-keys-4java) — Java sibling library
