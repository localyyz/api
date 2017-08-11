-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE carts (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status smallint DEFAULT 0 NOT NULL,

    -- shopify cart token mapping
    etc jsonb DEFAULT '{}' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);

CREATE TABLE cart_items (
    id serial PRIMARY KEY,
    cart_id bigint NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    variant_id bigint NOT NULL REFERENCES promos (id) ON DELETE CASCADE,
    
    quantity smallint DEFAULT 0 NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp,

    CONSTRAINT unique_cart_product_variant UNIQUE (cart_id, product_id, variant_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
