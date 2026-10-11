Keys
====

Composite cache keys are built from multiple versioned tags. The key string is an MD5
hash of the base key name combined with each tag's name and current version, so that
any tag version change produces a completely different key.

StandardKey
-----------

.. code-block:: python

   from fragmented_keys import StandardKey, StandardTag

   user_tag = StandardTag("User", "42")
   city_tag = StandardTag("City", "chicago")

   key = StandardKey("UserDashboard", [user_tag, city_tag])

   # Get the MD5 hash (suitable for use as a cache key)
   cache_key = key.get_key_str()

   # Get the raw string for debugging
   raw = key.get_key_str(hash=False)

Constructor
^^^^^^^^^^^

.. code-block:: python

   StandardKey(
       key: str,                          # Base key name
       tags: list[BaseTag] | None = None, # Initial tags
       group_id: str = "",                # Optional grouping identifier
   )

Methods
^^^^^^^

.. list-table::
   :header-rows: 1
   :widths: 35 65

   * - Method
     - Description
   * - ``get_key_str(hash=True)``
     - Build the composite key. Returns MD5 hex digest by default, or the raw string
       when ``hash=False``.
   * - ``add_tag(tag)``
     - Append a tag to this key's composition.

Bulk Version Fetching
^^^^^^^^^^^^^^^^^^^^^

When ``get_key_str()`` is called, tags are grouped by their cache handler's
``group_name()``. Each group's tag versions are fetched with a single ``get_multi()``
call for efficiency.

Tags where ``delegate_cache_query()`` returns ``False`` (e.g. ``ConstantTag``) are
skipped from the bulk fetch since their version is already known.

Key Format
^^^^^^^^^^

The raw key string follows this pattern:

.. code-block:: text

   {keyName}_{groupId}:t{tag1Name}:v{version1}:t{tag2Name}:v{version2}...

The final cache key is the MD5 hex digest of this string.
