
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE notifications ALTER COLUMN cart_id DROP NOT NULL;
ALTER TABLE notifications ALTER COLUMN variant_id DROP NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

