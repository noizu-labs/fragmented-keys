from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from redis import Redis


class RedisHandler:
    """Redis cache handler."""

    def __init__(self, client: Redis) -> None:
        self._client = client

    def group_name(self) -> str:
        return "RedisHandler"

    def get(self, key: str) -> str | None:
        val = self._client.get(key)
        if val is None:
            return None
        return val if isinstance(val, str) else val.decode("utf-8")

    def set(self, key: str, value: str, ttl: int | None = None) -> None:
        if ttl is not None:
            self._client.setex(key, ttl, value)
        else:
            self._client.set(key, value)

    def get_multi(self, keys: list[str]) -> dict[str, str]:
        if not keys:
            return {}
        values = self._client.mget(keys)
        result: dict[str, str] = {}
        for k, v in zip(keys, values):
            if v is not None:
                result[k] = v if isinstance(v, str) else v.decode("utf-8")
        return result
