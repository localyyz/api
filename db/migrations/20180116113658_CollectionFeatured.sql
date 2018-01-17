
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE collections ADD COLUMN featured bool DEFAULT false NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE collections DROP COLUMN featured;
