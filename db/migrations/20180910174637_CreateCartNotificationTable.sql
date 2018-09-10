
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE cart_notifications (
    id SERIAL PRIMARY KEY,
    cart_id INT REFERENCES carts (id) ON DELETE CASCADE NOT NULL,
    product_id INT REFERENCES products (id) ON DELETE CASCADE NOT NULL,
    variant_id INT REFERENCES product_variants (id) ON DELETE CASCADE NOT NULL,
    external_id varchar(64) DEFAULT '' NOT NULL,
    heading text DEFAULT '' NOT NULL,
    content text DEFAULT '' NOT NULL,
    scheduled_at timestamp without time zone DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE cart_notifications;
