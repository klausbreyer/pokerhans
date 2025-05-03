package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/klausbreyer/pokerhans/internal/config"
	"github.com/klausbreyer/pokerhans/internal/db"
	"github.com/klausbreyer/pokerhans/internal/handlers"
)

func main() {
	// Set up logger
	logger := log.New(os.Stdout, "pokerhans: ", log.LstdFlags|log.Lshortfile)

	// Reset environment variables
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASS")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT") 
	os.Unsetenv("DB_NAME")

	// Load environment variables from .env file
	envPath := filepath.Join(".", ".env")
	if err := config.LoadEnv(envPath); err != nil {
		logger.Printf("Warning: Unable to load .env file: %v", err)
	}

	// Initialize DB
	database, err := db.Connect()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Note: Migrations are now handled separately with the migrate CLI.
	// Use 'make migrate-up' to run migrations before starting the server.

	// Set up handlers
	h := handlers.New(logger, database)

	// Static files
	staticDir := "./web/static"
	logger.Printf("DEBUG: Setting up static file server for directory: %s", staticDir)
	
	// Check if static directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		logger.Printf("ERROR: Static directory does not exist: %s", staticDir)
	} else {
		logger.Printf("DEBUG: Static directory exists, checking contents")
		files, err := filepath.Glob(filepath.Join(staticDir, "css/*"))
		if err != nil {
			logger.Printf("ERROR: Failed to list static CSS files: %v", err)
		} else {
			for _, f := range files {
				logger.Printf("DEBUG: Found static file: %s", f)
			}
		}
	}
	
	// Create a custom file server handler with logging
	fileServer := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("=== STATIC FILE REQUEST ===")
		logger.Printf("PATH: %s", r.URL.Path)
		logger.Printf("METHOD: %s", r.Method)
		logger.Printf("REMOTE: %s", r.RemoteAddr)
		logger.Printf("HANDLER: StaticFileServer")
		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	}))

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("=== ROUTE CALL ===")
		logger.Printf("PATH: %s", r.URL.Path)
		logger.Printf("METHOD: %s", r.Method)
		logger.Printf("REMOTE: %s", r.RemoteAddr)
		logger.Printf("USER-AGENT: %s", r.UserAgent())
		
		if r.URL.Path != "/" {
			if strings.HasPrefix(r.URL.Path, "/season/") && r.Method == "GET" {
				logger.Printf("HANDLER: SeasonHandler")
				h.SeasonHandler(w, r)
				return
			}
			logger.Printf("HANDLER: NotFound (404)")
			http.NotFound(w, r)
			return
		}
		
		logger.Printf("HANDLER: HomeHandler")
		h.HomeHandler(w, r)
	})

	http.HandleFunc("/game/add", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("=== ROUTE CALL ===")
		logger.Printf("PATH: /game/add")
		logger.Printf("METHOD: %s", r.Method)
		logger.Printf("REMOTE: %s", r.RemoteAddr)
		
		if r.Method != "POST" {
			logger.Printf("HANDLER: MethodNotAllowed (405)")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		logger.Printf("HANDLER: AddGameHandler")
		h.AddGameHandler(w, r)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	logger.Printf("=== SERVER STARTING ===")
	logger.Printf("LISTENING ON: http://localhost:%s", port)
	logger.Printf("ROUTES:")
	logger.Printf("  - http://localhost:%s/           -> HomeHandler", port)
	logger.Printf("  - http://localhost:%s/season/:id -> SeasonHandler", port)
	logger.Printf("  - http://localhost:%s/game/add   -> AddGameHandler (POST)", port)
	logger.Printf("  - http://localhost:%s/static/*   -> Static files", port)
	logger.Printf("Migration Note: Run 'make migrate-up' if you need to apply database migrations")
	logger.Printf("Tailwind CSS: Run 'make css-watch' in another terminal for CSS hot reloading")
	logger.Printf("=== SERVER READY ===")

	// Create a custom HTTP server to log when it starts listening
	server := &http.Server{
		Addr:    addr,
		Handler: nil, // Use default mux
	}

	// Log before starting to make it clear we're about to listen
	logger.Printf("Starting to listen on port %s...", port)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}