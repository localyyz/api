
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD COLUMN category jsonb DEFAULT '{}' NOT NULL;
CREATE INDEX products_category_ginidx ON products USING gin (category);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN category;
