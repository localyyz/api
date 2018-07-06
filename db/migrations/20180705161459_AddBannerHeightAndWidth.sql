
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE collections ADD COLUMN image_width INT DEFAULT 0;
ALTER TABLE collections ADD COLUMN image_height INT DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE collections DROP COLUMN IF EXISTS image_width;
ALTER TABLE collections DROP COLUMN IF EXISTS image_height;
