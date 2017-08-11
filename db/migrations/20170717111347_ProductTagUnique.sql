
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_tags DROP CONSTRAINT unique_product_tag;
ALTER TABLE product_tags ADD CONSTRAINT unique_product_value_type UNIQUE (product_id, value, type);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
