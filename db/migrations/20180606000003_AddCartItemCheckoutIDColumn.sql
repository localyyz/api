
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE cart_items ADD COLUMN checkout_id bigint REFERENCES checkouts (id) ON DELETE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE cart_items DROP COLUMN IF EXISTS checkout_id;
