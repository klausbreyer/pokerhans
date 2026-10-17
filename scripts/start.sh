#!/bin/sh
set -eu

# The database sidecar can take longer to start than the application container.
attempt=0
until /app/bin/pokerhans eval 'Pokerhans.Release.migrate()'; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo 'Database migrations failed after 30 attempts' >&2
    exit 1
  fi
  sleep 2
done
/app/bin/pokerhans eval 'Pokerhans.Release.bootstrap()'
exec /app/bin/pokerhans start
