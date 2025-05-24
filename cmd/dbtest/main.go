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
	logger := log.New(os.Stdout, "dbtest: ", log.LstdFlags)

	// Load environment variables from .env file
	envPath := "../.env"
	if err := config.LoadEnv(envPath); err != nil {
		logger.Printf("Warning: Unable to load .env file: %v", err)
	}

	// Get DB Config
	dbConfig := config.GetDBConfig()
	logger.Printf("Using DSN: %s", dbConfig.DSN())

	// Try to connect to the database
	database, err := db.Connect()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Try to ping the database
	err = database.Ping()
	if err != nil {
		logger.Fatalf("Failed to ping database: %v", err)
	}

	// Query for seasons
	rows, err := database.Query("SELECT COUNT(*) FROM seasons")
	if err != nil {
		logger.Fatalf("Error querying seasons: %v", err)
	}
	defer rows.Close()

	// Count seasons
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			logger.Fatalf("Error scanning seasons count: %v", err)
		}
	}

	fmt.Printf("Database connection successful. Found %d seasons.\n", count)
}
