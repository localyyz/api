
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE users ADD COLUMN password_hash VARCHAR(60) DEFAULT '' NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE users DROP COLUMN password_hash;
