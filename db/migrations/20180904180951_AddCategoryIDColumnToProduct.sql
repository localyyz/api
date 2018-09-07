
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD COLUMN category_id bigint;
ALTER TABLE products ADD FOREIGN KEY (category_id) REFERENCES categories (id);
CREATE INDEX ON products(category_id);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN IF EXISTS category_id;

