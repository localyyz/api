
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_categories ADD COLUMN mapping VARCHAR(64) DEFAULT '' NOT NULL;
CREATE INDEX product_categories_mapping_idx ON product_categories (mapping);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_categories DROP COLUMN IF EXISTS mapping;
