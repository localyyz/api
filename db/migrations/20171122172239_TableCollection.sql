
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE collections (
    id serial PRIMARY KEY,

    name varchar(512) DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    ordering smallint DEFAULT 1 NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
);

CREATE TABLE collection_products (
    id serial PRIMARY KEY,

    collection_id bigint NOT NULL REFERENCES collections (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,

    created_at timestamp DEFAULT now() NOT NULL,

    CONSTRAINT unique_collection_product UNIQUE (collection_id, product_id)
);
CREATE INDEX collection_products_collection_id_idx ON collection_products (collection_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS collection_products;
DROP TABLE IF EXISTS collections;
