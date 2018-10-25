
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX product_images_external_id_idx ON product_images(external_id);
ALTER TABLE product_images ADD CONSTRAINT unique_product_images_external_id_product UNIQUE (product_id, external_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX product_images_external_id_idx;
ALTER TABLE product_images DROP CONSTRAINT IF EXISTS unique_product_images_external_id_product;
