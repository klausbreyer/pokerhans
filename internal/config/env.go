package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DBConfig contains database connection configuration
type DBConfig struct {
	User string
	Pass string
	Host string
	Port string
	Name string
}

// GetDBConfig returns the database configuration from environment variables
func GetDBConfig() DBConfig {
	return DBConfig{
		User: getEnvWithDefault("DB_USER", "root"),
		Pass: os.Getenv("DB_PASS"),
		Host: getEnvWithDefault("DB_HOST", "localhost"),
		Port: getEnvWithDefault("DB_PORT", "3306"),
		Name: getEnvWithDefault("DB_NAME", "pokerhans"),
	}
}

// DSN returns a formatted MySQL DSN (Data Source Name) string
func (c DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.User, c.Pass, c.Host, c.Port, c.Name)
}

// MigrateDSN returns a DSN string formatted for golang-migrate CLI
func (c DBConfig) MigrateDSN() string {
	return fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.User, c.Pass, c.Host, c.Port, c.Name)
}

// getEnvWithDefault returns the value of the environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// LoadEnv loads environment variables from a .env file
func LoadEnv(filePath string) error {
	// If file doesn't exist, skip loading
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	// Open the .env file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Split on the first equals sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// Set the environment variable if it's not already set
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set (allows command line to override .env)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
