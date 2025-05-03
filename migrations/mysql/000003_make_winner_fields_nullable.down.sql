-- First, update any NULL values with a default value (e.g., host_id or 1)
UPDATE games SET winner_id = COALESCE(winner_id, host_id) WHERE winner_id IS NULL;
UPDATE games SET second_place_id = COALESCE(second_place_id, host_id) WHERE second_place_id IS NULL;

-- Now revert the columns back to NOT NULL
ALTER TABLE games MODIFY COLUMN winner_id INT NOT NULL;
ALTER TABLE games MODIFY COLUMN second_place_id INT NOT NULL;