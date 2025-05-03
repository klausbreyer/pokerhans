package main

import (
	"fmt"
	"log"
	"os"

	"github.com/klausbreyer/pokerhans/internal/config"
	"github.com/klausbreyer/pokerhans/internal/db"
)

func main() {
	// Set up logger
	logger := log.New(os.Stdout, "seeddata: ", log.LstdFlags)

	// Load environment variables from .env file
	envPath := "../.env"
	if err := config.LoadEnv(envPath); err != nil {
		logger.Printf("Warning: Unable to load .env file: %v", err)
	}

	// Connect to database
	database, err := db.Connect()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Add a sample season
	_, err = database.Exec("INSERT INTO seasons (name) VALUES ('Summer 2025')")
	if err != nil {
		logger.Fatalf("Failed to add season: %v", err)
	}

	// Get the season ID
	var seasonID int
	err = database.QueryRow("SELECT id FROM seasons WHERE name = 'Summer 2025'").Scan(&seasonID)
	if err != nil {
		logger.Fatalf("Failed to get season ID: %v", err)
	}

	// Add some players
	playerNames := []string{"Alice", "Bob", "Charlie", "David", "Eva"}
	playerIDs := make([]int, len(playerNames))

	for i, name := range playerNames {
		// Add player
		result, err := database.Exec("INSERT INTO players (name) VALUES (?)", name)
		if err != nil {
			logger.Fatalf("Failed to add player %s: %v", name, err)
		}

		// Get player ID
		playerID, err := result.LastInsertId()
		if err != nil {
			logger.Fatalf("Failed to get player ID for %s: %v", name, err)
		}
		playerIDs[i] = int(playerID)

		// Add player to season
		_, err = database.Exec("INSERT INTO season_players (season_id, player_id) VALUES (?, ?)", seasonID, playerID)
		if err != nil {
			logger.Fatalf("Failed to add player %s to season: %v", name, err)
		}
	}

	fmt.Printf("Successfully added season 'Summer 2025' with %d players\n", len(playerNames))
}