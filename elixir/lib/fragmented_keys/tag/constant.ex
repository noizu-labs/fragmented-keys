defmodule FragmentedKeys.Tag.Constant do
  @moduledoc """
  A tag with a fixed version that never changes.

  Useful for incorporating non-versionable data into composite keys.
  Increment, reset, and set operations are no-ops.
  """

  defstruct [:tag, :instance, :version, :handler, :prefix]

  @type t :: %__MODULE__{
          tag: String.t(),
          instance: String.t(),
          version: float(),
          handler: struct() | nil,
          prefix: String.t()
        }

  @doc """
  Create a new constant tag.

  ## Options

    * `:version` — fixed version (default: `1.0`)
    * `:handler` — cache handler override (default: global default)
    * `:prefix` — tag name prefix (default: global prefix)
  """
  def new(tag, instance \\ "", opts \\ []) do
    %__MODULE__{
      tag: tag,
      instance: to_string(instance),
      version: opts[:version] || 1.0,
      handler: opts[:handler],
      prefix: opts[:prefix] || FragmentedKeys.Configuration.get_global_prefix()
    }
  end
end
