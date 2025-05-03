#!/bin/bash

# Set up trap to kill all child processes when script exits
trap 'kill $(jobs -p) 2>/dev/null' EXIT

# Check if Tailwind binary exists
if [ ! -f "./bin/tailwindcss" ]; then
  echo "Tailwind CSS binary not found. Installing..."
  make tailwind-install
fi

# Start Tailwind CSS watching for changes
echo "Starting Tailwind CSS watcher..."
./bin/tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch &

# Wait a bit for Tailwind to start
sleep 2

# Start the Go server
echo "Starting Go server..."
go run ./cmd/pokerhans

# Wait for all background processes to finish (which won't happen normally)
wait