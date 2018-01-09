
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_categories DROP COLUMN id;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

