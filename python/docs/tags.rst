Tags
====

Tags are versioned identifiers that form the building blocks of composite cache keys.
Each tag has a name, an instance, and a version. When a tag's version changes, every
composite key that includes it resolves to a different hash.

StandardTag
-----------

Version is stored in the cache backend and can be incremented. Each increment changes
the version by ``+0.1``, invalidating all dependent keys.

.. code-block:: python

   from fragmented_keys import StandardTag

   tag = StandardTag("User", "42")
   tag.get_tag_version()    # Fetches or seeds version from cache
   tag.increment()          # Version += 0.1, persisted to cache
   tag.reset_tag_version()  # Resets to a new time-based value

Version Lifecycle
^^^^^^^^^^^^^^^^^

- **Seed**: First access generates ``time.time() * 1000`` (milliseconds), stored in cache.
- **Read**: Subsequent reads return the cached version.
- **Increment**: ``version += 0.1``, persisted immediately.
- **Reset**: Replaced with a fresh millisecond timestamp.

ConstantTag
-----------

Version is fixed at construction time. Mutations are no-ops. Useful for incorporating
static dimensions into composite keys.

.. code-block:: python

   from fragmented_keys import ConstantTag

   tag = ConstantTag("ApiVersion", "v2", version=2.0)
   tag.increment()          # No-op
   tag.get_tag_version()    # Always 2.0

Constant tags return ``False`` from :meth:`delegate_cache_query`, so they are never
included in bulk ``get_multi`` calls — their version is always known without a cache
lookup.

BaseTag
-------

Both ``StandardTag`` and ``ConstantTag`` inherit from
:class:`~fragmented_keys.tag.base.BaseTag`, which provides:

- Version storage and retrieval against a cache backend
- Cache handler resolution (per-tag override or global default)
- Tag name generation: ``{tag}_{instance}{prefix}``

Tag Interface
-------------

All tags satisfy the :class:`~fragmented_keys.protocols.Tag` protocol:

.. list-table::
   :header-rows: 1
   :widths: 30 70

   * - Method
     - Description
   * - ``get_tag_name()``
     - Return the cache key used to store this tag's version.
   * - ``get_tag_version()``
     - Return the current version for this tag-instance.
   * - ``get_full_tag()``
     - Return the tag name with its current version.
   * - ``increment()``
     - Increment the tag version, invalidating dependent keys.
   * - ``reset_tag_version()``
     - Reset the tag version to a new value.
   * - ``set_tag_version(version, update=False)``
     - Manually set the tag version.
   * - ``set_cache_handler(handler)``
     - Override the cache handler for this tag.
   * - ``get_cache_handler()``
     - Return the cache handler used by this tag.
   * - ``delegate_cache_query(group)``
     - Check if this tag's version can be bulk-fetched with the given group.
