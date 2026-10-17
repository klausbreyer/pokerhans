FROM elixir:1.19.5-otp-28-slim AS build
RUN apt-get update && apt-get install -y --no-install-recommends build-essential git ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
ENV MIX_ENV=prod
RUN mix local.hex --force && mix local.rebar --force
COPY mix.exs mix.lock ./
COPY config config
RUN mix deps.get --only prod && mix deps.compile
COPY lib lib
COPY priv priv
COPY web/static/css web/static/css
RUN mix compile --warnings-as-errors && mix tailwind.install && mix tailwind default && mix release

# Keep Erlang's native library ABI compatible with the builder.
FROM elixir:1.19.5-otp-28-slim AS runtime
WORKDIR /app
ENV MIX_ENV=prod PORT=4000 ERL_FLAGS="+S 2:2"
COPY --from=build --chown=nobody:nogroup /app/_build/prod/rel/pokerhans ./
COPY --chmod=755 scripts/start.sh /app/start.sh
USER nobody
EXPOSE 4000
CMD ["/app/start.sh"]
