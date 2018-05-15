
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_images ADD width INT;
ALTER TABLE product_images ADD height INT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_images DROP COLUMN IF EXISTS width;
ALTER TABLE product_images DROP COLUMN IF EXISTS height;
