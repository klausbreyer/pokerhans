defmodule Pokerhans.Application do
  @moduledoc "Supervises the database pool and HTTP server."
  use Application

  @impl true
  def start(_type, _args) do
    server =
      if Application.fetch_env!(:pokerhans, :server),
        do: [
          {Bandit,
           plug: Pokerhans.Router,
           port: Application.fetch_env!(:pokerhans, :port),
           ip: {0, 0, 0, 0}}
        ],
        else: []

    Supervisor.start_link([Pokerhans.Repo] ++ server,
      strategy: :one_for_one,
      name: Pokerhans.Supervisor
    )
  end
end
