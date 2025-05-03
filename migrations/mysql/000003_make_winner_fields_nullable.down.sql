-- Revert winner_id back to NOT NULL
ALTER TABLE games MODIFY COLUMN winner_id INT NOT NULL;

-- Revert second_place_id back to NOT NULL
ALTER TABLE games MODIFY COLUMN second_place_id INT NOT NULL;