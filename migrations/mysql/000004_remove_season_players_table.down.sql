-- Recreate the season_players table
CREATE TABLE IF NOT EXISTS season_players (
    id INT AUTO_INCREMENT PRIMARY KEY,
    season_id INT NOT NULL,
    player_id INT NOT NULL,
    FOREIGN KEY (season_id) REFERENCES seasons(id),
    FOREIGN KEY (player_id) REFERENCES players(id),
    UNIQUE KEY unique_season_player (season_id, player_id)
);