defmodule FragmentedKeys do
  @moduledoc """
  Fragmented Key Management and Invalidation Library.

  Compose cache keys from multiple independently versioned "tags".
  When a tag version increments, all dependent cache keys produce
  different hashes — no bulk deletes needed.

  ## Quick start

      handler = FragmentedKeys.CacheHandler.Memory.new()
      FragmentedKeys.Configuration.set_default_cache_handler(handler)

      tag_user = FragmentedKeys.Tag.Standard.new("User", "42")
      tag_city = FragmentedKeys.Tag.Standard.new("City", "chicago")

      key = FragmentedKeys.Key.new("Dashboard", [tag_user, tag_city])
      key_str = FragmentedKeys.Key.get_key_str(key)

  ## Using KeyRing

      ring = FragmentedKeys.KeyRing.new(
        cache_handlers: %{"memory" => handler},
        default_cache_handler: "memory"
      )

      ring = FragmentedKeys.KeyRing.define_key(ring, "Users", ["universe", "planet", "city"])
      key_obj = FragmentedKeys.KeyRing.get_key_obj(ring, "Users", ["MilkyWay", "Earth", "Chicago"])
      key_str = FragmentedKeys.Key.get_key_str(key_obj)
  """
end
