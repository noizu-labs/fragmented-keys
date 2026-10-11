# Project Architecture — Summary

## Overview

Cache invalidation library that composes cache keys from independently versioned tag-instance pairs. Incrementing a tag version changes all dependent key hashes, causing cache misses without deleting entries. Python port of the PHP fragmented-keys library, using Redis.

## Core Components

- **Configuration** — Global default cache handler and key prefix
- **CacheHandler** — Protocol: `get`, `set`, `get_multi`
- **StandardTag** — Incrementable cached version (+0.1 per increment)
- **ConstantTag** — Fixed version, skipped in bulk cache fetches
- **StandardKey** — Composes tags into MD5-hashed cache key with bulk version fetching
- **FragmentedKeyRing** — Template factory for defining and instantiating keys with merged options
- **RedisHandler** — Redis backend (`mget`, `setex`)
- **MemoryHandler** — In-memory dict for testing

## Key Generation Flow

1. Group tags by cache handler, skip ConstantTags from bulk fetch
2. Bulk-fetch versions via `get_multi()` per handler group
3. Build raw string: `"{key}_{groupId}:t{tag}:v{version}..."`
4. Return MD5 hex digest

## Design Decisions

- Redis over Memcache (modern, `mget`/`setex`)
- Protocols over ABCs (duck typing for custom handlers)
- Orphan invalidation (old keys expire via TTL, no bulk deletes)
- MD5 hashing (fixed-length, backend-safe)

## Stack

Python 3.13+, redis-py >= 5.0, uv build system, pytest, PEP 561 typed
