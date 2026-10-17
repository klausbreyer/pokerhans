# Run with: mix run --no-start scripts/export_mysql.exs /private/path/source.json
# Source credentials are supplied as MYSQL_HOST, MYSQL_USER, MYSQL_PASSWORD,
# MYSQL_DATABASE and optional MYSQL_PORT. The source transaction is read-only.
Application.ensure_all_started(:myxql)
[path] = System.argv()
{:ok, conn} = MyXQL.start_link(
  hostname: System.fetch_env!("MYSQL_HOST"),
  port: String.to_integer(System.get_env("MYSQL_PORT", "3306")),
  username: System.fetch_env!("MYSQL_USER"),
  password: System.fetch_env!("MYSQL_PASSWORD"),
  database: System.fetch_env!("MYSQL_DATABASE"),
  pool_size: 1)
MyXQL.query!(conn, "SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ")
MyXQL.query!(conn, "SET SESSION TRANSACTION READ ONLY")
MyXQL.query!(conn, "SET time_zone = '+00:00'")
{:ok, data} = MyXQL.transaction(conn, fn tx ->
  Map.new(~w(seasons players games), fn table ->
    %{columns: columns, rows: rows} = MyXQL.query!(tx, "SELECT * FROM #{table} ORDER BY id")
    {table, Enum.map(rows, &Map.new(Enum.zip(columns, &1)))}
  end)
end)
{:ok, file} = File.open(path, [:write, :exclusive])
File.chmod!(path, 0o600)
IO.binwrite(file, Jason.encode!(data, pretty: true))
File.close(file)
IO.inspect(Map.new(data, fn {table, rows} -> {table, length(rows)} end), label: "Exported")
