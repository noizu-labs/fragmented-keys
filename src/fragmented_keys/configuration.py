from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from fragmented_keys.protocols import CacheHandler


class Configuration:
    """Global configuration for the fragmented keys library."""

    _default_cache_handler: CacheHandler | None = None
    _global_prefix: str = "DefaultPrefix"

    @classmethod
    def set_default_cache_handler(cls, handler: CacheHandler) -> None:
        cls._default_cache_handler = handler

    @classmethod
    def get_default_cache_handler(cls) -> CacheHandler:
        if cls._default_cache_handler is None:
            raise RuntimeError(
                "No default cache handler configured. "
                "Call Configuration.set_default_cache_handler() first."
            )
        return cls._default_cache_handler

    @classmethod
    def set_global_prefix(cls, prefix: str) -> None:
        cls._global_prefix = prefix

    @classmethod
    def get_global_prefix(cls) -> str:
        return cls._global_prefix

    @classmethod
    def reset(cls) -> None:
        """Reset configuration to defaults. Useful for testing."""
        cls._default_cache_handler = None
        cls._global_prefix = "DefaultPrefix"
