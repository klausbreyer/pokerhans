defmodule Pokerhans.MixProject do
  use Mix.Project

  def project do
    [
      app: :pokerhans,
      version: "1.0.0",
      elixir: "~> 1.19",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      aliases: ["ecto.setup": ["ecto.create", "ecto.migrate"]]
    ]
  end

  def application do
    [extra_applications: [:logger, :crypto], mod: {Pokerhans.Application, []}]
  end

  defp deps do
    [
      {:bandit, "~> 1.0"},
      {:plug, "~> 1.18"},
      {:ecto_sql, "~> 3.13"},
      {:postgrex, "~> 0.21"},
      {:phoenix_html, "~> 4.2"},
      {:jason, "~> 1.4"},
      {:myxql, "~> 0.8", only: :dev},
      {:tailwind, "~> 0.4", runtime: Mix.env() == :dev}
    ]
  end
end
