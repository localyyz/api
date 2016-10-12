
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE places ADD COLUMN image_url text;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE places DELETE COLUMN image_url;
