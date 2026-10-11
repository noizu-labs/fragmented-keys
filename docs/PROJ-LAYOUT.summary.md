# Project Layout — Summary

```
fragmented-keys-py/
├── src/fragmented_keys/            # Library source
│   ├── cache_handler/              #   Redis + Memory backends
│   ├── key/                        #   Composite cache key
│   ├── tag/                        #   Versioned tag-instance pairs
│   ├── protocols.py                #   Protocol definitions
│   ├── configuration.py            #   Global config
│   └── key_ring.py                 #   Key template factory
├── tests/                          # pytest suite (26 tests)
├── docs/                           # Documentation
├── pyproject.toml                  # Package + build config
└── README.md                       # Usage guide
```
