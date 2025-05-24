package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/klausbreyer/pokerhans/internal/models"
)

// Handler holds dependencies for the handlers
type Handler struct {
	Logger *log.Logger
	DB     *sql.DB
	Repo   *models.Repository
	Tmpl   *template.Template
}

// New creates a new Handler
func New(logger *log.Logger, db *sql.DB) *Handler {
	// Parse templates
	templatesPath := filepath.Join("web", "templates", "*.html")
	logger.Printf("DEBUG: Loading templates from %s", templatesPath)

	templates, err := filepath.Glob(templatesPath)
	if err != nil {
		logger.Printf("ERROR: Failed to glob templates: %v", err)
		// Continue with empty list if glob fails
		templates = []string{}
	}

	// Log found templates
	for _, t := range templates {
		logger.Printf("DEBUG: Found template file: %s", t)
	}

	// Simplify template parsing - use Must to panic on errors
	logger.Printf("DEBUG: Using simple template.ParseGlob approach")
	tmpl := template.Must(template.ParseGlob(templatesPath))

	// Log the templates we found
	logger.Printf("DEBUG: Templates found:")
	for _, t := range tmpl.Templates() {
		logger.Printf("DEBUG:   - Template: %s", t.Name())
	}

	return &Handler{
		Logger: logger,
		DB:     db,
		Repo:   models.NewRepository(db),
		Tmpl:   tmpl,
	}
}

// HomeHandler handles the root route
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Printf("ACTION: HomeHandler - Getting seasons")

	seasons, err := h.Repo.GetSeasons()
	if err != nil {
		h.Logger.Printf("ERROR: Getting seasons failed: %v", err)
		http.Error(w, "Failed to load seasons", http.StatusInternalServerError)
		return
	}

	h.Logger.Printf("DATA: Found %d seasons", len(seasons))

	// If there are seasons, redirect to the first one
	if len(seasons) > 0 {
		redirectURL := "/season/" + strconv.Itoa(seasons[0].ID)
		h.Logger.Printf("REDIRECT: To %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// No seasons, render empty home page
	h.Logger.Printf("RENDER: Rendering layout template with home content")
	data := struct {
		CurrentYear int
	}{
		CurrentYear: time.Now().Year(),
	}

	// Capture the template output to inspect it
	var buf bytes.Buffer
	err = h.Tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		h.Logger.Printf("ERROR: Template rendering failed: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	output := buf.String()
	h.Logger.Printf("HOME_TEMPLATE_OUTPUT_LENGTH: %d bytes", len(output))
	if len(output) == 0 {
		h.Logger.Printf("ERROR: Home template output is empty!")
	} else if len(output) < 50 {
		h.Logger.Printf("HOME_TEMPLATE_OUTPUT: %s", output) // Log only if it's short
	} else {
		h.Logger.Printf("HOME_TEMPLATE_OUTPUT_FIRST_50: %s...", output[:50])
	}

	// Set Content-Type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the output to the response
	_, err = w.Write(buf.Bytes())
	if err != nil {
		h.Logger.Printf("ERROR: Failed to write home response: %v", err)
		return
	}
}

// SeasonHandler handles the season view
func (h *Handler) SeasonHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Printf("ACTION: SeasonHandler - Processing season view")

	// Extract season ID from URL
	path := r.URL.Path
	re := regexp.MustCompile(`/season/(\d+)`)
	matches := re.FindStringSubmatch(path)

	if len(matches) < 2 {
		h.Logger.Printf("ERROR: Invalid season URL: %s", path)
		http.Error(w, "Invalid season URL", http.StatusBadRequest)
		return
	}

	seasonID, err := strconv.Atoi(matches[1])
	if err != nil {
		h.Logger.Printf("ERROR: Invalid season ID: %s", matches[1])
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	h.Logger.Printf("PARAM: Season ID = %d", seasonID)

	// Get all seasons for the dropdown
	seasons, err := h.Repo.GetSeasons()
	if err != nil {
		h.Logger.Printf("ERROR: Getting seasons failed: %v", err)
		http.Error(w, "Failed to load seasons", http.StatusInternalServerError)
		return
	}
	h.Logger.Printf("DATA: Found %d seasons total", len(seasons))

	// Get players for this season
	players, err := h.Repo.GetSeasonPlayers(seasonID)
	if err != nil {
		h.Logger.Printf("ERROR: Getting players failed: %v", err)
		http.Error(w, "Failed to load players", http.StatusInternalServerError)
		return
	}
	h.Logger.Printf("DATA: Found %d players for season %d", len(players), seasonID)

	// Get games for this season
	games, err := h.Repo.GetGames(seasonID)
	if err != nil {
		h.Logger.Printf("ERROR: Getting games failed: %v", err)
		http.Error(w, "Failed to load games", http.StatusInternalServerError)
		return
	}
	h.Logger.Printf("DATA: Found %d games for season %d", len(games), seasonID)

	// Reverse the order of games to show oldest first
	sort.Slice(games, func(i, j int) bool {
		return games[i].GameDate.Before(games[j].GameDate)
	})

	// Separate players into visited and not visited
	var visited, notVisited []models.PlayerStatus
	for _, p := range players {
		if p.HasHosted {
			visited = append(visited, p)
		} else {
			notVisited = append(notVisited, p)
		}
	}

	// Sort visited players by game date (oldest first)
	sort.Slice(visited, func(i, j int) bool {
		return visited[i].GameDate.Before(visited[j].GameDate)
	})

	// Sort players to visit by created_at date (oldest first)
	sort.Slice(notVisited, func(i, j int) bool {
		return notVisited[i].CreatedAt.Before(notVisited[j].CreatedAt)
	})
	h.Logger.Printf("DATA: %d players visited, %d players to visit", len(visited), len(notVisited))

	// Get all players for the dropdowns
	allPlayers, err := h.Repo.GetAllPlayers()
	if err != nil {
		h.Logger.Printf("ERROR: Getting all players failed: %v", err)
		http.Error(w, "Failed to load players", http.StatusInternalServerError)
		return
	}
	h.Logger.Printf("DATA: Found %d total players in system", len(allPlayers))

	// Current season
	var currentSeason models.Season
	for _, s := range seasons {
		if s.ID == seasonID {
			currentSeason = s
			break
		}
	}
	h.Logger.Printf("DATA: Current season name: %s", currentSeason.Name)

	// Determine if this is the latest season (highest ID)
	var highestID int
	for _, s := range seasons {
		if s.ID > highestID {
			highestID = s.ID
		}
	}
	isLatestSeason := seasonID == highestID

	data := struct {
		Seasons        []models.Season
		CurrentSeason  models.Season
		VisitedPlayers []models.PlayerStatus
		ToVisitPlayers []models.PlayerStatus
		Games          []models.Game
		AllPlayers     []models.Player
		CurrentDate    string
		CurrentYear    int
		IsLatestSeason bool
	}{
		Seasons:        seasons,
		CurrentSeason:  currentSeason,
		VisitedPlayers: visited,
		ToVisitPlayers: notVisited,
		Games:          games,
		AllPlayers:     allPlayers,
		CurrentDate:    time.Now().Format("2006-01-02"),
		CurrentYear:    time.Now().Year(),
		IsLatestSeason: isLatestSeason,
	}

	h.Logger.Printf("RENDER: Rendering layout template with season content")

	// Capture the template output to inspect it
	var buf bytes.Buffer
	err = h.Tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		h.Logger.Printf("ERROR: Template rendering failed: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	output := buf.String()
	h.Logger.Printf("TEMPLATE_OUTPUT_LENGTH: %d bytes", len(output))
	if len(output) == 0 {
		h.Logger.Printf("ERROR: Template output is empty!")
	} else if len(output) < 50 {
		h.Logger.Printf("TEMPLATE_OUTPUT: %s", output) // Log only if it's short
	} else {
		h.Logger.Printf("TEMPLATE_OUTPUT_FIRST_50: %s...", output[:50])
	}

	// Set Content-Type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the output to the response
	_, err = w.Write(buf.Bytes())
	if err != nil {
		h.Logger.Printf("ERROR: Failed to write response: %v", err)
		return
	}

	h.Logger.Printf("SUCCESS: Season page rendered successfully")
}

// AddGameHandler handles adding a new game
func (h *Handler) AddGameHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Printf("ACTION: AddGameHandler - Processing game submission")

	if err := r.ParseForm(); err != nil {
		h.Logger.Printf("ERROR: Parsing form data: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Parse form values
	seasonID, err := strconv.Atoi(r.FormValue("season_id"))
	if err != nil {
		h.Logger.Printf("ERROR: Invalid season_id: %s", r.FormValue("season_id"))
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	hostID, err := strconv.Atoi(r.FormValue("host_id"))
	if err != nil {
		h.Logger.Printf("ERROR: Invalid host_id: %s", r.FormValue("host_id"))
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	// Handle winner_id (optional)
	var winnerID *int
	winnerIDStr := r.FormValue("winner_id")
	if winnerIDStr != "" {
		winVal, err := strconv.Atoi(winnerIDStr)
		if err != nil {
			h.Logger.Printf("ERROR: Invalid winner_id: %s", winnerIDStr)
			http.Error(w, "Invalid winner ID", http.StatusBadRequest)
			return
		}
		winnerID = &winVal
	}

	// Handle second_place_id (optional)
	var secondPlaceID *int
	secondPlaceIDStr := r.FormValue("second_place_id")
	if secondPlaceIDStr != "" {
		secondVal, err := strconv.Atoi(secondPlaceIDStr)
		if err != nil {
			h.Logger.Printf("ERROR: Invalid second_place_id: %s", secondPlaceIDStr)
			http.Error(w, "Invalid second place ID", http.StatusBadRequest)
			return
		}
		secondPlaceID = &secondVal
	}

	gameDate, err := time.Parse("2006-01-02", r.FormValue("game_date"))
	if err != nil {
		h.Logger.Printf("ERROR: Invalid game_date format: %s", r.FormValue("game_date"))
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Log parameters with pointer-safe handling
	// Create string representations for logging
	var winnerIDLogStr, secondPlaceIDLogStr string

	winnerIDLogStr = "null"
	if winnerID != nil {
		winnerIDLogStr = fmt.Sprintf("%d", *winnerID)
	}

	secondPlaceIDLogStr = "null"
	if secondPlaceID != nil {
		secondPlaceIDLogStr = fmt.Sprintf("%d", *secondPlaceID)
	}

	h.Logger.Printf("PARAM: Season ID = %d, Host ID = %d, Winner ID = %s, Second Place ID = %s, Game Date = %s",
		seasonID, hostID, winnerIDLogStr, secondPlaceIDLogStr, gameDate.Format("2006-01-02"))

	// Add game to database
	err = h.Repo.AddGame(seasonID, hostID, winnerID, secondPlaceID, gameDate)
	if err != nil {
		h.Logger.Printf("ERROR: Adding game to database: %v", err)
		http.Error(w, "Failed to add game", http.StatusInternalServerError)
		return
	}

	h.Logger.Printf("SUCCESS: Game added successfully")

	// Redirect back to season page
	redirectURL := "/season/" + strconv.Itoa(seasonID)
	h.Logger.Printf("REDIRECT: To %s", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// UpdateGameDateHandler handles updating game dates
func (h *Handler) UpdateGameDateHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Printf("ACTION: UpdateGameDateHandler - Processing game date update")

	if err := r.ParseForm(); err != nil {
		h.Logger.Printf("ERROR: Parsing form data: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Parse form values
	gameID, err := strconv.Atoi(r.FormValue("game_id"))
	if err != nil {
		h.Logger.Printf("ERROR: Invalid game_id: %s", r.FormValue("game_id"))
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	seasonID, err := strconv.Atoi(r.FormValue("season_id"))
	if err != nil {
		h.Logger.Printf("ERROR: Invalid season_id: %s", r.FormValue("season_id"))
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	newDate, err := time.Parse("2006-01-02", r.FormValue("new_date"))
	if err != nil {
		h.Logger.Printf("ERROR: Invalid new_date format: %s", r.FormValue("new_date"))
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	h.Logger.Printf("PARAM: Game ID = %d, Season ID = %d, New Date = %s",
		gameID, seasonID, newDate.Format("2006-01-02"))

	// Update game date in database
	err = h.Repo.UpdateGameDate(gameID, newDate)
	if err != nil {
		h.Logger.Printf("ERROR: Updating game date in database: %v", err)
		http.Error(w, "Failed to update game date", http.StatusInternalServerError)
		return
	}

	h.Logger.Printf("SUCCESS: Game date updated successfully")

	// Redirect back to season page
	redirectURL := "/season/" + strconv.Itoa(seasonID)
	h.Logger.Printf("REDIRECT: To %s", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
