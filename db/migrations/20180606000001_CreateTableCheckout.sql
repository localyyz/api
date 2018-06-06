
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE checkouts (
    id SERIAL PRIMARY KEY,
    cart_id bigint REFERENCES carts (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    status smallint DEFAULT 0 NOT NULL,

    token varchar(64) DEFAULT '' NOT NULL,
    name varchar(64) DEFAULT '' NOT NULL,
    web_url text DEFAULT '' NOT NULL,

    payment_account_id varchar(64) DEFAULT '' NOT NULL,
    customer_id bigint,
    payment_id bigint,

    currency varchar(8) DEFAULT '' NOT NULL,
    discount_code varchar(32) DEFAULT '' NOT NULL,
	applied_discount jsonb DEFAULT '{}' NOT NULL,

    shipping_line jsonb DEFAULT '{}' NOT NULL,
    total_shipping numeric(15,6) DEFAULT 0.0 NOT NULL,

    tax_lines jsonb DEFAULT '[]' NOT NULL,
    total_tax numeric(15,6) DEFAULT 0.0 NOT NULL,
    taxes_included bool DEFAULT false NOT NULL,

    subtotal_price numeric(15,6) DEFAULT 0.0 NOT NULL,
    total_price numeric(15,6) DEFAULT 0.0 NOT NULL, 
    payment_due varchar(16) DEFAULT '' NOT NULL,

    created_at timestamp DEFAULT NOW() NOT NULL,
    updated_at timestamp
);

CREATE INDEX checkouts_cart_id_idx ON checkouts (cart_id);
CREATE INDEX checkouts_place_id_idx ON checkouts (place_id);
CREATE INDEX checkouts_token_idx ON checkouts (token);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS checkouts;
