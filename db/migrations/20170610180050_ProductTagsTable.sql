
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE product_tags (
    id serial PRIMARY KEY,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    value varchar(128) NOT NULL DEFAULT '',

    created_at timestamp DEFAULT now() NOT NULL,
    CONSTRAINT unique_product_tag UNIQUE (product_id, value)
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE product_tags;
