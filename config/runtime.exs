import Config

if url = System.get_env("DATABASE_URL") do
  config :pokerhans, Pokerhans.Repo, url: url
end

config :pokerhans,
  port: String.to_integer(System.get_env("PORT", "4000")),
  secure_cookie: config_env() == :prod,
  secret_key_base:
    System.get_env("SECRET_KEY_BASE") ||
      if(config_env() == :prod,
        do: raise("SECRET_KEY_BASE is required"),
        else: String.duplicate("development-only", 5)
      )

if config_env() == :prod and !System.get_env("DATABASE_URL") do
  raise "DATABASE_URL is required"
end
