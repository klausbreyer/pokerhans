defmodule Pokerhans.Repo.Migrations.CreatePoker do
  use Ecto.Migration

  def change do
    for table_name <- [:seasons, :players] do
      create table(table_name) do
        add(:name, :string, null: false)
        add(:created_at, :naive_datetime, default: fragment("CURRENT_TIMESTAMP"))
      end
    end

    create table(:games) do
      add(:season_id, references(:seasons), null: false)
      add(:host_id, references(:players), null: false)
      add(:winner_id, references(:players))
      add(:second_place_id, references(:players))
      add(:game_date, :date, null: false)
      add(:created_at, :naive_datetime, default: fragment("CURRENT_TIMESTAMP"))
    end

    create(index(:games, [:season_id, :game_date]))
  end
end
