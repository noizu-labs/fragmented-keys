defmodule FragmentedKeys.Tag do
  @moduledoc """
  Unified interface for tag operations across Standard and Constant tags.
  """

  alias FragmentedKeys.Tag.{Standard, Constant}
  alias FragmentedKeys.{CacheHandler, Configuration}

  @doc "Return the cache key used to store this tag's version."
  def get_tag_name(%{tag: tag, instance: instance, prefix: prefix}) do
    "#{tag}_#{instance}#{prefix}"
  end

  @doc "Return the tag name with its current version."
  def get_full_tag(tag) do
    "#{get_tag_name(tag)}@#{get_tag_version(tag)}"
  end

  @doc "Return the current version for this tag-instance."
  def get_tag_version(%Constant{version: v}), do: v

  def get_tag_version(%Standard{version: v}) when not is_nil(v), do: v

  def get_tag_version(%Standard{} = tag) do
    handler = get_cache_handler(tag)
    name = get_tag_name(tag)

    case CacheHandler.get(handler, name) do
      nil ->
        version = System.system_time(:millisecond) * 1.0
        CacheHandler.set(handler, name, version_to_string(version), nil)
        version

      raw ->
        parse_version(raw)
    end
  end

  @doc "Increment the tag version, invalidating dependent keys. Returns updated tag."
  def increment(%Constant{} = tag), do: tag

  def increment(%Standard{} = tag) do
    current = get_tag_version(tag)
    new_version = current + 0.1
    handler = get_cache_handler(tag)
    CacheHandler.set(handler, get_tag_name(tag), version_to_string(new_version), nil)
    %{tag | version: new_version}
  end

  @doc "Reset the tag version to a new microtime-based value. Returns updated tag."
  def reset_tag_version(%Constant{} = tag), do: tag

  def reset_tag_version(%Standard{} = tag) do
    new_version = System.system_time(:millisecond) * 1.0
    handler = get_cache_handler(tag)
    CacheHandler.set(handler, get_tag_name(tag), version_to_string(new_version), nil)
    %{tag | version: new_version}
  end

  @doc "Manually set the tag version. Returns updated tag."
  def set_tag_version(%Constant{} = tag, _version, _update), do: tag

  def set_tag_version(%Standard{} = tag, version, update) do
    tag = %{tag | version: version}

    if update do
      handler = get_cache_handler(tag)
      CacheHandler.set(handler, get_tag_name(tag), version_to_string(version), nil)
    end

    tag
  end

  @doc "Return the cache handler used by this tag."
  def get_cache_handler(%{handler: handler}) when not is_nil(handler), do: handler
  def get_cache_handler(_tag), do: Configuration.get_default_cache_handler()

  @doc "Override the cache handler for this tag. Returns updated tag."
  def set_cache_handler(tag, handler), do: %{tag | handler: handler}

  @doc "Check if this tag's version can be bulk-fetched with the given group."
  def delegate_cache_query(%Constant{}, _group), do: false

  def delegate_cache_query(tag, group) do
    CacheHandler.group_name(get_cache_handler(tag)) == group
  end

  # -- helpers ---------------------------------------------------------------

  @doc false
  def version_to_string(v) when is_float(v) do
    :erlang.float_to_binary(v, [:compact, {:decimals, 20}])
  end

  @doc false
  def parse_version(raw) when is_binary(raw) do
    case Float.parse(raw) do
      {v, _} -> v
      :error -> raise ArgumentError, "invalid version string: #{inspect(raw)}"
    end
  end
end
