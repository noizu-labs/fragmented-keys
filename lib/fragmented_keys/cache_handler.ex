defprotocol FragmentedKeys.CacheHandler do
  @moduledoc """
  Protocol for cache backend implementations.
  """

  @doc "Return a unique identifier for this handler type."
  @spec group_name(t) :: String.t()
  def group_name(handler)

  @doc "Fetch a single value by key. Returns nil on miss."
  @spec get(t, String.t()) :: String.t() | nil
  def get(handler, key)

  @doc "Store a value, optionally with a TTL in seconds."
  @spec set(t, String.t(), String.t(), non_neg_integer() | nil) :: :ok
  def set(handler, key, value, ttl)

  @doc "Fetch multiple values at once. Returns map of found key-value pairs."
  @spec get_multi(t, [String.t()]) :: %{String.t() => String.t()}
  def get_multi(handler, keys)
end
