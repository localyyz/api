
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE collections ADD COLUMN deleted_at timestamp without time zone;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE collections DROP COLUMN IF EXISTS deleted_at;
