
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE webhooks (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    topic text NOT NULL,
    external_id bigint NOT NULL,
    last_synced_at timestamp,
    created_at timestamp DEFAULT now() NOT NULL
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS webhooks;
