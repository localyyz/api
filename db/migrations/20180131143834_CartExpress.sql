
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE carts ADD COLUMN is_express bool DEFAULT false NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE carts DROP COLUMN is_express;
