
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE products (
    id serial PRIMARY KEY,

    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    external_id text DEFAULT '' NOT NULL,
    title text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    etc jsonb DEFAULT '{}' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

