
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_tags DROP COLUMN IF EXISTS id;
ALTER TABLE product_tags DROP COLUMN IF EXISTS created_at;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
