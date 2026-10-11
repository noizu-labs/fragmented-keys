from __future__ import annotations

from typing import Protocol, runtime_checkable


@runtime_checkable
class CacheHandler(Protocol):
    """Interface for cache backend implementations."""

    def group_name(self) -> str:
        """Return a unique identifier for this handler type."""
        ...

    def get(self, key: str) -> str | None:
        """Fetch a single value by key. Returns None on miss."""
        ...

    def set(self, key: str, value: str, ttl: int | None = None) -> None:
        """Store a value, optionally with a TTL in seconds."""
        ...

    def get_multi(self, keys: list[str]) -> dict[str, str]:
        """Fetch multiple values at once. Returns dict of found key-value pairs."""
        ...


@runtime_checkable
class Tag(Protocol):
    """Interface for versioned tag-instance pairs."""

    def get_tag_name(self) -> str:
        """Return the tag identifier including instance and prefix."""
        ...

    def get_tag_version(self) -> float:
        """Return the current version for this tag-instance."""
        ...

    def get_full_tag(self) -> str:
        """Return the tag name with its current version."""
        ...

    def increment(self) -> None:
        """Increment the tag version, invalidating dependent keys."""
        ...

    def reset_tag_version(self) -> None:
        """Reset the tag version to a new value."""
        ...

    def set_tag_version(self, version: float, update: bool = False) -> None:
        """Manually set the tag version."""
        ...

    def set_cache_handler(self, handler: CacheHandler) -> None:
        """Override the cache handler for this tag."""
        ...

    def get_cache_handler(self) -> CacheHandler:
        """Return the cache handler used by this tag."""
        ...

    def delegate_cache_query(self, group: str) -> bool:
        """Check if this tag's version can be bulk-fetched with the given group."""
        ...


@runtime_checkable
class Key(Protocol):
    """Interface for composite cache keys."""

    def get_key_str(self, hash: bool = True) -> str:
        """Generate the composite cache key string."""
        ...

    def add_tag(self, tag: Tag) -> None:
        """Add a tag to this key's composition."""
        ...


@runtime_checkable
class KeyRing(Protocol):
    """Interface for key template factories."""

    def define_key(self, key: str, params: list, globals: dict | None = None) -> None:
        """Define a reusable key template."""
        ...

    def tag(self, tag: str, instance: str, options: dict | None = None) -> Tag:
        """Factory method for creating tag instances."""
        ...
