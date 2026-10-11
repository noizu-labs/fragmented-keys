defmodule FragmentedKeysTest do
  @moduledoc """
  Tests mirroring the Python FragmentedKeysTest suite, using MemoryHandler.
  """

  use ExUnit.Case, async: false

  alias FragmentedKeys.{Configuration, Key, KeyRing, Tag}
  alias FragmentedKeys.CacheHandler.Memory
  alias FragmentedKeys.Tag.{Constant, Standard}

  setup do
    Configuration.reset()
    handler = Memory.new()
    Configuration.set_default_cache_handler(handler)
    Configuration.set_global_prefix("Test")
    %{handler: handler}
  end

  # ---------------------------------------------------------------------------
  # Standard Tag tests
  # ---------------------------------------------------------------------------

  describe "StandardTag" do
    test "same version without increment", _ctx do
      tag = Standard.new("User", "42")
      v1 = Tag.get_tag_version(tag)
      v2 = Tag.get_tag_version(tag)
      assert v1 == v2
    end

    test "different version after increment", _ctx do
      tag = Standard.new("User", "42")
      v1 = Tag.get_tag_version(tag)
      tag = Tag.increment(tag)
      v2 = Tag.get_tag_version(tag)
      assert v1 != v2
    end

    test "different instances have different tag names", _ctx do
      tag_a = Standard.new("User", "1")
      tag_b = Standard.new("User", "2")
      assert Tag.get_tag_name(tag_a) != Tag.get_tag_name(tag_b)
    end

    test "different tag names", _ctx do
      tag_a = Standard.new("User", "1")
      tag_b = Standard.new("City", "1")
      assert Tag.get_tag_name(tag_a) != Tag.get_tag_name(tag_b)
    end

    test "increment changes version by 0.1", _ctx do
      tag = Standard.new("User", "1")
      v1 = Tag.get_tag_version(tag)
      tag = Tag.increment(tag)
      v2 = Tag.get_tag_version(tag)
      assert_in_delta v2, v1 + 0.1, 1.0e-9
    end

    test "increment one does not affect other", _ctx do
      tag_a = Standard.new("User", "1")
      tag_b = Standard.new("User", "2")
      v_b_before = Tag.get_tag_version(tag_b)
      _tag_a = Tag.increment(tag_a)
      # Re-read from cache
      tag_b2 = Standard.new("User", "2")
      v_b_after = Tag.get_tag_version(tag_b2)
      assert v_b_before == v_b_after
    end

    test "increment different tag has no effect", _ctx do
      tag_user = Standard.new("User", "1")
      tag_city = Standard.new("City", "1")
      v_city_before = Tag.get_tag_version(tag_city)
      _tag_user = Tag.increment(tag_user)
      tag_city2 = Standard.new("City", "1")
      v_city_after = Tag.get_tag_version(tag_city2)
      assert v_city_before == v_city_after
    end
  end

  # ---------------------------------------------------------------------------
  # Constant Tag tests
  # ---------------------------------------------------------------------------

  describe "ConstantTag" do
    test "constant version is fixed", _ctx do
      tag = Constant.new("Site", "main", version: 5.0)
      assert Tag.get_tag_version(tag) == 5.0
    end

    test "increment is noop", _ctx do
      tag = Constant.new("Site", "main", version: 5.0)
      tag = Tag.increment(tag)
      assert Tag.get_tag_version(tag) == 5.0
    end

    test "reset is noop", _ctx do
      tag = Constant.new("Site", "main", version: 5.0)
      tag = Tag.reset_tag_version(tag)
      assert Tag.get_tag_version(tag) == 5.0
    end

    test "set version is noop", _ctx do
      tag = Constant.new("Site", "main", version: 5.0)
      tag = Tag.set_tag_version(tag, 99.0, true)
      assert Tag.get_tag_version(tag) == 5.0
    end
  end

  # ---------------------------------------------------------------------------
  # Key tests
  # ---------------------------------------------------------------------------

  describe "Key" do
    test "key consistent without increment", _ctx do
      tag = Standard.new("User", "42")
      key1 = Key.new("Profile", [tag])
      k1 = Key.get_key_str(key1)

      tag2 = Standard.new("User", "42")
      key2 = Key.new("Profile", [tag2])
      k2 = Key.get_key_str(key2)

      assert k1 == k2
    end

    # Contract: get_key_str/1 always reflects current cache state. Rendering a
    # key, then having a *different* tag object invalidate the same identity,
    # must produce a new string on the next render of the SAME key object.
    # (Guards against memoizing fetched versions onto the key's tags.)
    test "re-render reflects external invalidation", _ctx do
      tag = Standard.new("User", "42")
      key = Key.new("Profile", [tag])
      k1 = Key.get_key_str(key)

      # Independent object bumps the version in the shared cache.
      Tag.increment(Standard.new("User", "42"))

      k2 = Key.get_key_str(key)
      assert k1 != k2
    end

    test "key changes after increment", _ctx do
      tag = Standard.new("User", "42")
      key1 = Key.new("Profile", [tag])
      k1 = Key.get_key_str(key1)

      Tag.increment(tag)

      tag2 = Standard.new("User", "42")
      key2 = Key.new("Profile", [tag2])
      k2 = Key.get_key_str(key2)

      assert k1 != k2
    end

    test "key with multiple tags", _ctx do
      tag_user = Standard.new("User", "1")
      tag_city = Standard.new("City", "chicago")
      key = Key.new("Dashboard", [tag_user, tag_city])
      k = Key.get_key_str(key)
      assert is_binary(k)
      assert String.length(k) == 32
    end

    test "key with constant tags is stable", _ctx do
      tag = Constant.new("Site", "main", version: 1.0)
      key1 = Key.new("Home", [tag])
      key2 = Key.new("Home", [tag])
      assert Key.get_key_str(key1) == Key.get_key_str(key2)
    end

    test "raw key string", _ctx do
      tag = Constant.new("Site", "main", version: 1.0)
      key = Key.new("Home", [tag])
      raw = Key.get_key_str(key, false)
      assert raw =~ "Home"
      assert raw =~ "Site"
    end

    test "key with mixed handlers same without increment", _ctx do
      handler_a = Memory.new()
      handler_b = Memory.new()
      tag_a1 = Standard.new("User", "1", handler: handler_a)
      tag_b1 = Standard.new("City", "nyc", handler: handler_b)
      key1 = Key.new("Mixed", [tag_a1, tag_b1])
      k1 = Key.get_key_str(key1)

      tag_a2 = Standard.new("User", "1", handler: handler_a)
      tag_b2 = Standard.new("City", "nyc", handler: handler_b)
      key2 = Key.new("Mixed", [tag_a2, tag_b2])
      k2 = Key.get_key_str(key2)

      assert k1 == k2
    end

    test "key with mixed handlers different after increment", _ctx do
      handler_a = Memory.new()
      handler_b = Memory.new()
      tag_a = Standard.new("User", "1", handler: handler_a)
      tag_b = Standard.new("City", "nyc", handler: handler_b)
      key1 = Key.new("Mixed", [tag_a, tag_b])
      k1 = Key.get_key_str(key1)

      Tag.increment(tag_a)
      tag_a2 = Standard.new("User", "1", handler: handler_a)
      tag_b2 = Standard.new("City", "nyc", handler: handler_b)
      key2 = Key.new("Mixed", [tag_a2, tag_b2])
      k2 = Key.get_key_str(key2)

      assert k1 != k2
    end

    test "constant tag skipped in bulk fetch", _ctx do
      tag = Constant.new("Site", "main", version: 3.0)
      assert Tag.delegate_cache_query(tag, "MemoryHandler") == false
    end

    test "standard tag delegates to matching handler", _ctx do
      tag = Standard.new("User", "1")
      assert Tag.delegate_cache_query(tag, "MemoryHandler") == true
      assert Tag.delegate_cache_query(tag, "OtherHandler") == false
    end
  end

  # ---------------------------------------------------------------------------
  # KeyRing tests
  # ---------------------------------------------------------------------------

  describe "KeyRing" do
    test "define and get key", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"memory" => handler},
          default_cache_handler: "memory",
          default_prefix: "Test"
        )

      ring = KeyRing.define_key(ring, "Users", ["universe", "planet", "city"])
      key_obj = KeyRing.get_key_obj(ring, "Users", ["MilkyWay", "Earth", "Chicago"])
      k = Key.get_key_str(key_obj)
      assert is_binary(k)
      assert String.length(k) == 32
    end

    test "dynamic accessor", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"memory" => handler},
          default_cache_handler: "memory",
          default_prefix: "Test"
        )

      ring = KeyRing.define_key(ring, "Users", ["universe", "planet", "city"])
      key_obj = KeyRing.key_obj(ring, "users", ["MilkyWay", "Earth", "Chicago"])
      assert is_binary(Key.get_key_str(key_obj))
    end

    test "keyring matches manual key", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"memory" => handler},
          default_cache_handler: "memory",
          default_prefix: "Test"
        )

      ring = KeyRing.define_key(ring, "Profile", ["user"])
      ring_key = KeyRing.get_key_obj(ring, "Profile", ["42"])
      ring_str = Key.get_key_str(ring_key)

      manual_tag = Standard.new("user", "42", handler: handler, prefix: "Test")
      manual_key = Key.new("Profile", [manual_tag])
      manual_str = Key.get_key_str(manual_key)

      assert ring_str == manual_str
    end

    test "keyring with constant tag", %{handler: handler} do
      ring =
        KeyRing.new(
          global_tag_options: %{"universe" => %{"type" => "constant", "version" => 5.0}},
          cache_handlers: %{"memory" => handler},
          default_cache_handler: "memory",
          default_prefix: "Test"
        )

      ring = KeyRing.define_key(ring, "World", ["universe", "planet"])
      key1 = KeyRing.get_key_obj(ring, "World", ["MilkyWay", "Earth"])
      key2 = KeyRing.get_key_obj(ring, "World", ["MilkyWay", "Earth"])
      assert Key.get_key_str(key1) == Key.get_key_str(key2)
    end

    test "keyring with handler override", %{handler: handler} do
      alt_handler = Memory.new()

      ring =
        KeyRing.new(
          cache_handlers: %{"memory" => handler, "alt" => alt_handler},
          default_cache_handler: "memory",
          default_prefix: "Test"
        )

      ring =
        KeyRing.define_key(ring, "Mixed", [
          "user",
          %{"tag" => "city", "cache_handler" => "alt"}
        ])

      key_obj = KeyRing.get_key_obj(ring, "Mixed", ["42", "chicago"])
      assert is_binary(Key.get_key_str(key_obj))
    end

    test "keyring wrong arg count raises", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"memory" => handler},
          default_cache_handler: "memory",
          default_prefix: "Test"
        )

      ring = KeyRing.define_key(ring, "X", ["a", "b"])

      assert_raise ArgumentError, ~r/expects 2/, fn ->
        KeyRing.get_key_obj(ring, "X", ["only_one"])
      end
    end
  end
end
