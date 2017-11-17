
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE place_discounts (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    code varchar(255) NOT NULL,
    external_id bigint NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,

    CONSTRAINT unique_place_discount UNIQUE (place_id, external_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DELETE TABLE IF EXISTS place_discounts;
