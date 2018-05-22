
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD score SMALLINT DEFAULT -1;
ALTER TABLE product_images ADD score SMALLINT DEFAULT -1;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN IF EXISTS score;
ALTER TABLE product_images DROP COLUMN IF EXISTS score;
