package models

import (
	"database/sql"
	"time"
)

// Season represents a poker season
type Season struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Player represents a poker player
type Player struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Game represents a poker game
type Game struct {
	ID            int       `json:"id"`
	SeasonID      int       `json:"season_id"`
	HostID        int       `json:"host_id"`
	WinnerID      int       `json:"winner_id"`
	SecondPlaceID int       `json:"second_place_id"`
	GameDate      time.Time `json:"game_date"`
	CreatedAt     time.Time `json:"created_at"`
	
	// Additional fields for display
	HostName        string `json:"host_name"`
	WinnerName      string `json:"winner_name"`
	SecondPlaceName string `json:"second_place_name"`
}

// PlayerStatus includes information about whether a player has hosted a game
type PlayerStatus struct {
	Player
	HasHosted bool      `json:"has_hosted"`
	GameDate  time.Time `json:"game_date,omitempty"`
}

// Repository provides methods to interact with the database
type Repository struct {
	DB *sql.DB
}

// NewRepository creates a new Repository with the given database connection
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// GetSeasons returns all seasons
func (r *Repository) GetSeasons() ([]Season, error) {
	query := "SELECT id, name, created_at FROM seasons ORDER BY created_at DESC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasons []Season
	for rows.Next() {
		var s Season
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt); err != nil {
			return nil, err
		}
		seasons = append(seasons, s)
	}

	return seasons, nil
}

// GetSeasonPlayers returns all players for a given season with their hosting status
func (r *Repository) GetSeasonPlayers(seasonID int) ([]PlayerStatus, error) {
	query := `
		SELECT 
			p.id, 
			p.name, 
			p.created_at,
			g.id IS NOT NULL as has_hosted,
			IF(g.id IS NULL, '0001-01-01', DATE_FORMAT(g.game_date, '%Y-%m-%d')) as game_date
		FROM 
			players p
		JOIN 
			season_players sp ON p.id = sp.player_id
		LEFT JOIN 
			games g ON p.id = g.host_id AND g.season_id = sp.season_id
		WHERE 
			sp.season_id = ?
		ORDER BY 
			CASE WHEN g.id IS NULL THEN 0 ELSE 1 END, p.name
	`
	
	rows, err := r.DB.Query(query, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerStatus
	for rows.Next() {
		var p PlayerStatus
		var gameDateStr string
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.HasHosted, &gameDateStr); err != nil {
			return nil, err
		}
		
		// Parse the game date string
		p.GameDate, err = time.Parse("2006-01-02", gameDateStr)
		if err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, nil
}

// AddGame adds a new game to the database
func (r *Repository) AddGame(seasonID, hostID, winnerID, secondPlaceID int, gameDate time.Time) error {
	query := `
		INSERT INTO games (season_id, host_id, winner_id, second_place_id, game_date) 
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.DB.Exec(query, seasonID, hostID, winnerID, secondPlaceID, gameDate)
	return err
}

// GetGames returns all games for a given season
func (r *Repository) GetGames(seasonID int) ([]Game, error) {
	query := `
		SELECT 
			g.id, 
			g.season_id, 
			g.host_id, 
			g.winner_id, 
			g.second_place_id, 
			g.game_date, 
			g.created_at,
			host.name as host_name,
			winner.name as winner_name,
			second.name as second_place_name
		FROM 
			games g
		JOIN 
			players host ON g.host_id = host.id
		JOIN 
			players winner ON g.winner_id = winner.id
		JOIN 
			players second ON g.second_place_id = second.id
		WHERE 
			g.season_id = ?
		ORDER BY 
			g.game_date DESC
	`
	
	rows, err := r.DB.Query(query, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		if err := rows.Scan(
			&g.ID, 
			&g.SeasonID, 
			&g.HostID, 
			&g.WinnerID, 
			&g.SecondPlaceID, 
			&g.GameDate, 
			&g.CreatedAt,
			&g.HostName,
			&g.WinnerName,
			&g.SecondPlaceName,
		); err != nil {
			return nil, err
		}
		games = append(games, g)
	}

	return games, nil
}

// GetAllPlayers returns all players in the system
func (r *Repository) GetAllPlayers() ([]Player, error) {
	query := "SELECT id, name, created_at FROM players ORDER BY name"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, nil
}