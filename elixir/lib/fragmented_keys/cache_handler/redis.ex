defmodule FragmentedKeys.CacheHandler.Redis do
  @moduledoc """
  Redis cache handler using Redix.

  Requires the `:redix` dependency.
  """

  defstruct [:conn]

  @type t :: %__MODULE__{conn: pid() | atom()}

  @doc "Wrap a Redix connection pid or registered name as a cache handler."
  def new(conn), do: %__MODULE__{conn: conn}
end

defimpl FragmentedKeys.CacheHandler, for: FragmentedKeys.CacheHandler.Redis do
  def group_name(_handler), do: "RedisHandler"

  def get(%{conn: conn}, key) do
    case Redix.command(conn, ["GET", key]) do
      {:ok, nil} -> nil
      {:ok, val} -> val
      _ -> nil
    end
  end

  def set(%{conn: conn}, key, value, nil) do
    Redix.command(conn, ["SET", key, value])
    :ok
  end

  def set(%{conn: conn}, key, value, ttl) do
    Redix.command(conn, ["SETEX", key, to_string(ttl), value])
    :ok
  end

  def get_multi(_handler, []), do: %{}

  def get_multi(%{conn: conn}, keys) do
    case Redix.command(conn, ["MGET" | keys]) do
      {:ok, values} ->
        keys
        |> Enum.zip(values)
        |> Enum.reject(fn {_k, v} -> is_nil(v) end)
        |> Map.new()

      _ ->
        %{}
    end
  end
end
