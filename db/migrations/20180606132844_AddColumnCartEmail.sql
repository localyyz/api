
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE carts ADD COLUMN email varchar(64) DEFAULT '' NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE carts DROP COLUMN IF EXISTS email;
