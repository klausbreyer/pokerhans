package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/klausbreyer/pokerhans/internal/config"
	"github.com/klausbreyer/pokerhans/internal/db"
)

func main() {
	// Set up logger
	logger := log.New(os.Stdout, "demogen: ", log.LstdFlags)

	// Load environment variables from .env file
	envPath := ".env"
	if err := config.LoadEnv(envPath); err != nil {
		logger.Printf("Warning: Unable to load .env file: %v", err)
	}

	// Connect to database
	database, err := db.Connect()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create seasons
	seasons := []string{
		"Winter 2024",
		"Spring 2025",
		"Summer 2025",
		"Fall 2025",
	}

	seasonIDs := make(map[string]int)
	for _, seasonName := range seasons {
		// Add season
		_, err = database.Exec("INSERT INTO seasons (name) VALUES (?)", seasonName)
		if err != nil {
			logger.Fatalf("Failed to add season %s: %v", seasonName, err)
		}

		// Get the season ID
		var seasonID int
		err = database.QueryRow("SELECT id FROM seasons WHERE name = ?", seasonName).Scan(&seasonID)
		if err != nil {
			logger.Fatalf("Failed to get season ID for %s: %v", seasonName, err)
		}
		seasonIDs[seasonName] = seasonID
		fmt.Printf("Created season: %s (ID: %d)\n", seasonName, seasonID)
	}

	// Create players with more German poker-themed names
	playerNames := []string{
		"Max Mustermann", "Lisa Schmidt", "Jonas Weber", "Anna Müller", "Felix König",
		"Sophie Becker", "Lukas Hoffmann", "Emma Fischer", "Paul Wagner", "Laura Schneider",
		"Tim Meyer", "Julia Schulz", "Nico Bauer", "Lena Schäfer", "David Klein",
		"Marie Richter", "Fabian Wolf", "Nina Braun", "Philipp Zimmermann", "Katja Schwarz",
	}

	playerIDs := make(map[string]int)
	for _, name := range playerNames {
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
		playerIDs[name] = int(playerID)
		fmt.Printf("Created player: %s (ID: %d)\n", name, playerID)
	}

	// Assign players to seasons (not all players in all seasons)
	for seasonName, seasonID := range seasonIDs {
		// Select a random subset of players for each season (between 10-15 players)
		numPlayers := rand.Intn(6) + 10 // 10-15 players
		
		// Create a copy of player names to shuffle
		shuffledPlayers := make([]string, len(playerNames))
		copy(shuffledPlayers, playerNames)
		rand.Shuffle(len(shuffledPlayers), func(i, j int) {
			shuffledPlayers[i], shuffledPlayers[j] = shuffledPlayers[j], shuffledPlayers[i]
		})
		
		// Take the first numPlayers players
		seasonPlayers := shuffledPlayers[:numPlayers]
		
		for _, playerName := range seasonPlayers {
			playerID := playerIDs[playerName]
			// Add player to season
			_, err = database.Exec("INSERT INTO season_players (season_id, player_id) VALUES (?, ?)", seasonID, playerID)
			if err != nil {
				logger.Printf("Warning: Failed to add player %s to season %s: %v", playerName, seasonName, err)
				continue
			}
			fmt.Printf("Added player %s to season %s\n", playerName, seasonName)
		}
		
		// Create games for this season
		createGamesForSeason(database, logger, seasonID, seasonName, seasonPlayers, playerIDs)
	}

	fmt.Println("Successfully generated demo data!")
}

func createGamesForSeason(database *sql.DB, logger *log.Logger, seasonID int, seasonName string, seasonPlayers []string, playerIDs map[string]int) {
	// Determine the start date for the season
	var startDate time.Time
	switch seasonName {
	case "Winter 2024":
		startDate = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	case "Spring 2025":
		startDate = time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	case "Summer 2025":
		startDate = time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC)
	case "Fall 2025":
		startDate = time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC)
	default:
		startDate = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	// Create between 10-20 games for the season
	numGames := rand.Intn(11) + 10 // 10-20 games
	
	for i := 0; i < numGames; i++ {
		// Select random date for the game (within a 3-month period from start date)
		gameDate := startDate.Add(time.Duration(rand.Intn(90)) * 24 * time.Hour)
		
		// Select random host, winner, and second place
		rand.Shuffle(len(seasonPlayers), func(i, j int) {
			seasonPlayers[i], seasonPlayers[j] = seasonPlayers[j], seasonPlayers[i]
		})
		
		hostName := seasonPlayers[0]
		hostID := playerIDs[hostName]
		
		// Some games might not have winners recorded yet (about 20% chance)
		var winnerID, secondPlaceID *int
		if rand.Float32() > 0.2 {
			// Ensure winner and second place are different from host and each other
			winnerName := seasonPlayers[1]
			secondPlaceName := seasonPlayers[2]
			
			winnerIDValue := playerIDs[winnerName]
			secondPlaceIDValue := playerIDs[secondPlaceName]
			
			winnerID = &winnerIDValue
			secondPlaceID = &secondPlaceIDValue
		}
		
		// Add the game
		_, err := database.Exec(
			"INSERT INTO games (season_id, host_id, winner_id, second_place_id, game_date) VALUES (?, ?, ?, ?, ?)",
			seasonID, hostID, winnerID, secondPlaceID, gameDate.Format("2006-01-02"),
		)
		if err != nil {
			logger.Printf("Warning: Failed to create game for %s on %s: %v", 
				hostName, gameDate.Format("2006-01-02"), err)
			continue
		}
		
		winnerStr := "unknown"
		secondStr := "unknown"
		if winnerID != nil {
			for name, id := range playerIDs {
				if id == *winnerID {
					winnerStr = name
				}
				if secondPlaceID != nil && id == *secondPlaceID {
					secondStr = name
				}
			}
		}
		
		fmt.Printf("Created game: Season %s, Date: %s, Host: %s, Winner: %s, Second: %s\n",
			seasonName, gameDate.Format("2006-01-02"), hostName, winnerStr, secondStr)
	}
}