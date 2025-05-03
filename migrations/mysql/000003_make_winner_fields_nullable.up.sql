-- Modify winner_id to allow NULL
ALTER TABLE games MODIFY COLUMN winner_id INT NULL;

-- Modify second_place_id to allow NULL
ALTER TABLE games MODIFY COLUMN second_place_id INT NULL;