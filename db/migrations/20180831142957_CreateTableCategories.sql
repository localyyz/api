
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE table categories (
    id SERIAL PRIMARY KEY,
    value varchar(20) DEFAULT '' NOT NULL,
    label varchar(40) DEFAULT '' NOT NULL,
    image_url text,
    lft integer,
    rgt integer
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS categories;
