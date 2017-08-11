
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos DROP COLUMN user_id;
ALTER TABLE promos DROP COLUMN type;
ALTER TABLE promos DROP COLUMN image_url;
ALTER TABLE promos DROP COLUMN start_at;
ALTER TABLE promos DROP COLUMN end_at;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
