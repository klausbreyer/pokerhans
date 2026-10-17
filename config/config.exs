import Config

config :pokerhans, ecto_repos: [Pokerhans.Repo]
config :logger, level: :info
config :pokerhans, :server, config_env() != :test

config :pokerhans, Pokerhans.Repo,
  database: "pokerhans_#{config_env()}",
  hostname: "localhost",
  pool_size: 5

config :tailwind,
  version: "4.1.18",
  default: [
    args: ~w(--input=web/static/css/input.css --output=priv/static/css/output.css --minify),
    cd: Path.expand("..", __DIR__)
  ]

if config_env() == :test do
  config :pokerhans, Pokerhans.Repo, pool: Ecto.Adapters.SQL.Sandbox
end
