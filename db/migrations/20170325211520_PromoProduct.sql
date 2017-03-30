
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos ADD COLUMN product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE promos DROP COLUMN product_id;
