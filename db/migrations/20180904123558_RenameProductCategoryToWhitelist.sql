
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_categories RENAME TO whitelist;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE whitelist RENAME TO product_categories;

