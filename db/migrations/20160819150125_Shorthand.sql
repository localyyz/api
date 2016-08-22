
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE locales ADD COLUMN shorthand text DEFAULT '' NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE locales DROP COLUMN shorthand;
