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
Required GitHub secrets are `BUNNYNET_API_KEY` and `DATABASE_URL`:

```sh
gh secret set BUNNYNET_API_KEY --repo klausbreyer/pokerhans
gh secret set DATABASE_URL --repo klausbreyer/pokerhans
```

Paste the database URL at the prompt, for example
`postgresql://USER:PASSWORD@HOST:5432/pokerhans?ssl=true`.
Percent-encode special characters in credentials. `ssl=true` enables TLS with
certificate and hostname verification. Production enables TLS by default.
For `*.your-database.de`, the bundled [Hetzner CA](priv/certs/README.md) is used.
Use a dedicated database for Pokerhans.

The deployment runs one web container in Germany, connected to external
PostgreSQL. GitHub supplies `DATABASE_URL`; other runtime secrets, including
`SECRET_KEY_BASE`, are preserved on Bunny. The startup script applies migrations
before starting the release. Disable CDN caching for pages containing CSRF tokens.

The container defaults to `RELEASE_MODE=interactive`, which loads compiled modules
on first use instead of preloading every module at boot. This reduces startup
memory use; the first call to a module pays its loading cost. Applications still
start normally, and the migration and bootstrap `eval` commands are unaffected.
Set `RELEASE_MODE=embedded` in the container environment to restore preloading.
See the [Mix release documentation](https://hexdocs.pm/mix/1.19.5/Mix.Tasks.Release.html#module-environment-variables).

Compare memory after visiting the home page, seasons and individual season pages,
once the rolling update has settled to one ready pod. Read
`regions[].pods[].ramUsageBytes` from `bunny api GET /mc/apps/066vxj3nesUoaZf/overview`.
An immediate post-restart reading does not reflect warmed memory use.

Database moves require a separate migration: freeze writes, export current data,
restore and verify it, then update Bunny and the GitHub secret together.
Deployments reject a sidecar or a database URL that differs from Bunny's verified
target. They never switch databases implicitly.

```sh
python3 scripts/deploy.py                    # read-only health checks
python3 scripts/deploy.py --tag <commit>     # requires DATABASE_URL in the environment
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

## Migration record

On 2026-10-17, all 5 seasons, 9 players and 27 games were migrated from the
production MySQL database to PostgreSQL. The final export was taken after
stopping the Fly writer and matched the imported UTC snapshot exactly.
SHA-256: `8305f98c68e134887d5ff02bf6b04273915c0c38df77461b426af7a6d2415b2b`.

- Production: https://pokerhans.breyer.berlin
- Bunny app: `066vxj3nesUoaZf`, endpoint `mc-j5ztc5ffb8.bunny.run`
- Private backups: `~/Backups/pokerhans/20261017T151325Z/`
- Cloudflare CNAME `pokerhans` points to `mc-j5ztc5ffb8.b-cdn.net`, DNS only.

A replacement pod retained the same volume and all pages matched the original
application after the restart. The temporary import environment was removed.
The old Fly machine `17814473a47598` is stopped with autostart and restart
disabled. The legacy Fly deployment and MySQL migration workflows are disabled.
The original MySQL database remains intact. Its original Fly configuration and
source snapshot are in the private backup directory. New PostgreSQL writes must
be preserved before any rollback.

### External PostgreSQL cutover

Later on 2026-10-17, the current Bunny PostgreSQL data was transferred to the
user-provided `pokerhans` database on `mkqq.your-database.de`. The app was placed
in maintenance mode during the snapshot. Every value in all five tables,
including migration metadata, and all three sequence states were verified.
The source snapshot SHA-256 is
`cfa0ebb0d0f6d9a7deddc7e82a984191d9a9c7005101eed16ae9b17db13dcc91`.

The application now uses certificate- and hostname-verified TLS. All six pages
matched the pre-cutover site, and a CSRF-protected date update succeeded without
changing stored values. The temporary export endpoint, PostgreSQL container and
Bunny volume were removed.

The snapshot and a PostgreSQL custom-format dump are stored privately at
`~/Backups/pokerhans/20261017T153827Z-external/`. The dump was restored to an
isolated local database and every table was compared with the source snapshot.
