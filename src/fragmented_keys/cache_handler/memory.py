from __future__ import annotations


class MemoryHandler:
    """In-memory cache handler for testing and temporary caching."""

    _cache: dict[str, str]

    def __init__(self) -> None:
        self._cache = {}

    def group_name(self) -> str:
        return "MemoryHandler"

    def get(self, key: str) -> str | None:
        return self._cache.get(key)

    def set(self, key: str, value: str, ttl: int | None = None) -> None:
        self._cache[key] = value

    def get_multi(self, keys: list[str]) -> dict[str, str]:
        return {k: self._cache[k] for k in keys if k in self._cache}

    def clear(self) -> None:
        """Clear all cached values."""
        self._cache.clear()
