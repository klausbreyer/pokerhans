defmodule Pokerhans.View do
  @moduledoc "Escaped server-rendered versions of the existing Pokerhans templates."
  require EEx

  for name <- ~w(layout home season leaderboard) do
    EEx.function_from_file(
      :defp,
      String.to_atom(name),
      Path.join(__DIR__, "templates/#{name}.html.eex"),
      [:assigns],
      engine: Phoenix.HTML.Engine
    )
  end

  @doc "Renders a known template with HTML escaping."
  def render("layout", data), do: layout(data: data)
  def render("home", data), do: home(data: data)
  def render("season", data), do: season(data: data)
  def render("leaderboard", data), do: leaderboard(data: data)

  defp present?(value), do: value not in [nil, false, []]
end
