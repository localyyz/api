
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE collections ADD COLUMN place_ids bigint[];
ALTER TABLE collections ADD COLUMN categories varchar(64)[];

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TALBE collections DROP COLUMN IF EXISTS place_ids;
ALTER TALBE collections DROP COLUMN IF EXISTS categories;
