from __future__ import annotations

import time
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from fragmented_keys.protocols import CacheHandler

from fragmented_keys.tag.base import BaseTag


class StandardTag(BaseTag):
    """A tag whose version is stored in cache and can be incremented.

    Incrementing the version causes all composite keys that include this
    tag-instance to resolve to a new cache key, effectively invalidating
    the old cached values without deleting them.
    """

    def __init__(
        self,
        tag: str,
        instance: str = "",
        version: float | None = None,
        handler: CacheHandler | None = None,
        prefix: str | None = None,
    ) -> None:
        super().__init__(tag, instance, version, handler, prefix)

    def increment(self) -> None:
        """Increment the tag version by 0.1 and persist."""
        current = self.get_tag_version()
        new_version = current + 0.1
        self._version = new_version
        self._store_version(new_version)

    def reset_tag_version(self) -> None:
        """Reset the tag version to a new microtime-based value."""
        new_version = time.time() * 1000
        self._version = new_version
        self._store_version(new_version)
