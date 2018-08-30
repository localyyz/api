
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE table shipping_zones (
    id SERIAL PRIMARY KEY,
    place_id INT REFERENCES places (id) ON DELETE CASCADE NOT NULL,
    external_id bigint,

    type smallint DEFAULT 0 NOT NULL,
    name text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    country text DEFAULT '' NOT NULL,
    regions jsonb DEFAULT '[]' NOT NULL,
    price numeric(15,6) DEFAULT 0.0 NOT NULL,

    weight_low numeric DEFAULT 0.0 NOT NULL,
    weight_high numeric DEFAULT 0.0 NOT NULL,
    subtotal_low numeric DEFAULT 0.0 NOT NULL,
    subtotal_high numeric DEFAULT 0.0 NOT NULL
);

CREATE INDEX shipping_zone_place_id_idx ON shipping_zones (place_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shipping_zones;
