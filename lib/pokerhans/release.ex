defmodule Pokerhans.Release do
  @moduledoc "Release database migrations and verified legacy data import."
  alias Pokerhans.Repo

  @tables [
    {"seasons", ~w(id name created_at)},
    {"players", ~w(id name created_at)},
    {"games", ~w(id season_id host_id winner_id second_place_id game_date created_at)}
  ]

  @doc "Applies pending migrations without starting the HTTP server."
  def migrate do
    Application.load(:pokerhans)
    {:ok, _, _} = Ecto.Migrator.with_repo(Repo, &Ecto.Migrator.run(&1, :up, all: true))
  end

  @doc "Imports a JSON snapshot into empty tables, atomically, then verifies every value."
  def import_file(path) do
    contents = File.read!(path)
    data = Jason.decode!(contents)
    checksum = :crypto.hash(:sha256, contents) |> Base.encode16(case: :lower)

    {:ok, result, _} =
      Ecto.Migrator.with_repo(Repo, fn _ ->
        Repo.transaction(fn ->
          Repo.query!("LOCK TABLE seasons, players, games IN ACCESS EXCLUSIVE MODE")

          for {table, columns} <- @tables do
            %{rows: [[count]]} = Repo.query!("SELECT COUNT(*) FROM #{table}")
            if count != 0, do: raise("Import requires empty tables: #{table}")

            rows =
              Enum.map(Map.fetch!(data, table), fn row ->
                Map.new(columns, fn key ->
                  {String.to_atom(key), decode(key, Map.fetch!(row, key))}
                end)
              end)

            Repo.insert_all(table, rows)
            # Explicit imported IDs must not collide with subsequent inserts.
            Repo.query!(
              "SELECT setval(pg_get_serial_sequence('#{table}', 'id'), COALESCE(MAX(id), 1), COUNT(*) > 0) FROM #{table}"
            )
          end

          counts = verify_data!(data)
          Repo.insert_all("import_receipts", [%{checksum: checksum, counts: counts}])
          counts
        end)
      end)

    case result do
      {:ok, counts} -> counts
      {:error, reason} -> raise "Import rolled back: #{inspect(reason)}"
    end
  end

  @doc "Imports the one-time migration snapshot, once per checksum, before serving requests."
  def bootstrap do
    if encoded = System.get_env("MYSQL_IMPORT_BASE64") do
      contents = Base.decode64!(encoded)
      checksum = :crypto.hash(:sha256, contents) |> Base.encode16(case: :lower)

      {:ok, exists, _} =
        Ecto.Migrator.with_repo(Repo, fn _ ->
          Repo.query!("SELECT checksum FROM import_receipts WHERE checksum = $1", [checksum]).num_rows ==
            1
        end)

      unless exists do
        path =
          Path.join(
            System.tmp_dir!(),
            "pokerhans-import-#{System.unique_integer([:positive])}.json"
          )

        try do
          File.write!(path, contents, [:exclusive])
          File.chmod!(path, 0o600)
          IO.inspect(import_file(path), label: "Verified MySQL import")
        after
          File.rm(path)
        end
      end
    end
  end

  @doc "Compares every imported column with a source snapshot."
  def verify_file(path) do
    data = path |> File.read!() |> Jason.decode!()
    {:ok, counts, _} = Ecto.Migrator.with_repo(Repo, fn _ -> verify_data!(data) end)
    counts
  end

  defp verify_data!(data) do
    Map.new(@tables, fn {table, columns} ->
      %{rows: actual} = Repo.query!("SELECT #{Enum.join(columns, ",")} FROM #{table} ORDER BY id")

      expected =
        data
        |> Map.fetch!(table)
        |> Enum.sort_by(& &1["id"])
        |> Enum.map(fn row -> Enum.map(columns, &decode(&1, Map.fetch!(row, &1))) end)

      if actual != expected, do: raise("Data verification failed: #{table}")
      {table, length(actual)}
    end)
  end

  defp decode(_, nil), do: nil
  defp decode("game_date", value), do: Date.from_iso8601!(value)
  defp decode("created_at", value), do: NaiveDateTime.from_iso8601!(value)
  defp decode(_, value), do: value
end
