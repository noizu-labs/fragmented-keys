defmodule FragmentedKeys.Configuration do
  @moduledoc """
  Global configuration for the fragmented keys library.

  Uses `:persistent_term` for storage, so no supervision tree is required.
  """

  @prefix_key {__MODULE__, :global_prefix}
  @handler_key {__MODULE__, :default_cache_handler}

  @doc "Set the default cache handler used by tags that don't specify one."
  def set_default_cache_handler(handler) do
    :persistent_term.put(@handler_key, handler)
  end

  @doc "Return the default cache handler. Raises if none has been configured."
  def get_default_cache_handler do
    case :persistent_term.get(@handler_key, nil) do
      nil ->
        raise RuntimeError,
              "No default cache handler configured. " <>
                "Call Configuration.set_default_cache_handler/1 first."

      handler ->
        handler
    end
  end

  @doc "Set the global prefix used in tag name generation."
  def set_global_prefix(prefix) when is_binary(prefix) do
    :persistent_term.put(@prefix_key, prefix)
  end

  @doc "Return the global prefix. Defaults to `\"DefaultPrefix\"`."
  def get_global_prefix do
    :persistent_term.get(@prefix_key, "DefaultPrefix")
  end

  @doc "Reset configuration to defaults. Useful for testing."
  def reset do
    :persistent_term.erase(@handler_key)
    :persistent_term.erase(@prefix_key)
  end
end
