
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos ADD COLUMN etc jsonb DEFAULT '{}' NOT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE promos DROP COLUMN etc;
