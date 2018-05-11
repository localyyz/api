
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD COLUMN status SMALLINT DEFAULT 0 NOT NULL; 
CREATE INDEX product_status_index ON products (status);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN IF EXISTS status;
