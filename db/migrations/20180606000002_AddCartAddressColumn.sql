
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE carts ADD COLUMN shipping_address jsonb DEFAULT '{}' NOT NULL;
ALTER TABLE carts ADD COLUMN billing_address jsonb DEFAULT '{}' NOT NULL;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE carts DROP COLUMN IF EXISTS shipping_address;
ALTER TABLE carts DROP COLUMN IF EXISTS billing_address;
