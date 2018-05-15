
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_images ADD width INT DEFAULT 0;
ALTER TABLE product_images ADD height INT DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_images DROP COLUMN IF EXISTS width;
ALTER TABLE product_images DROP COLUMN IF EXISTS height;
