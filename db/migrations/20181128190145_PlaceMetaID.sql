-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE place_meta ADD COLUMN id SERIAL PRIMARY KEY;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE place_meta DROP COLUMN IF EXISTS id;
