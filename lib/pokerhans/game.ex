defmodule Pokerhans.Game do
  @moduledoc "A hosted game with optional first and second places."
  use Ecto.Schema
  import Ecto.Changeset

  schema "games" do
    field(:season_id, :integer)
    field(:host_id, :integer)
    field(:winner_id, :integer)
    field(:second_place_id, :integer)
    field(:game_date, :date)
    field(:created_at, :naive_datetime)
  end

  @doc "Validates game input and maps database foreign key errors to field errors."
  def changeset(game, attrs) do
    game
    |> cast(attrs, [:season_id, :host_id, :winner_id, :second_place_id, :game_date])
    |> validate_required([:season_id, :host_id, :game_date])
    |> foreign_key_constraint(:season_id)
    |> foreign_key_constraint(:host_id)
    |> foreign_key_constraint(:winner_id)
    |> foreign_key_constraint(:second_place_id)
  end
end
