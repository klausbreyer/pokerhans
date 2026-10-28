.PHONY: setup start build test css css-watch
setup:
	mix deps.get
	mix ecto.setup
	mix tailwind.install
	mix tailwind default
start:
	mix run --no-halt
build:
	MIX_ENV=prod mix tailwind default
	MIX_ENV=prod mix release
test:
	MIX_ENV=test mix ecto.create
	MIX_ENV=test mix ecto.migrate
	mix test
css:
	mix tailwind default
css-watch:
	mix tailwind default --watch
