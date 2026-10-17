defmodule Pokerhans.PokerTest do
  use ExUnit.Case, async: false
  import Plug.Test
  alias Pokerhans.{Poker, Repo, Router, Release}

  setup do
    :ok = Ecto.Adapters.SQL.Sandbox.checkout(Repo)
    Repo.query!("INSERT INTO seasons (id, name) VALUES (3, 'Old'), (4, 'Current')")
    Repo.query!("INSERT INTO players (id, name) VALUES (1, 'Alice'), (2, 'Bob'), (3, '<script>')")
    :ok
  end

  test "scores use all wins, season-four cutoff, nullable places and competition ranks" do
    Repo.query!("""
    INSERT INTO games (season_id, host_id, winner_id, second_place_id, game_date) VALUES
    (3, 1, 3, 2, '2020-01-01'), (4, 1, 1, 2, '2024-01-01'),
    (4, 2, 2, 1, '2024-02-01'), (4, 3, NULL, NULL, '2024-03-01')
    """)

    assert Enum.map(Poker.overall_wins(), &{&1.player_id, &1.wins, &1.rank}) == [
             {3, 1, 1},
             {1, 1, 1},
             {2, 1, 1}
           ]

    assert Enum.map(Poker.points(), &{&1.player_id, &1.points, &1.rank}) == [{1, 4, 1}, {2, 4, 1}]
    assert List.last(Poker.games(4)).winner_name == ""
  end

  test "validates dates and references, accepts missing places, and protects historical seasons" do
    attrs = %{
      "season_id" => "4",
      "host_id" => "1",
      "game_date" => "2026-10-17",
      "winner_id" => ""
    }

    assert {:ok, game} = Poker.add_game(attrs)
    assert game.winner_id == nil
    assert {:error, _} = Poker.add_game(Map.put(attrs, "host_id", "999"))
    assert {:error, _} = Poker.add_game(Map.put(attrs, "game_date", "2026-02-30"))
    assert {:error, :season_closed} = Poker.add_game(Map.put(attrs, "season_id", "3"))
    assert {:error, :not_found} = Poker.update_date(game.id, 3, "2026-10-18")
    assert {:error, _} = Poker.update_date(game.id, 4, "")
    assert {:ok, changed} = Poker.update_date(game.id, 4, "2026-10-18")
    assert changed.game_date == ~D[2026-10-18]
  end

  test "renders existing pages, escapes names, and rejects missing CSRF tokens" do
    Repo.query!(
      "UPDATE players SET created_at = CASE id WHEN 1 THEN TIMESTAMP '2024-10-18' WHEN 2 THEN TIMESTAMP '2025-01-03' ELSE TIMESTAMP '2024-12-13' END"
    )

    assert call(:get, "/").status == 200
    assert call(:get, "/seasons").resp_headers |> Enum.member?({"location", "/season/4"})
    page = call(:get, "/season/4")
    assert page.status == 200
    assert page.resp_body =~ "&lt;script&gt;"
    assert page.resp_body =~ "name=\"_csrf_token\""

    assert [_, alice, script, bob] =
             Regex.run(
               ~r/(Add Game at Alice.*?)(Add Game at &lt;script&gt;.*?)(Add Game at Bob)/s,
               page.resp_body
             )

    assert alice != "" and script != "" and bob != ""
    assert call(:get, "/season/999").status == 404
    assert call(:get, "/season/4garbage").status == 404
    assert_raise Plug.CSRFProtection.InvalidCSRFTokenError, fn -> call(:post, "/game/add") end
    assert call(:get, "/game/add").status == 405
  end

  test "imports every field, advances sequences, and refuses to overwrite data" do
    Repo.query!("DELETE FROM players")
    Repo.query!("DELETE FROM seasons")

    path =
      Path.join(System.tmp_dir!(), "pokerhans-import-#{System.unique_integer([:positive])}.json")

    data = %{
      "seasons" => [%{"id" => 10, "name" => "Über uns", "created_at" => "2020-01-01T12:34:56"}],
      "players" => [%{"id" => 20, "name" => "A & B", "created_at" => nil}],
      "games" => [
        %{
          "id" => 30,
          "season_id" => 10,
          "host_id" => 20,
          "winner_id" => nil,
          "second_place_id" => nil,
          "game_date" => "2020-02-29",
          "created_at" => "2020-02-29T23:59:59"
        }
      ]
    }

    File.write!(path, Jason.encode!(data))
    on_exit(fn -> File.rm(path) end)
    assert Release.import_file(path) == %{"seasons" => 1, "players" => 1, "games" => 1}
    assert Release.verify_file(path) == %{"seasons" => 1, "players" => 1, "games" => 1}
    assert_raise RuntimeError, ~r/empty tables/, fn -> Release.import_file(path) end

    assert {:ok, game} =
             Poker.add_game(%{
               "season_id" => "10",
               "host_id" => "20",
               "game_date" => "2026-10-17"
             })

    assert game.id > 30
  end

  defp call(method, path), do: method |> conn(path) |> Router.call(Router.init([]))
end
