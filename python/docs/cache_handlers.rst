Cache Handlers
==============

Cache handlers are the backend storage for tag versions. They implement the
:class:`~fragmented_keys.protocols.CacheHandler` protocol.

RedisHandler
------------

Primary backend using ``redis-py``. Supports ``mget`` for bulk version fetches and
``setex`` for TTL.

.. code-block:: python

   from redis import Redis
   from fragmented_keys import RedisHandler

   handler = RedisHandler(Redis(host="127.0.0.1", port=6379))

MemoryHandler
-------------

In-memory dict-based handler for testing.

.. code-block:: python

   from fragmented_keys import MemoryHandler

   handler = MemoryHandler()
   handler.clear()  # Reset all cached values

Custom Handlers
---------------

Implement the :class:`~fragmented_keys.protocols.CacheHandler` protocol to create a
custom backend:

.. code-block:: python

   from fragmented_keys.protocols import CacheHandler

   class MyHandler:
       def group_name(self) -> str:
           return "MyHandler"

       def get(self, key: str) -> str | None:
           ...

       def set(self, key: str, value: str, ttl: int | None = None) -> None:
           ...

       def get_multi(self, keys: list[str]) -> dict[str, str]:
           ...

CacheHandler Protocol
---------------------

.. list-table::
   :header-rows: 1
   :widths: 35 65

   * - Method
     - Description
   * - ``group_name()``
     - Return a unique identifier for this handler type. Used to group tags for
       bulk ``get_multi`` calls.
   * - ``get(key)``
     - Fetch a single value by key. Returns ``None`` on miss.
   * - ``set(key, value, ttl=None)``
     - Store a value, optionally with a TTL in seconds.
   * - ``get_multi(keys)``
     - Fetch multiple values at once. Returns dict of found key-value pairs.
