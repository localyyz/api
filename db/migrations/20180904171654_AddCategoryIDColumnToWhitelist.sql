
-- +goose Up
ALTER TABLE whitelist ADD COLUMN category_id bigint;
ALTER TABLE whitelist ADD FOREIGN KEY (category_id) REFERENCES categories (id);
ALTER TABLE whitelist DROP COLUMN image_url;
ALTER TABLE whitelist DROP CONSTRAINT unique_product_category_value;
ALTER TABLE whitelist ADD CONSTRAINT unique_whitelist_value_gender unique (value, gender);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE whitelist DROP COLUMN category_id;
