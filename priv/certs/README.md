# Hetzner PostgreSQL CA

`hetzner-postgresql-ca.pem` is the public database CA published at
https://docs.hetzner.com/managed/databases/postgresql/ (retrieved 2026-10-17).

Runtime connections to `*.your-database.de` use this CA with Postgrex certificate
and hostname verification. Other production hosts use system trust roots.
