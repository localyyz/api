
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
DROP TABLE IF EXISTS product_tags;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
