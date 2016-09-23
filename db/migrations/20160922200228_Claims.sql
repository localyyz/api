
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE claims (
    id serial PRIMARY KEY,
    place_id bigint REFERENCES places (id) ON DELETE SET NULL,
    promo_id bigint REFERENCES promos (id) ON DELETE SET NULL,
    user_id bigint REFERENCES users (id),

    hash text default '' NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE claims;
