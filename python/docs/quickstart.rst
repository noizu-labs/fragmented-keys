Quick Start
===========

Installation
------------

.. code-block:: bash

   uv add fragmented-keys

Or with pip:

.. code-block:: bash

   pip install fragmented-keys

Basic Usage
-----------

.. code-block:: python

   from redis import Redis
   from fragmented_keys import Configuration, RedisHandler, StandardTag, StandardKey

   # Setup
   client = Redis(host="127.0.0.1", port=6379)
   Configuration.set_default_cache_handler(RedisHandler(client))
   Configuration.set_global_prefix("MyApp")

   # Create tags
   user_tag = StandardTag("User", "42")
   city_tag = StandardTag("City", "chicago")

   # Build a composite key
   key = StandardKey("UserDashboard", [user_tag, city_tag])
   cache_key = key.get_key_str()  # MD5 hex digest

   # Use it with Redis
   data = client.get(cache_key)
   if data is None:
       data = fetch_from_database()
       client.set(cache_key, data)

   # Invalidate all keys depending on User:42
   user_tag = StandardTag("User", "42")
   user_tag.increment()
   # Next call to get_key_str() with User:42 will produce a different key

How It Works
------------

1. Each :class:`~fragmented_keys.tag.standard.StandardTag` stores its current version in the
   cache backend under the key ``{tag}_{instance}{prefix}``.

2. When building a composite key, :class:`~fragmented_keys.key.standard.StandardKey` bulk-fetches
   all tag versions using ``get_multi`` (grouped by handler) for efficiency.
   :class:`~fragmented_keys.tag.constant.ConstantTag` instances are skipped in bulk fetches since
   their version is fixed.

3. The final key is ``md5("{keyName}_{groupId}:t{tag1Name}:v{version1}:t{tag2Name}:v{version2}...")``.

4. Calling ``tag.increment()`` bumps the stored version by ``0.1``, so subsequent key builds
   produce a different MD5 — effectively invalidating all cached data under the old key without
   deleting it.
