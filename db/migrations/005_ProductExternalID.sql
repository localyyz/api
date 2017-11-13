
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products RENAME external_id TO external_handle;
ALTER TABLE products ADD COLUMN external_id bigint;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN IF EXISTS external_id;
ALTER TABLE products RENAME external_handle TO external_id;
