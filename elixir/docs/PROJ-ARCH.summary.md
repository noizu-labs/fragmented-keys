# Architecture Summary

**FragmentedKeys** — Elixir library for composite cache key management via independently versioned tags.

## Core Concept
Cache keys are derived from tag versions. Incrementing a tag produces a new key hash — no bulk deletes needed. Old entries expire naturally.

## Components
- **Key**: Composes MD5 hash from base name + tag versions
- **Tag.Standard**: Cache-backed, incrementable version
- **Tag.Constant**: Fixed version, mutation is no-op
- **KeyRing**: Factory for key templates with named tag slots
- **CacheHandler**: Protocol for storage backends (Memory, Redis)
- **Configuration**: Global defaults via :persistent_term

## Design Choices
- Protocol-based dispatch for cache handlers (no atom registry)
- No supervision tree — uses :persistent_term and Agent
- Bulk version fetching via get_multi, grouped by handler
- MD5 hashing for cache key generation
