from __future__ import annotations

import time
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from fragmented_keys.protocols import CacheHandler

from fragmented_keys.configuration import Configuration


class BaseTag:
    """Base class for tag-instance version management.

    A tag represents a logical grouping whose version is stored in a cache
    backend.  Composite cache keys incorporate the tag's current version so
    that incrementing a tag automatically invalidates every key that depends
    on it.
    """

    def __init__(
        self,
        tag: str,
        instance: str = "",
        version: float | None = None,
        handler: CacheHandler | None = None,
        prefix: str | None = None,
    ) -> None:
        self._tag = tag
        self._instance = str(instance)
        self._version = version
        self._handler = handler
        self._prefix = prefix if prefix is not None else Configuration.get_global_prefix()

    # -- name helpers ----------------------------------------------------------

    def get_tag_name(self) -> str:
        """Return the cache key used to store this tag's version."""
        return f"{self._tag}_{self._instance}{self._prefix}"

    def get_full_tag(self) -> str:
        """Return a human-readable string of tag name + current version."""
        return f"{self.get_tag_name()}@{self.get_tag_version()}"

    # -- cache handler ---------------------------------------------------------

    def get_cache_handler(self) -> CacheHandler:
        if self._handler is not None:
            return self._handler
        return Configuration.get_default_cache_handler()

    def set_cache_handler(self, handler: CacheHandler) -> None:
        self._handler = handler

    def delegate_cache_query(self, group: str) -> bool:
        """Return True if this tag's handler matches *group* (bulk-fetch)."""
        return self.get_cache_handler().group_name() == group

    # -- version persistence ---------------------------------------------------

    def _get_version(self) -> float:
        """Retrieve the version from cache, initialising if absent."""
        handler = self.get_cache_handler()
        raw = handler.get(self.get_tag_name())
        if raw is not None:
            return float(raw)
        # First access – seed with current time in ms for uniqueness.
        version = time.time() * 1000
        handler.set(self.get_tag_name(), str(version))
        return version

    def _store_version(self, version: float) -> None:
        self.get_cache_handler().set(self.get_tag_name(), str(version))

    # -- public API (overridden by subclasses) ---------------------------------

    def get_tag_version(self) -> float:
        if self._version is not None:
            return self._version
        self._version = self._get_version()
        return self._version

    def set_tag_version(self, version: float, update: bool = False) -> None:
        self._version = version
        if update:
            self._store_version(version)

    def increment(self) -> None:
        raise NotImplementedError

    def reset_tag_version(self) -> None:
        raise NotImplementedError
