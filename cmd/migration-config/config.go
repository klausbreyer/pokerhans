package main

import (
	"fmt"
	"os"

	"github.com/klausbreyer/pokerhans/internal/config"
)

// Simple utility to get database config for migrations
// To be used in Makefile to avoid hardcoding database credentials
func main() {
	dbConfig := config.GetDBConfig()
	fmt.Print(dbConfig.MigrateDSN())
	os.Exit(0)
}
