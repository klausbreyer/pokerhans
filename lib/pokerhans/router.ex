defmodule Pokerhans.Router do
  @moduledoc "HTTP routes, signed sessions and CSRF protection for game forms."
  use Plug.Router
  alias Pokerhans.{Poker, Repo, View}

  plug(Plug.RequestId)
  plug(Plug.Logger)
  plug(Plug.Static, at: "/static", from: {:pokerhans, "priv/static"})
  plug(:session)
  plug(:fetch_session)
  plug(Plug.Parsers, parsers: [:urlencoded], pass: [], length: 64_000)
  plug(Plug.CSRFProtection)
  plug(:match)
  plug(:dispatch)

  get "/health" do
    case Repo.query("SELECT 1", [], timeout: 2_000) do
      {:ok, _} -> send_resp(conn, 200, "ok")
      _ -> send_resp(conn, 503, "database unavailable")
    end
  end

  get "/" do
    page(conn, %{
      page: "leaderboard",
      overall_wins: Poker.overall_wins(),
      points_leaderboard: Poker.points(),
      leaderboard_start_season: 4
    })
  end

  get "/leaderboard" do
    redirect(conn, "/", 301)
  end

  get "/seasons" do
    case Poker.seasons() do
      [season | _] -> redirect(conn, "/season/#{season.id}")
      [] -> page(conn, %{page: "home", seasons: []})
    end
  end

  get "/season/:id" do
    seasons = Poker.seasons()
    season = Enum.find(seasons, &(&1.id == integer(id)))

    if season do
      games = Poker.games(season.id)
      players = Poker.players()
      hosted = MapSet.new(games, & &1.host_id)

      to_visit =
        players
        |> Enum.reject(&MapSet.member?(hosted, &1.id))
        |> Enum.sort_by(
          &{if(&1.created_at, do: NaiveDateTime.to_iso8601(&1.created_at), else: ""), &1.id}
        )

      visited = Enum.map(games, fn game -> %{name: game.host_name, game_date: game.game_date} end)

      page(conn, %{
        page: "season",
        seasons: seasons,
        current_season: season,
        games: games,
        all_players: players,
        to_visit_players: to_visit,
        visited_players: visited,
        is_latest_season: hd(seasons).id == season.id,
        current_date: Date.to_iso8601(Date.utc_today())
      })
    else
      send_resp(conn, 404, "Season not found")
    end
  end

  post "/game/add" do
    case Poker.add_game(conn.body_params) do
      {:ok, game} -> redirect(conn, "/season/#{game.season_id}")
      {:error, _} -> send_resp(conn, 422, "Invalid game or closed season")
    end
  end

  post "/game/update-date" do
    with game_id when is_integer(game_id) <- integer(conn.body_params["game_id"]),
         season_id when is_integer(season_id) <- integer(conn.body_params["season_id"]),
         {:ok, _} <- Poker.update_date(game_id, season_id, conn.body_params["new_date"]) do
      redirect(conn, "/season/#{season_id}")
    else
      _ -> send_resp(conn, 422, "Invalid game date or closed season")
    end
  end

  match _ do
    if conn.request_path in ["/", "/seasons", "/leaderboard", "/game/add", "/game/update-date"],
      do: send_resp(conn, 405, "Method not allowed"),
      else: send_resp(conn, 404, "Not found")
  end

  defp session(conn, _) do
    conn = %{conn | secret_key_base: Application.fetch_env!(:pokerhans, :secret_key_base)}

    opts =
      Plug.Session.init(
        store: :cookie,
        key: "_pokerhans",
        signing_salt: "pokerhans-session",
        same_site: "Lax",
        secure: Application.fetch_env!(:pokerhans, :secure_cookie),
        http_only: true
      )

    Plug.Session.call(conn, opts)
  end

  defp page(conn, data) do
    body = View.render("layout", Map.put(data, :current_year, Date.utc_today().year))

    conn
    |> put_resp_content_type("text/html")
    |> put_resp_header("cache-control", "no-store")
    |> put_resp_header("x-content-type-options", "nosniff")
    |> send_resp(200, Phoenix.HTML.safe_to_string(body))
  end

  defp redirect(conn, path, status \\ 303),
    do: conn |> put_resp_header("location", path) |> send_resp(status, "")

  defp integer(value) when is_binary(value) do
    case Integer.parse(value) do
      {id, ""} when id > 0 -> id
      _ -> nil
    end
  end

  defp integer(_), do: nil
end
