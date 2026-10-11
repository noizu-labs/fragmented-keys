KeyRing
=======

:class:`~fragmented_keys.key_ring.FragmentedKeyRing` is a factory for defining reusable
key templates and producing configured :class:`~fragmented_keys.key.standard.StandardKey`
instances.

Basic Usage
-----------

.. code-block:: python

   from redis import Redis
   from fragmented_keys import FragmentedKeyRing, RedisHandler, MemoryHandler

   redis_handler = RedisHandler(Redis())
   memory_handler = MemoryHandler()

   ring = FragmentedKeyRing(
       global_options={"type": "standard"},
       global_tag_options={
           "universe": {"type": "constant", "version": 1.0},
       },
       default_cache_handler="redis",
       cache_handlers={
           "redis": redis_handler,
           "memory": memory_handler,
       },
       default_prefix="MyApp",
   )

Defining Key Templates
----------------------

Use ``define_key`` to register a template. Each parameter is either a tag name string
or a dict with at least a ``"tag"`` key and optional overrides.

.. code-block:: python

   ring.define_key("Users", [
       "universe",
       {"tag": "planet", "cache_handler": "memory"},
       "city",
   ])

Retrieving Key Objects
----------------------

.. code-block:: python

   # Positional values map 1:1 to the defined parameters
   key_obj = ring.get_key_obj("Users", ["MilkyWay", "Earth", "Chicago"])
   cache_key = key_obj.get_key_str()

Dynamic Accessor
^^^^^^^^^^^^^^^^

A convenience syntax is available via ``__getattr__``:

.. code-block:: python

   # Equivalent to ring.get_key_obj("Users", ["MilkyWay", "Earth", "Chicago"])
   key_obj = ring.get_users_key_obj("MilkyWay", "Earth", "Chicago")

The method name follows the pattern ``get_<key_name>_key_obj``. Key name matching
is case-insensitive and ignores underscores.

Tag Options
-----------

Options are merged in priority order:

.. code-block:: text

   global_options → global_tag_options[tag_name] → per-key overrides

.. list-table::
   :header-rows: 1
   :widths: 20 20 60

   * - Option
     - Type
     - Description
   * - ``type``
     - ``"standard"`` | ``"constant"``
     - Tag type (default: ``"standard"``)
   * - ``version``
     - ``float``
     - Initial version value
   * - ``cache_handler``
     - ``str``
     - Handler key from ``cache_handlers`` dict
   * - ``prefix``
     - ``str``
     - Override global prefix for this tag

Constructor
-----------

.. code-block:: python

   FragmentedKeyRing(
       global_options: dict | None = None,        # Default options for all tags
       global_tag_options: dict | None = None,     # Per-tag-name option overrides
       default_cache_handler: str = "memory",      # Default handler key
       cache_handlers: dict | None = None,         # Named handler instances
       default_prefix: str = "DefaultPrefix",      # Key namespace prefix
   )
