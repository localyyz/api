
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shares ADD COLUMN hash varchar(36) DEFAULT '' NOT NULL;
ALTER TABLE shares ADD CONSTRAINT unique_hash UNIQUE (hash);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE shares DROP COLUMN hash;
