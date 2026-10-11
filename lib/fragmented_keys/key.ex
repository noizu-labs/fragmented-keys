defmodule FragmentedKeys.Key do
  @moduledoc """
  A composite cache key built from multiple versioned tags.

  The final key string is an MD5 hash of the base key name combined
  with each tag's name and current version, so that any tag version
  change produces a completely different key.
  """

  alias FragmentedKeys.{CacheHandler, Tag}

  defstruct [:key, :group_id, tags: []]

  @type t :: %__MODULE__{
          key: String.t(),
          group_id: String.t(),
          tags: [struct()]
        }

  @doc "Create a new composite key from a name and list of tags."
  def new(key, tags \\ [], opts \\ []) do
    %__MODULE__{
      key: key,
      tags: tags,
      group_id: opts[:group_id] || ""
    }
  end

  @doc "Append a tag to this key's composition. Returns updated key."
  def add_tag(%__MODULE__{tags: tags} = k, tag) do
    %{k | tags: tags ++ [tag]}
  end

  @doc """
  Build the composite key string.

  When `hash` is `true` (default) the key is returned as an MD5
  hex digest suitable for use as a cache key. Pass `false` to
  get the raw string for debugging.
  """
  def get_key_str(%__MODULE__{} = key, hash \\ true) do
    tags = gather_group_versions(key.tags)

    tag_parts =
      Enum.map(tags, fn tag ->
        ":t#{Tag.get_tag_name(tag)}:v#{Tag.get_tag_version(tag)}"
      end)

    raw = "#{key.key}_#{key.group_id}" <> Enum.join(tag_parts)

    if hash do
      :crypto.hash(:md5, raw) |> Base.encode16(case: :lower)
    else
      raw
    end
  end

  # -- bulk version fetching -------------------------------------------------

  defp gather_group_versions(tags) do
    # Group delegatable tags by handler group name
    {grouped, handler_map} =
      Enum.reduce(tags, {%{}, %{}}, fn tag, {groups, handlers} ->
        handler = Tag.get_cache_handler(tag)
        group = CacheHandler.group_name(handler)

        if Tag.delegate_cache_query(tag, group) do
          groups = Map.update(groups, group, [tag], &(&1 ++ [tag]))
          handlers = Map.put_new(handlers, group, handler)
          {groups, handlers}
        else
          {groups, handlers}
        end
      end)

    # Bulk fetch versions for each handler group
    version_map =
      Enum.reduce(grouped, %{}, fn {group, group_tags}, acc ->
        handler = Map.fetch!(handler_map, group)
        keys = Enum.map(group_tags, &Tag.get_tag_name/1)
        results = CacheHandler.get_multi(handler, keys)
        Map.merge(acc, results)
      end)

    # Apply fetched versions to tag copies
    Enum.map(tags, fn tag ->
      name = Tag.get_tag_name(tag)

      case Map.get(version_map, name) do
        nil -> tag
        raw -> %{tag | version: Tag.parse_version(raw)}
      end
    end)
  end
end
