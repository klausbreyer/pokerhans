defmodule Pokerhans.Repo.Migrations.CreateImportReceipts do
  use Ecto.Migration

  def change do
    create table(:import_receipts, primary_key: false) do
      add(:checksum, :string, primary_key: true)
      add(:counts, :map, null: false)
      add(:imported_at, :naive_datetime, null: false, default: fragment("CURRENT_TIMESTAMP"))
    end
  end
end
