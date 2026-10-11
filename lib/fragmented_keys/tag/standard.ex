defmodule FragmentedKeys.Tag.Standard do
  @moduledoc """
  A tag whose version is stored in cache and can be incremented.

  Incrementing the version causes all composite keys that include this
  tag-instance to resolve to a new cache key, effectively invalidating
  the old cached values without deleting them.
  """

  defstruct [:tag, :instance, :version, :handler, :prefix]

  @type t :: %__MODULE__{
          tag: String.t(),
          instance: String.t(),
          version: float() | nil,
          handler: struct() | nil,
          prefix: String.t()
        }

  @doc """
  Create a new standard tag.

  ## Options

    * `:version` — initial version (default: fetched from cache on first access)
    * `:handler` — cache handler override (default: global default)
    * `:prefix` — tag name prefix (default: global prefix)
  """
  def new(tag, instance \\ "", opts \\ []) do
    %__MODULE__{
      tag: tag,
      instance: to_string(instance),
      version: opts[:version],
      handler: opts[:handler],
      prefix: opts[:prefix] || FragmentedKeys.Configuration.get_global_prefix()
    }
  end
end
