.PHONY: run build test css css-watch tailwind-install migrate-up migrate-down migrate-create

# Default target
all: css build run

# Build the application
build:
	go build -o bin/pokerhans ./cmd/pokerhans

# Run the application
run:
	go run ./cmd/pokerhans

# Run tests
test:
	go test ./...

# Install golang-migrate CLI (requires wget or curl)
migrate-install:
	@if [ ! -f "bin/migrate" ]; then \
		echo "Installing golang-migrate CLI..."; \
		mkdir -p bin; \
		if [ "$$(uname)" = "Darwin" ]; then \
			if [ "$$(uname -m)" = "arm64" ]; then \
				curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.0/migrate.darwin-arm64.tar.gz | tar xvz; \
				mv migrate bin/migrate; \
			else \
				curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.0/migrate.darwin-amd64.tar.gz | tar xvz; \
				mv migrate bin/migrate; \
			fi \
		elif [ "$$(uname)" = "Linux" ]; then \
			curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.0/migrate.linux-amd64.tar.gz | tar xvz; \
			mv migrate bin/migrate; \
		else \
			echo "Unsupported OS. Please download migrate CLI manually from https://github.com/golang-migrate/migrate/releases"; \
			exit 1; \
		fi \
	else \
		echo "migrate CLI already installed"; \
	fi

# Create a new migration (Usage: make migrate-create name=migration_name)
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Please provide a migration name. Example: make migrate-create name=add_users_table"; \
		exit 1; \
	fi
	./bin/migrate create -ext sql -dir migrations/mysql -seq $(name)

# Run migrations up using direct DSN to avoid module path issues
migrate-up:
	@. ./.env 2>/dev/null || true; \
	export DB_USER=$${DB_USER:-root}; \
	export DB_PASS=$${DB_PASS:-PASSPASS}; \
	export DB_HOST=$${DB_HOST:-localhost}; \
	export DB_PORT=$${DB_PORT:-3306}; \
	export DB_NAME=$${DB_NAME:-pokerhans}; \
	./bin/migrate -path migrations/mysql -database "mysql://$${DB_USER}:$${DB_PASS}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}?parseTime=true" up

# Run migrations down
migrate-down:
	@. ./.env 2>/dev/null || true; \
	export DB_USER=$${DB_USER:-root}; \
	export DB_PASS=$${DB_PASS:-PASSPASS}; \
	export DB_HOST=$${DB_HOST:-localhost}; \
	export DB_PORT=$${DB_PORT:-3306}; \
	export DB_NAME=$${DB_NAME:-pokerhans}; \
	./bin/migrate -path migrations/mysql -database "mysql://$${DB_USER}:$${DB_PASS}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}?parseTime=true" down

# Force migration version (Usage: make migrate-force version=000001)
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Please provide a version. Example: make migrate-force version=000001"; \
		exit 1; \
	fi; \
	. ./.env 2>/dev/null || true; \
	export DB_USER=$${DB_USER:-root}; \
	export DB_PASS=$${DB_PASS:-PASSPASS}; \
	export DB_HOST=$${DB_HOST:-localhost}; \
	export DB_PORT=$${DB_PORT:-3306}; \
	export DB_NAME=$${DB_NAME:-pokerhans}; \
	./bin/migrate -path migrations/mysql -database "mysql://$${DB_USER}:$${DB_PASS}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}?parseTime=true" force $(version)

# Install Tailwind CSS binary
tailwind-install:
	@if [ "$$(uname)" = "Darwin" ]; then \
		echo "Installing Tailwind CSS for macOS..."; \
		curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64; \
		chmod +x tailwindcss-macos-x64; \
		mv tailwindcss-macos-x64 bin/tailwindcss; \
	elif [ "$$(uname)" = "Linux" ]; then \
		echo "Installing Tailwind CSS for Linux..."; \
		curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64; \
		chmod +x tailwindcss-linux-x64; \
		mv tailwindcss-linux-x64 bin/tailwindcss; \
	else \
		echo "Sorry, automatic installation is only supported on macOS and Linux."; \
		echo "Please download Tailwind CSS manually from https://github.com/tailwindlabs/tailwindcss/releases"; \
		exit 1; \
	fi

# Build CSS once
css:
	./bin/tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# Watch CSS files for changes
css-watch:
	./bin/tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch

# Initialize development environment
init: 
	@echo "Creating bin directory..."
	mkdir -p bin
	@echo "Installing Tailwind CSS binary..."
	$(MAKE) tailwind-install
	@echo "Installing golang-migrate CLI..."
	$(MAKE) migrate-install
	@echo "In Tailwind CSS v4 werden keine separaten Konfigurationsdateien mehr benötigt - alle Konfigurationen sind in input.css"

# Dev mode: Run CSS watch and Go server in parallel (requires tmux or multiple terminals)
dev:
	@echo "Start two terminals:"
	@echo "Terminal 1: make css-watch"
	@echo "Terminal 2: make run"