# Cache Handler Protocol

## Protocol Definition

`FragmentedKeys.CacheHandler` is an Elixir protocol with four functions:

| Function | Signature | Purpose |
|----------|-----------|---------|
| `group_name/1` | `handler → String.t()` | Unique identifier for handler type (used for bulk grouping) |
| `get/2` | `handler, key → String.t() \| nil` | Fetch a single value |
| `set/4` | `handler, key, value, ttl → :ok` | Store a value (ttl in seconds, or nil) |
| `get_multi/2` | `handler, [key] → %{key => value}` | Bulk fetch; returns only found entries |

## Implementations

### Memory (`CacheHandler.Memory`)

- Backed by an Elixir `Agent` holding a plain map
- Each `new/0` call creates an independent cache instance (separate Agent process)
- TTL parameter is accepted but ignored (entries persist until `clear/0` or process death)
- Group name: `"MemoryHandler"`
- Use case: testing, development, short-lived caches

### Redis (`CacheHandler.Redis`)

- Backed by a `Redix` connection (pid or named process)
- Wraps `new/1` around an existing Redix connection — does not manage the connection lifecycle
- TTL is implemented via `SETEX`; nil TTL uses plain `SET`
- Bulk fetch uses `MGET`
- Group name: `"RedisHandler"`
- Use case: production, shared caches across nodes
- Requires optional dependency: `{:redix, "~> 1.5", optional: true}`

## Handler Resolution

Tags resolve their handler in order:
1. Explicit handler set on the tag struct (`:handler` field)
2. Global default from `Configuration.get_default_cache_handler/0`

`KeyRing` manages its own handler registry and injects handlers into tags during creation.

## Extending

To add a new backend, define a struct and implement the protocol:

```elixir
defmodule MyApp.CacheHandler.Memcached do
  defstruct [:conn]
end

defimpl FragmentedKeys.CacheHandler, for: MyApp.CacheHandler.Memcached do
  def group_name(_), do: "MemcachedHandler"
  def get(%{conn: c}, key), do: ...
  def set(%{conn: c}, key, value, ttl), do: ...
  def get_multi(%{conn: c}, keys), do: ...
end
```
