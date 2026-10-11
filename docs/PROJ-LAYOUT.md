# Project Layout

```
fragmented-keys-py/
├── src/                            # Library source
│   └── fragmented_keys/            # Main package
│       ├── __init__.py             #   Public API — re-exports all user-facing classes
│       ├── protocols.py            #   CacheHandler, Tag, Key, KeyRing protocols
│       ├── configuration.py        #   Global defaults (handler, prefix)
│       ├── key_ring.py             #   FragmentedKeyRing — key template factory
│       ├── py.typed                #   PEP 561 typed-package marker
│       ├── cache_handler/          #   Cache backend implementations
│       │   ├── __init__.py
│       │   ├── redis_handler.py    #     RedisHandler — primary backend (redis-py)
│       │   └── memory.py           #     MemoryHandler — in-memory dict for testing
│       ├── key/                    #   Composite cache key
│       │   ├── __init__.py
│       │   └── standard.py         #     StandardKey — MD5 hash of versioned tag composition
│       └── tag/                    #   Versioned tag-instance pairs
│           ├── __init__.py
│           ├── base.py             #     BaseTag — version storage/retrieval, cache delegation
│           ├── standard.py         #     StandardTag — incrementable version
│           └── constant.py         #     ConstantTag — fixed version, mutation no-ops
├── tests/                          # Test suite (pytest)
│   ├── __init__.py
│   ├── conftest.py                 #   Fixtures: memory_handler, auto config reset
│   └── test_fragmented_keys.py     #   26 tests — tags, keys, key ring, delegation
├── docs/                           # Documentation
│   └── PROJ-LAYOUT.md             #   This file
├── pyproject.toml                  # Package metadata, dependencies, build config
├── uv.lock                         # Locked dependency versions
├── .python-version                 # Python 3.13
├── .gitignore                      # Ignores .venv, dist, __pycache__, etc.
├── .envrc                          # direnv configuration
├── LICENSE                         # MIT
└── README.md                       # Usage guide, API reference, installation
```
