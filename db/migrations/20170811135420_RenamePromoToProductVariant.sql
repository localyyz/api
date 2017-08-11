
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos RENAME TO product_variants;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_variants RENAME TO promos;
