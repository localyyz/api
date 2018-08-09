
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
DELETE FROM collections WHERE lightning = true;
ALTER TABLE collections DROP COLUMN IF EXISTS lightning;
ALTER TABLE collections DROP COLUMN IF EXISTS start_at;
ALTER TABLE collections DROP COLUMN IF EXISTS end_at;
ALTER TABLE collections DROP COLUMN IF EXISTS status;
ALTER TABLE collections DROP COLUMN IF EXISTS cap;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
