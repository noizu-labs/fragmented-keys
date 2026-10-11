Configuration
=============

The :class:`~fragmented_keys.configuration.Configuration` class provides global defaults
shared across all tags and keys.

.. code-block:: python

   from fragmented_keys import Configuration

   Configuration.set_default_cache_handler(handler)  # Fallback handler
   Configuration.set_global_prefix("MyApp")           # Key namespace prefix

Methods
-------

.. list-table::
   :header-rows: 1
   :widths: 40 60

   * - Method
     - Description
   * - ``set_default_cache_handler(handler)``
     - Set the default cache handler used by tags that don't specify one.
   * - ``get_default_cache_handler()``
     - Return the default cache handler. Raises ``RuntimeError`` if not set.
   * - ``set_global_prefix(prefix)``
     - Set the global prefix for tag name generation.
   * - ``get_global_prefix()``
     - Return the current global prefix (default: ``"DefaultPrefix"``).
   * - ``reset()``
     - Reset all configuration to defaults. Useful for testing.
