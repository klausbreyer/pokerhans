package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/go-sql-driver/mysql"
	
	"github.com/klausbreyer/pokerhans/internal/config"
)

// Connect establishes a connection to the MySQL database
func Connect() (*sql.DB, error) {
	dbConfig := config.GetDBConfig()
	dsn := dbConfig.DSN()
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// InitSchema sets up the database schema using migrations
func InitSchema(db *sql.DB) error {
	// Initialize the MySQL driver for migrations
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create MySQL migration driver: %w", err)
	}

	// Get the absolute path to the migrations directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	migrationsPath := filepath.Join(workDir, "migrations", "mysql")
	migrationSource := fmt.Sprintf("file://%s", migrationsPath)
	
	// Create a new migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		migrationSource,
		"mysql", 
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	
	return nil
}