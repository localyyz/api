
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE cart_items ADD COLUMN place_id bigint ;
ALTER TABLE cart_items ADD FOREIGN KEY (place_id) REFERENCES places (id) ON DELETE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE cart_items DROP COLUMN place_id;
