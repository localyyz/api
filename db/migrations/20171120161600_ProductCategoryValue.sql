
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_categories RENAME value TO legacy_value;
ALTER TABLE product_categories ADD COLUMN value varchar(64) DEFAULT '' NOT NULL;
ALTER TABLE product_categories ADD COLUMN type smallint DEFAULT 0 NOT NULL;

UPDATE product_categories SET type = id;

INSERT INTO product_categories (value, type, name)
SELECT unnest(legacy_value) as value, type, ''
FROM product_categories;

ALTER TABLE product_categories DROP COLUMN legacy_value;
ALTER TABLE product_categories DROP COLUMN name;

DELETE FROM product_categories WHERE id < 7;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

