# Pokerhans

Elixir application using Plug, Bandit, Ecto and PostgreSQL, hosted on Bunny Magic Containers.

## Commands
- Setup: `make setup`
- Run: `make start`
- Tests: `make test`
- Formatting: `mix format --check-formatted`
- Production release: `make build`

## Conventions
- Keep the existing routes and Tailwind interface stable.
- Use Ecto changesets for form input and parameterized database queries.
- Document public modules and functions.
- Never import migration snapshots over existing data.
- Bunny runs one web pod in DE using external PostgreSQL. Database changes require a verified data migration before deployment.
- Until that migration is complete, retain the current Postgres sidecar and volume.
- Keep credentials and database exports outside Git.
