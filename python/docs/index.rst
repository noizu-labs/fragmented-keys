fragmented-keys
================

Fragmented Key Management and Invalidation Library for use with Redis.

A Python port of `noizu-labs/fragmented-keys <https://github.com/noizu-labs/fragmented-keys>`_ (PHP),
adapted to use Redis as the cache backend.

Fragmented Keys enable efficient cache invalidation by composing cache keys from multiple
independently versioned **tags**. Instead of deleting cached entries individually, you increment
a tag's version — all keys that depend on that tag automatically resolve to a new cache key,
leaving the old entries to expire naturally.

.. code-block:: text

   Cache Key = md5( base_key + tag1:version1 + tag2:version2 + ... )

When any tag version changes, the resulting key changes, producing a cache miss and triggering
a fresh data fetch.

.. toctree::
   :maxdepth: 2
   :caption: Contents

   quickstart
   tags
   keys
   keyring
   cache_handlers
   configuration
   api
