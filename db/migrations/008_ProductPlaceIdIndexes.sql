
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX CONCURRENTLY product_variants_place_id_idx ON product_variants (place_id);
CREATE INDEX CONCURRENTLY products_place_id_idx ON products (place_id);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

