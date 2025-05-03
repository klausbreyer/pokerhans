# Pokerhans

A simple web application for tracking poker games within a local group, organized by seasons. Each player hosts one game per season, and the app keeps track of hosts, winners, and runners-up.

## Features

- Season selection via dropdown
- Player overview showing visited and not-yet-visited players
- Easy game entry form for recording results
- Copy-paste friendly format for sharing status
- Game history view

## Tech Stack

- **Backend**: Go with standard library (net/http)
- **Database**: MySQL with golang-migrate
- **Frontend**: Server-rendered HTML with Tailwind CSS v4 (binary, config-free)
- **Configuration**: Environment variables via .env file

## Prerequisites

- Go 1.21+
- MySQL database
- curl (for downloading Tailwind binary and golang-migrate)

## Local Development

1. **Setup Database**

```sql
CREATE DATABASE pokerhans;
```

2. **Configure Environment Variables**

The application uses a `.env` file for configuration. Copy the example file to get started:

```bash
cp .env.example .env
```

Then edit the `.env` file with your database credentials:

```
# Database Configuration
DB_USER=your_db_user
DB_PASS=your_db_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=pokerhans
```

3. **Initialize the Project**

```bash
# Install Tailwind CSS binary, golang-migrate CLI, and set up directories
make init

# Build the CSS once
make css

# Run database migrations
make migrate-up
```

4. **Development Mode**

There are two main ways to run the application in development mode:

**Option 1: Using the dev script (recommended)**

```bash
./dev.sh
```

This script starts both the Tailwind CSS watcher and the Go server in one terminal.

**Option 2: Using separate terminals**

Terminal 1 (Tailwind CSS watcher):

```bash
make css-watch
```

Terminal 2 (Go server):

```bash
make run
# or
go run ./cmd/pokerhans
```

The application will be available at `http://localhost:8080`

## Database Migrations

The application uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations:

```bash
# Apply all migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Create a new migration
make migrate-create name=add_new_table

# Force a specific migration version
make migrate-force version=000001
```

Migration files are stored in the `migrations/mysql` directory as SQL files.

## Adding Seasons and Players

Seasons and players are added directly to the database. Use the following SQL:

```sql
-- Add a new season
INSERT INTO seasons (name) VALUES ('Season 2023');

-- Add a new player
INSERT INTO players (name) VALUES ('John Doe');

-- Associate a player with a season
INSERT INTO season_players (season_id, player_id)
VALUES (1, 1);
```

## Make Commands

- `make init` - Initialize the development environment
- `make build` - Build the Go application
- `make run` - Run the Go server
- `make test` - Run tests
- `make css` - Build the CSS once
- `make css-watch` - Watch CSS files for changes
- `make tailwind-install` - Download and install Tailwind CSS binary
- `make migrate-install` - Download and install golang-migrate CLI
- `make migrate-up` - Apply database migrations
- `make migrate-down` - Rollback database migrations
- `make migrate-create name=xyz` - Create a new migration
- `make migrate-force version=N` - Force a specific migration version
- `make dev` - Print instructions for development mode
- `make all` - Build CSS, build Go application, and run
