
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE product_categories (
    id serial PRIMARY KEY,
    name varchar(64) NOT NULL,
    value VARCHAR(2048)[] DEFAULT '{}' NOT NULL
);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS product_categories;
