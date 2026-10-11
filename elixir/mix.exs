defmodule FragmentedKeys.MixProject do
  use Mix.Project

  @version "0.1.1"
  @source_url "https://github.com/noizu-labs/fragmented-keys-ex"

  def project do
    [
      app: :fragmented_keys,
      version: @version,
      elixir: "~> 1.19",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      description: "Fragmented Key Management and Invalidation Library",
      package: package(),
      docs: docs(),
      source_url: @source_url,
      homepage_url: @source_url
    ]
  end

  def application do
    [
      extra_applications: [:logger, :crypto]
    ]
  end

  defp deps do
    [
      {:redix, "~> 1.5", optional: true},
      {:ex_doc, "~> 0.35", only: :dev, runtime: false}
    ]
  end

  defp package do
    [
      licenses: ["MIT"],
      files: ~w(lib .formatter.exs mix.exs README.md LICENSE),
      links: %{
        "GitHub" => @source_url,
        "PHP Version" => "https://github.com/noizu-labs/fragmented-keys",
        "Python Version" => "https://github.com/noizu-labs/fragmented-keys-py",
        "Java Version" => "https://github.com/noizu-labs/fragmented-keys-4java"
      }
    ]
  end

  defp docs do
    [
      main: "FragmentedKeys",
      before_closing_head_tag: &before_closing_head_tag/1,
      extras: [
        "README.md",
        "docs/PROJ-ARCH.md",
        "docs/arch/invalidation.md",
        "docs/arch/cache-handler.md",
        "docs/PROJ-LAYOUT.md",
        "docs/layout/lib.md",
        "LICENSE"
      ],
      groups_for_extras: [
        Architecture: [
          "docs/PROJ-ARCH.md",
          "docs/arch/invalidation.md",
          "docs/arch/cache-handler.md"
        ],
        Layout: [
          "docs/PROJ-LAYOUT.md",
          "docs/layout/lib.md"
        ]
      ],
      source_ref: "v#{@version}",
      groups_for_modules: [
        "Cache Handlers": [
          FragmentedKeys.CacheHandler,
          FragmentedKeys.CacheHandler.Memory,
          FragmentedKeys.CacheHandler.Redis
        ],
        Tags: [
          FragmentedKeys.Tag,
          FragmentedKeys.Tag.Standard,
          FragmentedKeys.Tag.Constant
        ],
        Keys: [
          FragmentedKeys.Key,
          FragmentedKeys.KeyRing
        ]
      ]
    ]
  end

  defp before_closing_head_tag(:html) do
    """
    <script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
    <script>
      document.addEventListener("DOMContentLoaded", function () {
        mermaid.initialize({startOnLoad: false, theme: "default"});
        let id = 0;
        for (const codeEl of document.querySelectorAll("pre code.mermaid")) {
          const preEl = codeEl.parentElement;
          const graphDefinition = codeEl.textContent;
          const graphEl = document.createElement("div");
          const graphId = "mermaid-graph-" + id++;
          mermaid.render(graphId, graphDefinition).then(({svg, bindFunctions}) => {
            graphEl.innerHTML = svg;
            bindFunctions?.(graphEl);
            preEl.insertAdjacentElement("afterend", graphEl);
            preEl.remove();
          });
        }
      });
    </script>
    """
  end

  defp before_closing_head_tag(_), do: ""
end
