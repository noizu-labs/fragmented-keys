"""Integration tests against a real Redis instance on localhost:6379."""

import time

import pytest
from redis import Redis

from fragmented_keys import (
    Configuration,
    ConstantTag,
    FragmentedKeyRing,
    MemoryHandler,
    RedisHandler,
    StandardKey,
    StandardTag,
)


@pytest.fixture()
def redis_client():
    client = Redis(host="127.0.0.1", port=6379)
    client.ping()
    return client


@pytest.fixture()
def redis_handler(redis_client):
    handler = RedisHandler(redis_client)
    Configuration.set_default_cache_handler(handler)
    Configuration.set_global_prefix("FragKeyTest")
    return handler


@pytest.fixture(autouse=True)
def _cleanup(redis_client):
    """Clean up test keys after each test."""
    yield
    # Delete all keys with our test prefix
    for key in redis_client.scan_iter("*FragKeyTest*"):
        redis_client.delete(key)
    Configuration.reset()


# ---------------------------------------------------------------------------
# Standard Tag tests
# ---------------------------------------------------------------------------

class TestStandardTagRedis:
    def test_same_version_without_increment(self, redis_handler):
        tag = StandardTag("User", "42")
        v1 = tag.get_tag_version()
        v2 = tag.get_tag_version()
        assert v1 == v2

    def test_version_persists_across_instances(self, redis_handler):
        tag1 = StandardTag("User", "100")
        v1 = tag1.get_tag_version()

        # New instance should read same version from Redis
        tag2 = StandardTag("User", "100")
        v2 = tag2.get_tag_version()
        assert v1 == v2

    def test_different_version_after_increment(self, redis_handler):
        tag = StandardTag("User", "42")
        v1 = tag.get_tag_version()
        tag.increment()
        v2 = tag.get_tag_version()
        assert v1 != v2
        assert v2 == pytest.approx(v1 + 0.1, abs=1e-9)

    def test_increment_persists_to_redis(self, redis_handler):
        tag = StandardTag("User", "200")
        tag.get_tag_version()
        tag.increment()
        v_after = tag.get_tag_version()

        # New instance should see incremented version
        tag2 = StandardTag("User", "200")
        v2 = tag2.get_tag_version()
        assert v2 == v_after

    def test_different_instances_different_versions(self, redis_handler):
        tag_a = StandardTag("User", "1")
        tag_b = StandardTag("User", "2")
        tag_a.get_tag_version()
        tag_b.get_tag_version()
        assert tag_a.get_tag_name() != tag_b.get_tag_name()

    def test_different_tag_names(self, redis_handler):
        tag_a = StandardTag("User", "1")
        tag_b = StandardTag("City", "1")
        assert tag_a.get_tag_name() != tag_b.get_tag_name()

    def test_increment_one_does_not_affect_other(self, redis_handler):
        tag_a = StandardTag("User", "1")
        tag_b = StandardTag("User", "2")
        v_b_before = tag_b.get_tag_version()
        tag_a.increment()

        tag_b2 = StandardTag("User", "2")
        v_b_after = tag_b2.get_tag_version()
        assert v_b_before == v_b_after

    def test_increment_different_tag_no_effect(self, redis_handler):
        tag_user = StandardTag("User", "1")
        tag_city = StandardTag("City", "1")
        v_city_before = tag_city.get_tag_version()
        tag_user.increment()

        tag_city2 = StandardTag("City", "1")
        v_city_after = tag_city2.get_tag_version()
        assert v_city_before == v_city_after

    def test_reset_tag_version(self, redis_handler):
        tag = StandardTag("User", "300")
        v1 = tag.get_tag_version()
        time.sleep(0.01)  # Ensure different timestamp
        tag.reset_tag_version()
        v2 = tag.get_tag_version()
        assert v1 != v2


# ---------------------------------------------------------------------------
# Constant Tag tests
# ---------------------------------------------------------------------------

class TestConstantTagRedis:
    def test_constant_version_fixed(self, redis_handler):
        tag = ConstantTag("Site", "main", version=5.0)
        assert tag.get_tag_version() == 5.0

    def test_increment_is_noop(self, redis_handler):
        tag = ConstantTag("Site", "main", version=5.0)
        tag.increment()
        assert tag.get_tag_version() == 5.0

    def test_delegate_cache_query_returns_false(self, redis_handler):
        tag = ConstantTag("Site", "main", version=3.0)
        assert tag.delegate_cache_query("RedisHandler") is False


# ---------------------------------------------------------------------------
# Key tests
# ---------------------------------------------------------------------------

class TestStandardKeyRedis:
    def test_key_consistent_without_increment(self, redis_handler):
        tag1 = StandardTag("User", "42")
        key1 = StandardKey("Profile", [tag1])
        k1 = key1.get_key_str()

        tag2 = StandardTag("User", "42")
        key2 = StandardKey("Profile", [tag2])
        k2 = key2.get_key_str()
        assert k1 == k2

    def test_key_changes_after_increment(self, redis_handler):
        tag = StandardTag("User", "42")
        key1 = StandardKey("Profile", [tag])
        k1 = key1.get_key_str()

        tag.increment()

        tag2 = StandardTag("User", "42")
        key2 = StandardKey("Profile", [tag2])
        k2 = key2.get_key_str()
        assert k1 != k2

    def test_key_with_multiple_tags(self, redis_handler):
        tag_user = StandardTag("User", "1")
        tag_city = StandardTag("City", "chicago")
        key = StandardKey("Dashboard", [tag_user, tag_city])
        k = key.get_key_str()
        assert isinstance(k, str)
        assert len(k) == 32

    def test_key_with_constant_tags_stable(self, redis_handler):
        tag = ConstantTag("Site", "main", version=1.0)
        key1 = StandardKey("Home", [tag])
        key2 = StandardKey("Home", [tag])
        assert key1.get_key_str() == key2.get_key_str()

    def test_key_with_mixed_handlers_same_without_increment(self, redis_handler):
        """Redis + Memory handlers return same key if no increment."""
        mem_handler = MemoryHandler()
        tag_a1 = StandardTag("User", "1", handler=redis_handler)
        tag_b1 = StandardTag("City", "nyc", handler=mem_handler)
        key1 = StandardKey("Mixed", [tag_a1, tag_b1])
        k1 = key1.get_key_str()

        tag_a2 = StandardTag("User", "1", handler=redis_handler)
        tag_b2 = StandardTag("City", "nyc", handler=mem_handler)
        key2 = StandardKey("Mixed", [tag_a2, tag_b2])
        k2 = key2.get_key_str()
        assert k1 == k2

    def test_key_with_mixed_handlers_different_after_increment(self, redis_handler):
        """Redis + Memory handlers return different key after increment."""
        mem_handler = MemoryHandler()
        tag_a = StandardTag("User", "1", handler=redis_handler)
        tag_b = StandardTag("City", "nyc", handler=mem_handler)
        key1 = StandardKey("Mixed", [tag_a, tag_b])
        k1 = key1.get_key_str()

        tag_a.increment()
        tag_a2 = StandardTag("User", "1", handler=redis_handler)
        tag_b2 = StandardTag("City", "nyc", handler=mem_handler)
        key2 = StandardKey("Mixed", [tag_a2, tag_b2])
        k2 = key2.get_key_str()
        assert k1 != k2

    def test_bulk_fetch_via_mget(self, redis_handler):
        """Verify multi-get optimization works end-to-end with Redis."""
        tags = [StandardTag("BulkTag", str(i)) for i in range(5)]
        key = StandardKey("BulkTest", tags)
        k1 = key.get_key_str()

        # All versions should now be in Redis; rebuild with fresh tags
        tags2 = [StandardTag("BulkTag", str(i)) for i in range(5)]
        key2 = StandardKey("BulkTest", tags2)
        k2 = key2.get_key_str()
        assert k1 == k2

        # Increment one tag and verify key changes
        tags[2].increment()
        tags3 = [StandardTag("BulkTag", str(i)) for i in range(5)]
        key3 = StandardKey("BulkTest", tags3)
        k3 = key3.get_key_str()
        assert k1 != k3


# ---------------------------------------------------------------------------
# KeyRing tests
# ---------------------------------------------------------------------------

class TestKeyRingRedis:
    def test_define_and_get_key(self, redis_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"redis": redis_handler},
            default_cache_handler="redis",
            default_prefix="FragKeyTest",
        )
        ring.define_key("Users", ["universe", "planet", "city"])
        key_obj = ring.get_key_obj("Users", ["MilkyWay", "Earth", "Chicago"])
        k = key_obj.get_key_str()
        assert isinstance(k, str)
        assert len(k) == 32

    def test_keyring_matches_manual_key(self, redis_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"redis": redis_handler},
            default_cache_handler="redis",
            default_prefix="FragKeyTest",
        )
        ring.define_key("Profile", ["user"])
        ring_key = ring.get_key_obj("Profile", ["42"])
        ring_str = ring_key.get_key_str()

        manual_tag = StandardTag("user", "42", handler=redis_handler, prefix="FragKeyTest")
        manual_key = StandardKey("Profile", [manual_tag])
        manual_str = manual_key.get_key_str()

        assert ring_str == manual_str

    def test_keyring_with_constant_tag(self, redis_handler):
        ring = FragmentedKeyRing(
            global_tag_options={"universe": {"type": "constant", "version": 5.0}},
            cache_handlers={"redis": redis_handler},
            default_cache_handler="redis",
            default_prefix="FragKeyTest",
        )
        ring.define_key("World", ["universe", "planet"])
        key1 = ring.get_key_obj("World", ["MilkyWay", "Earth"])
        key2 = ring.get_key_obj("World", ["MilkyWay", "Earth"])
        assert key1.get_key_str() == key2.get_key_str()

    def test_keyring_with_mixed_handlers(self, redis_handler):
        mem_handler = MemoryHandler()
        ring = FragmentedKeyRing(
            cache_handlers={"redis": redis_handler, "memory": mem_handler},
            default_cache_handler="redis",
            default_prefix="FragKeyTest",
        )
        ring.define_key(
            "Mixed",
            ["user", {"tag": "city", "cache_handler": "memory"}],
        )
        key_obj = ring.get_key_obj("Mixed", ["42", "chicago"])
        assert isinstance(key_obj.get_key_str(), str)

    def test_dynamic_accessor(self, redis_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"redis": redis_handler},
            default_cache_handler="redis",
            default_prefix="FragKeyTest",
        )
        ring.define_key("Users", ["universe", "planet"])
        key_obj = ring.get_users_key_obj("MilkyWay", "Earth")
        assert isinstance(key_obj.get_key_str(), str)
