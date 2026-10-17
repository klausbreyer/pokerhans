defmodule Pokerhans.Repo do
  @moduledoc "PostgreSQL repository."
  use Ecto.Repo, otp_app: :pokerhans, adapter: Ecto.Adapters.Postgres
end
