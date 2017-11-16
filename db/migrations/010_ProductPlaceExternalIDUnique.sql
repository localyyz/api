
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD CONSTRAINT unique_product_place_external_id UNIQUE (place_id, external_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

