from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from fragmented_keys.protocols import CacheHandler

from fragmented_keys.tag.base import BaseTag


class ConstantTag(BaseTag):
    """A tag with a fixed version that never changes.

    Useful for incorporating non-versionable data into composite keys.
    Calls to *increment*, *reset_tag_version*, and *set_tag_version*
    are no-ops.
    """

    def __init__(
        self,
        tag: str,
        instance: str = "",
        version: float = 1.0,
        handler: CacheHandler | None = None,
        prefix: str | None = None,
    ) -> None:
        super().__init__(tag, instance, version, handler, prefix)

    def get_tag_version(self) -> float:
        # Always return the fixed version; never hits cache.
        assert self._version is not None
        return self._version

    def increment(self) -> None:
        pass  # no-op

    def reset_tag_version(self) -> None:
        pass  # no-op

    def set_tag_version(self, version: float, update: bool = False) -> None:
        pass  # no-op

    def delegate_cache_query(self, group: str) -> bool:
        # Constant tags never need cache lookups for their version.
        return False
