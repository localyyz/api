
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE feature_products (
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    ordering smallint DEFAULT 1 NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    featured_at timestamp DEFAULT NOW() NOT NULL,
    end_featured_at timestamp
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE feature_products;
