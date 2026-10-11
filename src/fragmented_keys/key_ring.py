from __future__ import annotations

from typing import Any

from fragmented_keys.cache_handler.memory import MemoryHandler
from fragmented_keys.key.standard import StandardKey
from fragmented_keys.tag.constant import ConstantTag
from fragmented_keys.tag.standard import StandardTag
from fragmented_keys.tag.base import BaseTag


class FragmentedKeyRing:
    """Factory for creating and managing related keys with predefined configurations.

    Define key templates once with ``define_key``, then retrieve configured
    ``StandardKey`` objects with ``get_key_obj`` or the convenience
    ``get_<name>_key_obj(...)`` dynamic method.
    """

    def __init__(
        self,
        global_options: dict[str, Any] | None = None,
        global_tag_options: dict[str, dict[str, Any]] | None = None,
        default_cache_handler: str = "memory",
        cache_handlers: dict[str, Any] | None = None,
        default_prefix: str = "DefaultPrefix",
    ) -> None:
        self._global_options: dict[str, Any] = global_options or {}
        self._global_tag_options: dict[str, dict[str, Any]] = global_tag_options or {}
        self._default_handler_name = default_cache_handler
        self._cache_handlers: dict[str, Any] = cache_handlers or {
            "memory": MemoryHandler()
        }
        self._default_prefix = default_prefix
        self._key_definitions: dict[str, dict] = {}

    # -- handler resolution ----------------------------------------------------

    def _resolve_handler(self, name: str | None = None) -> Any:
        name = name or self._default_handler_name
        handler = self._cache_handlers.get(name)
        if handler is None:
            raise ValueError(f"Unknown cache handler: {name!r}")
        return handler

    # -- tag options -----------------------------------------------------------

    def set_tag_options(self, tag: str, options: dict[str, Any]) -> None:
        self._global_tag_options[tag] = options

    def get_tag_options(self, tag: str, extra: dict[str, Any] | None = None) -> dict[str, Any]:
        opts: dict[str, Any] = {}
        opts.update(self._global_options)
        opts.update(self._global_tag_options.get(tag, {}))
        if extra:
            opts.update(extra)
        return opts

    def set_global_options(self, options: dict[str, Any]) -> None:
        self._global_options = options

    def get_global_options(self) -> dict[str, Any]:
        return dict(self._global_options)

    # -- tag factory -----------------------------------------------------------

    def tag(self, tag: str, instance: str, options: dict[str, Any] | None = None) -> BaseTag:
        """Create a tag instance with merged options."""
        opts = self.get_tag_options(tag, options)
        handler = self._resolve_handler(opts.get("cache_handler"))
        prefix = opts.get("prefix", self._default_prefix)
        tag_type = opts.get("type", "standard")
        version = opts.get("version")

        if tag_type == "constant":
            return ConstantTag(
                tag=tag,
                instance=instance,
                version=version if version is not None else 1.0,
                handler=handler,
                prefix=prefix,
            )
        return StandardTag(
            tag=tag,
            instance=instance,
            version=version,
            handler=handler,
            prefix=prefix,
        )

    # -- key definitions -------------------------------------------------------

    def define_key(
        self,
        key: str,
        params: list[str | dict[str, Any]],
        globals: dict[str, Any] | None = None,
    ) -> None:
        """Define a reusable key template.

        *params* is a list where each element is either a tag name string
        or a dict with at least a ``"tag"`` key and optional overrides
        (``cache_handler``, ``type``, ``version``, ``prefix``).
        """
        self._key_definitions[key] = {
            "params": params,
            "globals": globals or {},
        }

    def get_key_obj(self, key: str, tag_values: list[str]) -> StandardKey:
        """Return a ``StandardKey`` populated from a defined template."""
        defn = self._key_definitions.get(key)
        if defn is None:
            raise ValueError(f"Key {key!r} is not defined")

        params = defn["params"]
        extra_globals = defn["globals"]

        if len(tag_values) != len(params):
            raise ValueError(
                f"Key {key!r} expects {len(params)} tag values, got {len(tag_values)}"
            )

        tags: list[BaseTag] = []
        for param, value in zip(params, tag_values):
            if isinstance(param, str):
                opts = dict(extra_globals)
                tag_obj = self.tag(param, value, opts)
            else:
                opts = dict(extra_globals)
                opts.update({k: v for k, v in param.items() if k != "tag"})
                tag_name = param["tag"]
                tag_obj = self.tag(tag_name, value, opts)
            tags.append(tag_obj)

        return StandardKey(key=key, tags=tags)

    # -- dynamic access --------------------------------------------------------

    def __getattr__(self, name: str) -> Any:
        """Support ``ring.get_<key>_key_obj(val1, val2, ...)`` syntax."""
        if name.startswith("get_") and name.endswith("_key_obj"):
            key_name = name[4:-8]  # strip get_ and _key_obj
            # Convert snake_case to the original key name
            # Try exact match first, then title-case
            matched = None
            for defined in self._key_definitions:
                if defined.lower() == key_name.lower().replace("_", ""):
                    matched = defined
                    break
                if defined.lower() == key_name.lower():
                    matched = defined
                    break
            if matched is None:
                raise AttributeError(f"No key defined matching {key_name!r}")

            def factory(*args: str) -> StandardKey:
                return self.get_key_obj(matched, list(args))

            return factory

        raise AttributeError(f"{type(self).__name__!r} has no attribute {name!r}")
