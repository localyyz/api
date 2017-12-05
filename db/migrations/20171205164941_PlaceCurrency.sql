
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE places ADD COLUMN currency VARCHAR(8) DEFAULT '' NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE places DROP COLUMN IF EXISTS currency;
