# Project Layout Summary

```
fragmented-keys-ex/
├── lib/
│   ├── fragmented_keys/
│   │   ├── cache_handler/          # Memory, Redis implementations
│   │   ├── tag/                    # Standard, Constant tag types
│   │   ├── cache_handler.ex        # CacheHandler protocol
│   │   ├── configuration.ex        # Global config
│   │   ├── key.ex                  # Composite key builder
│   │   ├── key_ring.ex             # Key template factory
│   │   └── tag.ex                  # Tag operations interface
│   └── fragmented_keys.ex          # Top-level module
├── test/                           # Tests
├── docs/                           # Documentation
│   ├── arch/                       # Architecture details
│   └── layout/                     # Layout details
├── .formatter.exs
├── .tool-versions                  # asdf runtimes
├── LICENSE                         # MIT
├── mix.exs                         # Project config
├── mix.lock
└── README.md
```
