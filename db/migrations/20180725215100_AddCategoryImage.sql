
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_categories ADD image_url text DEFAULT '' NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_categories DROP COLUMN image_url;;
