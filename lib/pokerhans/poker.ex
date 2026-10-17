defmodule Pokerhans.Poker do
  @moduledoc "Season history, hosting rotation and the original 3/1 scoring rules."
  alias Pokerhans.{Repo, Game}

  @doc "Lists seasons newest first."
  def seasons, do: rows("SELECT * FROM seasons ORDER BY id DESC")

  @doc "Lists players alphabetically."
  def players, do: rows("SELECT * FROM players ORDER BY name, id")

  @doc "Returns season games, including optional placement names, oldest first."
  def games(season_id) do
    rows(
      """
      SELECT g.*, h.name AS host_name, COALESCE(w.name, '') AS winner_name,
             COALESCE(s.name, '') AS second_place_name
      FROM games g JOIN players h ON h.id = g.host_id
      LEFT JOIN players w ON w.id = g.winner_id
      LEFT JOIN players s ON s.id = g.second_place_id
      WHERE season_id = $1 ORDER BY game_date, g.id
      """,
      [season_id]
    )
  end

  @doc "Ranks all wins, sharing ranks for ties and leaving the corresponding gaps."
  def overall_wins do
    rows("""
    SELECT p.id AS player_id, p.name AS player_name, COUNT(*)::bigint AS wins
    FROM games g JOIN players p ON p.id = g.winner_id
    GROUP BY p.id, p.name ORDER BY wins DESC, p.name, p.id
    """)
    |> rank([:wins])
  end

  @doc "Ranks three points for wins and one for second places from season four."
  def points do
    rows("""
    SELECT p.id AS player_id, p.name AS player_name,
      SUM(score.points)::bigint AS points, SUM(score.wins)::bigint AS wins,
      SUM(score.seconds)::bigint AS seconds
    FROM (
      SELECT winner_id AS player_id, 3 AS points, 1 AS wins, 0 AS seconds
      FROM games WHERE season_id >= 4 AND winner_id IS NOT NULL
      UNION ALL
      SELECT second_place_id, 1, 0, 1
      FROM games WHERE season_id >= 4 AND second_place_id IS NOT NULL
    ) score JOIN players p ON p.id = score.player_id
    GROUP BY p.id, p.name
    ORDER BY points DESC, wins DESC, seconds DESC, p.name, p.id
    """)
    |> rank([:points, :wins, :seconds])
  end

  @doc "Adds a game to the latest season."
  def add_game(attrs) do
    changeset = Game.changeset(%Game{}, attrs)

    if latest_season?(Ecto.Changeset.get_field(changeset, :season_id)),
      do: Repo.insert(changeset),
      else: {:error, :season_closed}
  end

  @doc "Changes a game's date only when its season matches the latest season."
  def update_date(game_id, season_id, date) do
    case Repo.get(Game, game_id) do
      %Game{season_id: ^season_id} = game ->
        if latest_season?(season_id),
          do: game |> Game.changeset(%{"game_date" => date}) |> Repo.update(),
          else: {:error, :season_closed}

      _ ->
        {:error, :not_found}
    end
  end

  defp latest_season?(id), do: id != nil and List.first(seasons())[:id] == id

  defp rank(entries, keys) do
    entries
    |> Enum.with_index(1)
    |> Enum.map_reduce({nil, 0}, fn {entry, index}, {previous, previous_rank} ->
      score = Map.take(entry, keys)
      rank = if score == previous, do: previous_rank, else: index
      {Map.put(entry, :rank, rank), {score, rank}}
    end)
    |> elem(0)
  end

  # Column names come exclusively from the static queries in this module.
  defp rows(sql, params \\ []) do
    %{columns: columns, rows: rows} = Repo.query!(sql, params)
    keys = Enum.map(columns, &String.to_atom/1)
    Enum.map(rows, &Map.new(Enum.zip(keys, &1)))
  end
end
