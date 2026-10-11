defmodule FragmentedKeys.KeyRing do
  @moduledoc """
  Factory for creating and managing related keys with predefined configurations.

  Define key templates once with `define_key/3`, then retrieve configured
  `Key` objects with `get_key_obj/3` or via `key_obj/3`.
  """

  alias FragmentedKeys.{CacheHandler, Key}
  alias FragmentedKeys.Tag.{Standard, Constant}

  defstruct [
    :default_handler_name,
    :default_prefix,
    global_options: %{},
    global_tag_options: %{},
    cache_handlers: %{},
    key_definitions: %{}
  ]

  @type t :: %__MODULE__{}

  @doc """
  Create a new KeyRing.

  ## Options

    * `:cache_handlers` — map of name to handler struct (default: `%{"memory" => Memory.new()}`)
    * `:default_cache_handler` — name of the default handler (default: `"memory"`)
    * `:default_prefix` — prefix for tag names (default: `"DefaultPrefix"`)
    * `:global_options` — options merged into every tag (default: `%{}`)
    * `:global_tag_options` — per-tag-name option overrides (default: `%{}`)
  """
  def new(opts \\ []) do
    default_handler_name = Keyword.get(opts, :default_cache_handler, "memory")
    default_prefix = Keyword.get(opts, :default_prefix, "DefaultPrefix")

    cache_handlers =
      Keyword.get(opts, :cache_handlers, %{
        "memory" => CacheHandler.Memory.new()
      })

    %__MODULE__{
      global_options: Keyword.get(opts, :global_options, %{}),
      global_tag_options: Keyword.get(opts, :global_tag_options, %{}),
      default_handler_name: default_handler_name,
      default_prefix: default_prefix,
      cache_handlers: cache_handlers,
      key_definitions: %{}
    }
  end

  # -- handler resolution ----------------------------------------------------

  defp resolve_handler(%__MODULE__{} = ring, nil) do
    resolve_handler(ring, ring.default_handler_name)
  end

  defp resolve_handler(%__MODULE__{cache_handlers: handlers}, name) do
    case Map.get(handlers, name) do
      nil -> raise ArgumentError, "Unknown cache handler: #{inspect(name)}"
      handler -> handler
    end
  end

  # -- tag options -----------------------------------------------------------

  @doc "Set option overrides for a specific tag name."
  def set_tag_options(%__MODULE__{} = ring, tag, options) do
    %{ring | global_tag_options: Map.put(ring.global_tag_options, tag, options)}
  end

  @doc "Return merged options for a tag: global -> per-tag -> extra."
  def get_tag_options(%__MODULE__{} = ring, tag, extra \\ %{}) do
    ring.global_options
    |> Map.merge(Map.get(ring.global_tag_options, tag, %{}))
    |> Map.merge(extra)
  end

  @doc "Replace the global options applied to all tags."
  def set_global_options(%__MODULE__{} = ring, options) do
    %{ring | global_options: options}
  end

  @doc "Return the current global options map."
  def get_global_options(%__MODULE__{} = ring), do: ring.global_options

  # -- tag factory -----------------------------------------------------------

  @doc "Create a tag instance with merged options from the ring's configuration."
  def tag(%__MODULE__{} = ring, tag_name, instance, options \\ %{}) do
    opts = get_tag_options(ring, tag_name, options)
    handler = resolve_handler(ring, opts["cache_handler"])
    prefix = opts["prefix"] || ring.default_prefix
    tag_type = opts["type"] || "standard"
    version = opts["version"]

    case tag_type do
      "constant" ->
        Constant.new(tag_name, instance,
          version: version || 1.0,
          handler: handler,
          prefix: prefix
        )

      _ ->
        Standard.new(tag_name, instance,
          version: version,
          handler: handler,
          prefix: prefix
        )
    end
  end

  # -- key definitions -------------------------------------------------------

  @doc """
  Define a reusable key template. Returns updated ring.

  `params` is a list where each element is either a tag name string
  or a map with at least a `"tag"` key and optional overrides
  (`"cache_handler"`, `"type"`, `"version"`, `"prefix"`).
  """
  def define_key(%__MODULE__{} = ring, key, params, globals \\ %{}) do
    %{ring | key_definitions: Map.put(ring.key_definitions, key, %{params: params, globals: globals})}
  end

  @doc "Return a `Key` populated from a defined template."
  def get_key_obj(%__MODULE__{} = ring, key, tag_values) do
    defn = Map.get(ring.key_definitions, key)

    unless defn do
      raise ArgumentError, "Key #{inspect(key)} is not defined"
    end

    params = defn.params

    if length(tag_values) != length(params) do
      raise ArgumentError,
            "Key #{inspect(key)} expects #{length(params)} tag values, got #{length(tag_values)}"
    end

    tags =
      Enum.zip(params, tag_values)
      |> Enum.map(fn {param, value} ->
        case param do
          name when is_binary(name) ->
            tag(ring, name, value, defn.globals)

          %{"tag" => tag_name} = param_map ->
            extra =
              param_map
              |> Map.delete("tag")
              |> Map.merge(defn.globals)

            tag(ring, tag_name, value, extra)
        end
      end)

    Key.new(key, tags)
  end

  @doc """
  Dynamic key object accessor.

  Supports `key_obj(ring, :users, ["val1", "val2"])` style access
  where `:users` or `"users"` matches a defined key name (case-insensitive).
  """
  def key_obj(%__MODULE__{} = ring, key_name, tag_values) when is_atom(key_name) do
    key_obj(ring, Atom.to_string(key_name), tag_values)
  end

  def key_obj(%__MODULE__{} = ring, key_name, tag_values) when is_binary(key_name) do
    matched =
      Enum.find(Map.keys(ring.key_definitions), fn defined ->
        String.downcase(defined) == String.downcase(key_name) ||
          String.downcase(defined) == String.downcase(String.replace(key_name, "_", ""))
      end)

    unless matched do
      raise ArgumentError, "No key defined matching #{inspect(key_name)}"
    end

    get_key_obj(ring, matched, tag_values)
  end
end
