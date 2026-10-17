# Pokerhans

Poker seasons, hosting rotation and leaderboards. Elixir, Plug/Bandit, Ecto,
PostgreSQL and server-rendered EEx templates with Tailwind CSS.

## Local development

Requires Elixir 1.19, Erlang/OTP 28 and an existing PostgreSQL server.

```sh
make setup
make start
```

Open http://localhost:4000. Development uses `pokerhans_dev`, tests use
`pokerhans_test`. Export `DATABASE_URL` to override the database connection.
Environment files are not loaded automatically. `make test` creates/migrates
the test database and runs the tests. `mix format --check-formatted` checks formatting.

Seasons and players are managed directly in the database, as before. The latest
season accepts new games and date changes. Winner and second place are optional.
The home page shows all-time wins and 3/1 scoring from season four onward.

## Bunny deployment

GitHub Actions tests and publishes `ghcr.io/klausbreyer/pokerhans:<commit>`.
Pushes to `main` deploy that immutable image to the app configured in `bunny.json`.
The required GitHub secret is `BUNNYNET_API_KEY`:

```sh
gh secret set BUNNYNET_API_KEY --repo klausbreyer/pokerhans
```

PostgreSQL runs as a sidecar on localhost with a persistent volume. Keep exactly
one pod in Germany. Extra pods get separate databases, not replicas. Deployments
check the volume instance ID before and after replacing the web container.
A volume handoff causes brief downtime. Never delete/recreate or scale the app
as a deployment procedure.

Runtime secrets `DATABASE_URL`, `POSTGRES_PASSWORD` and `SECRET_KEY_BASE` live in
Bunny and are preserved by deployments. The database is not publicly exposed.
The startup script waits for PostgreSQL, applies migrations and starts the release.
The CDN must disable caching for dynamic pages because forms contain CSRF tokens.

```sh
python3 scripts/deploy.py                    # read-only health/storage checks
python3 scripts/deploy.py --tag <commit>     # deploy a published image
```

## MySQL migration

The exporter uses the Elixir MyXQL client. No MySQL server or CLI installation
is required. Export `MYSQL_HOST`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`
and optionally `MYSQL_PORT`, then:

```sh
mix run --no-start scripts/export_mysql.exs /private/path/source.json
mix run --no-start -e 'Pokerhans.Release.import_file("/private/path/source.json")'
mix run --no-start -e 'Pokerhans.Release.verify_file("/private/path/source.json")'
```

The exporter reads a consistent, read-only transaction with UTC timestamps.
The importer requires empty destination tables, retains every ID and nullable
placement, advances sequences, and compares every source value within the same
transaction. Invalid references or mismatches roll back the import. A receipt
records the source SHA-256 and table counts.

For initial container provisioning, `MYSQL_IMPORT_BASE64` can contain the export.
Startup imports it once and records its checksum atomically. Remove this temporary
environment variable after verification. Existing data is never overwritten.

Stop the old writer before the final export and change DNS only after the new
application and its imported data have been verified. Retain the source database
and a private export for rollback. After new writes reach Postgres, switching back
requires reconciling those writes first. Bunny volumes do not provide automatic
backups; retain external PostgreSQL backups before relying on this as sole storage.
