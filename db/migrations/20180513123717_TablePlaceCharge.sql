
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE place_charges (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    external_id bigint NOT NULL,
    charge_type smallint DEFAULT 0 NOT NULL,
    amount numeric(15,6) DEFAULT 0.0 NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,
    expire_at timestamp NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS place_charges;
