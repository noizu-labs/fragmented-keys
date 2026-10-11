defmodule FragmentedKeysRedisTest do
  @moduledoc """
  End-to-end integration tests against a live Redis instance, mirroring the
  Python `TestStandardTagRedis` / `TestConstantTagRedis` / `TestStandardKeyRedis`
  / `TestKeyRingRedis` suites.

  Tagged `:redis` and excluded by default (see `test_helper.exs`). Run with:

      mix test --include redis

  Override the target with `FK_REDIS_HOST` / `FK_REDIS_PORT`. For the bundled
  docker-compose service: `FK_REDIS_PORT=6380 mix test --include redis`.
  """

  use ExUnit.Case, async: false

  @moduletag :redis

  alias FragmentedKeys.{Configuration, Key, KeyRing, Tag}
  alias FragmentedKeys.CacheHandler.{Memory, Redis}
  alias FragmentedKeys.Tag.{Constant, Standard}

  @prefix "FragKeyTest"

  setup_all do
    host = System.get_env("FK_REDIS_HOST", "127.0.0.1")
    port = System.get_env("FK_REDIS_PORT", "6379") |> String.to_integer()

    conn = start_supervised!({Redix, host: host, port: port})

    case Redix.command(conn, ["PING"]) do
      {:ok, "PONG"} ->
        :ok

      other ->
        raise "Redis not reachable at #{host}:#{port} (#{inspect(other)}). " <>
                "Start it with `docker compose up redis` (then FK_REDIS_PORT=6380) " <>
                "or a local `redis-server`."
    end

    {:ok, conn: conn, handler: Redis.new(conn)}
  end

  setup %{conn: conn, handler: handler} do
    Configuration.reset()
    Configuration.set_default_cache_handler(handler)
    Configuration.set_global_prefix(@prefix)

    on_exit(fn ->
      case Redix.command(conn, ["KEYS", "*#{@prefix}*"]) do
        {:ok, [_ | _] = keys} -> Redix.command(conn, ["DEL" | keys])
        _ -> :ok
      end

      Configuration.reset()
    end)

    :ok
  end

  # ---------------------------------------------------------------------------
  # Standard Tag tests
  # ---------------------------------------------------------------------------

  describe "StandardTag (Redis)" do
    test "same version without increment" do
      tag = Standard.new("User", "42")
      assert Tag.get_tag_version(tag) == Tag.get_tag_version(tag)
    end

    test "version persists across instances" do
      v1 = Tag.get_tag_version(Standard.new("User", "100"))
      v2 = Tag.get_tag_version(Standard.new("User", "100"))
      assert v1 == v2
    end

    test "different version after increment" do
      tag = Standard.new("User", "42")
      v1 = Tag.get_tag_version(tag)
      tag = Tag.increment(tag)
      v2 = Tag.get_tag_version(tag)
      assert v1 != v2
      assert_in_delta v2, v1 + 0.1, 1.0e-9
    end

    test "increment persists to redis" do
      tag = Standard.new("User", "200")
      Tag.get_tag_version(tag)
      tag = Tag.increment(tag)
      v_after = Tag.get_tag_version(tag)

      v2 = Tag.get_tag_version(Standard.new("User", "200"))
      assert v2 == v_after
    end

    test "different instances have different tag names" do
      assert Tag.get_tag_name(Standard.new("User", "1")) !=
               Tag.get_tag_name(Standard.new("User", "2"))
    end

    test "different tag names" do
      assert Tag.get_tag_name(Standard.new("User", "1")) !=
               Tag.get_tag_name(Standard.new("City", "1"))
    end

    test "increment one does not affect other" do
      v_b_before = Tag.get_tag_version(Standard.new("User", "2"))
      Tag.increment(Standard.new("User", "1"))
      v_b_after = Tag.get_tag_version(Standard.new("User", "2"))
      assert v_b_before == v_b_after
    end

    test "increment different tag has no effect" do
      v_city_before = Tag.get_tag_version(Standard.new("City", "1"))
      Tag.increment(Standard.new("User", "1"))
      v_city_after = Tag.get_tag_version(Standard.new("City", "1"))
      assert v_city_before == v_city_after
    end

    test "reset changes version" do
      tag = Standard.new("User", "300")
      v1 = Tag.get_tag_version(tag)
      Process.sleep(10)
      tag = Tag.reset_tag_version(tag)
      v2 = Tag.get_tag_version(tag)
      assert v1 != v2
    end
  end

  # ---------------------------------------------------------------------------
  # Constant Tag tests
  # ---------------------------------------------------------------------------

  describe "ConstantTag (Redis)" do
    test "constant version is fixed" do
      assert Tag.get_tag_version(Constant.new("Site", "main", version: 5.0)) == 5.0
    end

    test "increment is noop" do
      tag = Constant.new("Site", "main", version: 5.0)
      assert Tag.get_tag_version(Tag.increment(tag)) == 5.0
    end

    test "delegate_cache_query returns false" do
      assert Tag.delegate_cache_query(Constant.new("Site", "main", version: 3.0), "RedisHandler") ==
               false
    end
  end

  # ---------------------------------------------------------------------------
  # Key tests
  # ---------------------------------------------------------------------------

  describe "Key (Redis)" do
    test "key consistent without increment" do
      k1 = Key.get_key_str(Key.new("Profile", [Standard.new("User", "42")]))
      k2 = Key.get_key_str(Key.new("Profile", [Standard.new("User", "42")]))
      assert k1 == k2
    end

    test "key changes after increment" do
      tag = Standard.new("User", "42")
      k1 = Key.get_key_str(Key.new("Profile", [tag]))
      Tag.increment(tag)
      k2 = Key.get_key_str(Key.new("Profile", [Standard.new("User", "42")]))
      assert k1 != k2
    end

    test "key with multiple tags" do
      key = Key.new("Dashboard", [Standard.new("User", "1"), Standard.new("City", "chicago")])
      k = Key.get_key_str(key)
      assert is_binary(k)
      assert String.length(k) == 32
    end

    test "key with constant tags is stable" do
      tag = Constant.new("Site", "main", version: 1.0)
      assert Key.get_key_str(Key.new("Home", [tag])) == Key.get_key_str(Key.new("Home", [tag]))
    end

    test "key with mixed handlers same without increment", %{handler: redis_handler} do
      mem = Memory.new()
      key1 =
        Key.new("Mixed", [
          Standard.new("User", "1", handler: redis_handler),
          Standard.new("City", "nyc", handler: mem)
        ])

      k1 = Key.get_key_str(key1)

      key2 =
        Key.new("Mixed", [
          Standard.new("User", "1", handler: redis_handler),
          Standard.new("City", "nyc", handler: mem)
        ])

      assert k1 == Key.get_key_str(key2)
    end

    test "key with mixed handlers different after increment", %{handler: redis_handler} do
      mem = Memory.new()
      tag_a = Standard.new("User", "1", handler: redis_handler)
      tag_b = Standard.new("City", "nyc", handler: mem)
      k1 = Key.get_key_str(Key.new("Mixed", [tag_a, tag_b]))

      Tag.increment(tag_a)

      key2 =
        Key.new("Mixed", [
          Standard.new("User", "1", handler: redis_handler),
          Standard.new("City", "nyc", handler: mem)
        ])

      assert k1 != Key.get_key_str(key2)
    end

    test "bulk fetch via MGET" do
      tags = for i <- 0..4, do: Standard.new("BulkTag", Integer.to_string(i))
      k1 = Key.get_key_str(Key.new("BulkTest", tags))

      tags2 = for i <- 0..4, do: Standard.new("BulkTag", Integer.to_string(i))
      k2 = Key.get_key_str(Key.new("BulkTest", tags2))
      assert k1 == k2

      Tag.increment(Enum.at(tags, 2))
      tags3 = for i <- 0..4, do: Standard.new("BulkTag", Integer.to_string(i))
      k3 = Key.get_key_str(Key.new("BulkTest", tags3))
      assert k1 != k3
    end
  end

  # ---------------------------------------------------------------------------
  # KeyRing tests
  # ---------------------------------------------------------------------------

  describe "KeyRing (Redis)" do
    test "define and get key", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"redis" => handler},
          default_cache_handler: "redis",
          default_prefix: @prefix
        )

      ring = KeyRing.define_key(ring, "Users", ["universe", "planet", "city"])
      k = Key.get_key_str(KeyRing.get_key_obj(ring, "Users", ["MilkyWay", "Earth", "Chicago"]))
      assert is_binary(k)
      assert String.length(k) == 32
    end

    test "keyring matches manual key", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"redis" => handler},
          default_cache_handler: "redis",
          default_prefix: @prefix
        )

      ring = KeyRing.define_key(ring, "Profile", ["user"])
      ring_str = Key.get_key_str(KeyRing.get_key_obj(ring, "Profile", ["42"]))

      manual_tag = Standard.new("user", "42", handler: handler, prefix: @prefix)
      manual_str = Key.get_key_str(Key.new("Profile", [manual_tag]))

      assert ring_str == manual_str
    end

    test "keyring with constant tag", %{handler: handler} do
      ring =
        KeyRing.new(
          global_tag_options: %{"universe" => %{"type" => "constant", "version" => 5.0}},
          cache_handlers: %{"redis" => handler},
          default_cache_handler: "redis",
          default_prefix: @prefix
        )

      ring = KeyRing.define_key(ring, "World", ["universe", "planet"])
      k1 = Key.get_key_str(KeyRing.get_key_obj(ring, "World", ["MilkyWay", "Earth"]))
      k2 = Key.get_key_str(KeyRing.get_key_obj(ring, "World", ["MilkyWay", "Earth"]))
      assert k1 == k2
    end

    test "keyring with mixed handlers", %{handler: handler} do
      mem = Memory.new()

      ring =
        KeyRing.new(
          cache_handlers: %{"redis" => handler, "memory" => mem},
          default_cache_handler: "redis",
          default_prefix: @prefix
        )

      ring =
        KeyRing.define_key(ring, "Mixed", [
          "user",
          %{"tag" => "city", "cache_handler" => "memory"}
        ])

      assert is_binary(Key.get_key_str(KeyRing.get_key_obj(ring, "Mixed", ["42", "chicago"])))
    end

    test "dynamic accessor", %{handler: handler} do
      ring =
        KeyRing.new(
          cache_handlers: %{"redis" => handler},
          default_cache_handler: "redis",
          default_prefix: @prefix
        )

      ring = KeyRing.define_key(ring, "Users", ["universe", "planet"])
      assert is_binary(Key.get_key_str(KeyRing.key_obj(ring, "users", ["MilkyWay", "Earth"])))
    end
  end
end
