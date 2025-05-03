// This file is deprecated - migrations are now handled using
// github.com/golang-migrate/migrate with SQL files in the migrations/mysql directory.
//
// Keeping for reference purposes only.
package db

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}