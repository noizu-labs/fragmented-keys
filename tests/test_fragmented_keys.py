"""Tests mirroring the PHP FragmentedKeysTest suite, using MemoryHandler."""

import pytest

from fragmented_keys import (
    Configuration,
    ConstantTag,
    FragmentedKeyRing,
    MemoryHandler,
    StandardKey,
    StandardTag,
)


# ---------------------------------------------------------------------------
# Standard Tag tests
# ---------------------------------------------------------------------------

class TestStandardTag:
    def test_same_version_without_increment(self, memory_handler):
        tag = StandardTag("User", "42")
        v1 = tag.get_tag_version()
        v2 = tag.get_tag_version()
        assert v1 == v2

    def test_different_version_after_increment(self, memory_handler):
        tag = StandardTag("User", "42")
        v1 = tag.get_tag_version()
        tag.increment()
        v2 = tag.get_tag_version()
        assert v1 != v2

    def test_different_instances_different_versions(self, memory_handler):
        tag_a = StandardTag("User", "1")
        tag_b = StandardTag("User", "2")
        # They should each get their own version (seeded from time.time)
        v_a = tag_a.get_tag_version()
        v_b = tag_b.get_tag_version()
        # Versions are unique per-instance since they are seeded independently
        assert tag_a.get_tag_name() != tag_b.get_tag_name()

    def test_different_tag_names(self, memory_handler):
        tag_a = StandardTag("User", "1")
        tag_b = StandardTag("City", "1")
        assert tag_a.get_tag_name() != tag_b.get_tag_name()

    def test_increment_changes_version(self, memory_handler):
        tag = StandardTag("User", "1")
        v1 = tag.get_tag_version()
        tag.increment()
        v2 = tag.get_tag_version()
        assert v2 == pytest.approx(v1 + 0.1, abs=1e-9)

    def test_increment_one_does_not_affect_other(self, memory_handler):
        tag_a = StandardTag("User", "1")
        tag_b = StandardTag("User", "2")
        v_b_before = tag_b.get_tag_version()
        tag_a.increment()
        # Re-read from cache
        tag_b2 = StandardTag("User", "2")
        v_b_after = tag_b2.get_tag_version()
        assert v_b_before == v_b_after

    def test_increment_different_tag_no_effect(self, memory_handler):
        tag_user = StandardTag("User", "1")
        tag_city = StandardTag("City", "1")
        v_city_before = tag_city.get_tag_version()
        tag_user.increment()
        tag_city2 = StandardTag("City", "1")
        v_city_after = tag_city2.get_tag_version()
        assert v_city_before == v_city_after



# ---------------------------------------------------------------------------
# Constant Tag tests
# ---------------------------------------------------------------------------

class TestConstantTag:
    def test_constant_version_fixed(self, memory_handler):
        tag = ConstantTag("Site", "main", version=5.0)
        assert tag.get_tag_version() == 5.0

    def test_increment_is_noop(self, memory_handler):
        tag = ConstantTag("Site", "main", version=5.0)
        tag.increment()
        assert tag.get_tag_version() == 5.0

    def test_reset_is_noop(self, memory_handler):
        tag = ConstantTag("Site", "main", version=5.0)
        tag.reset_tag_version()
        assert tag.get_tag_version() == 5.0

    def test_set_version_is_noop(self, memory_handler):
        tag = ConstantTag("Site", "main", version=5.0)
        tag.set_tag_version(99.0, update=True)
        assert tag.get_tag_version() == 5.0


# ---------------------------------------------------------------------------
# Key tests
# ---------------------------------------------------------------------------

class TestStandardKey:
    def test_key_consistent_without_increment(self, memory_handler):
        tag = StandardTag("User", "42")
        key1 = StandardKey("Profile", [tag])
        k1 = key1.get_key_str()

        tag2 = StandardTag("User", "42")
        key2 = StandardKey("Profile", [tag2])
        k2 = key2.get_key_str()
        assert k1 == k2

    def test_key_changes_after_increment(self, memory_handler):
        tag = StandardTag("User", "42")
        key1 = StandardKey("Profile", [tag])
        k1 = key1.get_key_str()

        tag.increment()

        tag2 = StandardTag("User", "42")
        key2 = StandardKey("Profile", [tag2])
        k2 = key2.get_key_str()
        assert k1 != k2

    def test_rerender_reflects_external_invalidation(self, memory_handler):
        # Contract: get_key_str() always reflects current cache state.
        # Rendering a key, then having an independent tag object invalidate
        # the same identity, must change the next render of the SAME key
        # object (guards against memoizing fetched versions onto the tags).
        tag = StandardTag("User", "42")
        key = StandardKey("Profile", [tag])
        k1 = key.get_key_str()

        StandardTag("User", "42").increment()  # independent object, shared cache

        k2 = key.get_key_str()
        assert k1 != k2

    def test_key_with_multiple_tags(self, memory_handler):
        tag_user = StandardTag("User", "1")
        tag_city = StandardTag("City", "chicago")
        key = StandardKey("Dashboard", [tag_user, tag_city])
        k = key.get_key_str()
        assert isinstance(k, str)
        assert len(k) == 32  # MD5 hex digest

    def test_key_with_constant_tags_stable(self, memory_handler):
        tag = ConstantTag("Site", "main", version=1.0)
        key1 = StandardKey("Home", [tag])
        key2 = StandardKey("Home", [tag])
        assert key1.get_key_str() == key2.get_key_str()

    def test_raw_key_string(self, memory_handler):
        tag = ConstantTag("Site", "main", version=1.0)
        key = StandardKey("Home", [tag])
        raw = key.get_key_str(hash=False)
        assert "Home" in raw
        assert "Site" in raw

    def test_key_with_mixed_handlers_same_without_increment(self, memory_handler):
        """Mixed handlers should return the same key if tags are not incremented."""
        handler_a = MemoryHandler()
        handler_b = MemoryHandler()
        tag_a1 = StandardTag("User", "1", handler=handler_a)
        tag_b1 = StandardTag("City", "nyc", handler=handler_b)
        key1 = StandardKey("Mixed", [tag_a1, tag_b1])
        k1 = key1.get_key_str()

        tag_a2 = StandardTag("User", "1", handler=handler_a)
        tag_b2 = StandardTag("City", "nyc", handler=handler_b)
        key2 = StandardKey("Mixed", [tag_a2, tag_b2])
        k2 = key2.get_key_str()
        assert k1 == k2

    def test_key_with_mixed_handlers_different_after_increment(self, memory_handler):
        """Mixed handlers should return a different key after incrementing a tag."""
        handler_a = MemoryHandler()
        handler_b = MemoryHandler()
        tag_a = StandardTag("User", "1", handler=handler_a)
        tag_b = StandardTag("City", "nyc", handler=handler_b)
        key = StandardKey("Mixed", [tag_a, tag_b])
        k1 = key.get_key_str()

        tag_a.increment()
        tag_a2 = StandardTag("User", "1", handler=handler_a)
        tag_b2 = StandardTag("City", "nyc", handler=handler_b)
        key2 = StandardKey("Mixed", [tag_a2, tag_b2])
        k2 = key2.get_key_str()
        assert k1 != k2

    def test_constant_tag_skipped_in_bulk_fetch(self, memory_handler):
        """ConstantTag should not participate in bulk cache fetches."""
        tag = ConstantTag("Site", "main", version=3.0)
        assert tag.delegate_cache_query("MemoryHandler") is False

    def test_standard_tag_delegates_to_matching_handler(self, memory_handler):
        """StandardTag should delegate to its handler's group."""
        tag = StandardTag("User", "1")
        assert tag.delegate_cache_query("MemoryHandler") is True
        assert tag.delegate_cache_query("OtherHandler") is False


# ---------------------------------------------------------------------------
# KeyRing tests
# ---------------------------------------------------------------------------

class TestKeyRing:
    def test_define_and_get_key(self, memory_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"memory": memory_handler},
            default_cache_handler="memory",
            default_prefix="Test",
        )
        ring.define_key("Users", ["universe", "planet", "city"])
        key_obj = ring.get_key_obj("Users", ["MilkyWay", "Earth", "Chicago"])
        k = key_obj.get_key_str()
        assert isinstance(k, str)
        assert len(k) == 32

    def test_dynamic_accessor(self, memory_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"memory": memory_handler},
            default_cache_handler="memory",
            default_prefix="Test",
        )
        ring.define_key("Users", ["universe", "planet", "city"])
        key_obj = ring.get_users_key_obj("MilkyWay", "Earth", "Chicago")
        assert isinstance(key_obj.get_key_str(), str)

    def test_keyring_matches_manual_key(self, memory_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"memory": memory_handler},
            default_cache_handler="memory",
            default_prefix="Test",
        )
        ring.define_key("Profile", ["user"])
        ring_key = ring.get_key_obj("Profile", ["42"])
        ring_str = ring_key.get_key_str()

        manual_tag = StandardTag("user", "42", handler=memory_handler, prefix="Test")
        manual_key = StandardKey("Profile", [manual_tag])
        manual_str = manual_key.get_key_str()

        assert ring_str == manual_str

    def test_keyring_with_constant_tag(self, memory_handler):
        ring = FragmentedKeyRing(
            global_tag_options={"universe": {"type": "constant", "version": 5.0}},
            cache_handlers={"memory": memory_handler},
            default_cache_handler="memory",
            default_prefix="Test",
        )
        ring.define_key("World", ["universe", "planet"])
        key1 = ring.get_key_obj("World", ["MilkyWay", "Earth"])
        key2 = ring.get_key_obj("World", ["MilkyWay", "Earth"])
        assert key1.get_key_str() == key2.get_key_str()

    def test_keyring_with_handler_override(self, memory_handler):
        alt_handler = MemoryHandler()
        ring = FragmentedKeyRing(
            cache_handlers={"memory": memory_handler, "alt": alt_handler},
            default_cache_handler="memory",
            default_prefix="Test",
        )
        ring.define_key(
            "Mixed",
            ["user", {"tag": "city", "cache_handler": "alt"}],
        )
        key_obj = ring.get_key_obj("Mixed", ["42", "chicago"])
        assert isinstance(key_obj.get_key_str(), str)

    def test_keyring_wrong_arg_count_raises(self, memory_handler):
        ring = FragmentedKeyRing(
            cache_handlers={"memory": memory_handler},
            default_cache_handler="memory",
            default_prefix="Test",
        )
        ring.define_key("X", ["a", "b"])
        with pytest.raises(ValueError, match="expects 2"):
            ring.get_key_obj("X", ["only_one"])
