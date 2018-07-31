
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE collections ADD COLUMN merchant_id INT DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE collections DROP COLUMN IF EXISTS merchant_id;
