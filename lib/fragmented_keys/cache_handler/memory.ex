defmodule FragmentedKeys.CacheHandler.Memory do
  @moduledoc """
  In-memory cache handler backed by an Agent. For testing and temporary caching.

  Each `new/0` call creates an independent cache instance.
  """

  defstruct [:pid]

  @type t :: %__MODULE__{pid: pid()}

  @doc "Create a new independent in-memory cache instance. Accepts `Agent.start_link/2` options."
  def new(opts \\ []) do
    {:ok, pid} = Agent.start_link(fn -> %{} end, opts)
    %__MODULE__{pid: pid}
  end

  @doc "Clear all cached values."
  def clear(%__MODULE__{pid: pid}) do
    Agent.update(pid, fn _ -> %{} end)
  end

  @doc "Stop the underlying Agent process."
  def stop(%__MODULE__{pid: pid}) do
    Agent.stop(pid)
  end
end

defimpl FragmentedKeys.CacheHandler, for: FragmentedKeys.CacheHandler.Memory do
  def group_name(_handler), do: "MemoryHandler"

  def get(%{pid: pid}, key) do
    Agent.get(pid, &Map.get(&1, key))
  end

  def set(%{pid: pid}, key, value, _ttl) do
    Agent.update(pid, &Map.put(&1, key, value))
    :ok
  end

  def get_multi(%{pid: pid}, keys) do
    Agent.get(pid, fn cache ->
      for k <- keys, v = Map.get(cache, k), v != nil, into: %{}, do: {k, v}
    end)
  end
end
