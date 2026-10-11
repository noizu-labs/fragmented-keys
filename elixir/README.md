# FragmentedKeys

Fragmented Key Management and Invalidation Library for Elixir.

Compose cache keys from multiple independently versioned "tags". When a tag
version increments, all dependent cache keys produce different hashes — no
bulk deletes needed.

Elixir port of [fragmented-keys-py](https://github.com/noizu-labs/fragmented-keys-py).

## Installation

Add `fragmented_keys` to your list of dependencies in `mix.exs`:

```elixir
def deps do
  [
    {:fragmented_keys, "~> 0.1.0"}
  ]
end
```

For Redis support, also add `:redix`:

```elixir
{:redix, "~> 1.5"}
```

## Quick Start

```elixir
# Set up a cache handler
handler = FragmentedKeys.CacheHandler.Memory.new()
FragmentedKeys.Configuration.set_default_cache_handler(handler)

# Create tags
tag_user = FragmentedKeys.Tag.Standard.new("User", "42")
tag_city = FragmentedKeys.Tag.Standard.new("City", "chicago")

# Build a composite cache key
key = FragmentedKeys.Key.new("Dashboard", [tag_user, tag_city])
key_str = FragmentedKeys.Key.get_key_str(key)
# => "a1b2c3..." (MD5 hex digest)

# Increment a tag to invalidate all dependent keys
FragmentedKeys.Tag.increment(tag_user)

# New keys with the same identity will now produce a different hash
tag_user2 = FragmentedKeys.Tag.Standard.new("User", "42")
key2 = FragmentedKeys.Key.new("Dashboard", [tag_user2, tag_city])
FragmentedKeys.Key.get_key_str(key2) != key_str
# => true
```

## Using KeyRing

`KeyRing` provides a template factory for defining reusable key patterns:

```elixir
ring = FragmentedKeys.KeyRing.new(
  cache_handlers: %{"memory" => handler},
  default_cache_handler: "memory"
)

# Define a key template with tag parameters
ring = FragmentedKeys.KeyRing.define_key(ring, "Users", ["universe", "planet", "city"])

# Create a key from the template
key_obj = FragmentedKeys.KeyRing.get_key_obj(ring, "Users", ["MilkyWay", "Earth", "Chicago"])
key_str = FragmentedKeys.Key.get_key_str(key_obj)
```

### Constant Tags

Use constant tags for fixed values that should never change:

```elixir
ring = FragmentedKeys.KeyRing.new(
  global_tag_options: %{"site" => %{"type" => "constant", "version" => 1.0}},
  cache_handlers: %{"memory" => handler},
  default_cache_handler: "memory"
)
```

### Per-Tag Handler Overrides

Different tags can use different cache backends:

```elixir
ring = FragmentedKeys.KeyRing.define_key(ring, "Mixed", [
  "user",
  %{"tag" => "city", "cache_handler" => "alt"}
])
```

## How It Works

1. Each **tag** (e.g., `User:42`, `City:chicago`) has a version stored in a cache backend
2. A **key** composes multiple tags into a single cache key by hashing the concatenation of all tag names and versions
3. When a tag's version is incremented, any key that includes that tag will resolve to a different hash
4. Old cache entries expire naturally via TTL — no explicit deletion required

## Cache Handlers

- `FragmentedKeys.CacheHandler.Memory` — Agent-backed in-memory store (testing/development)
- `FragmentedKeys.CacheHandler.Redis` — Redis backend via Redix (production)

## License

MIT — see [LICENSE](LICENSE).
