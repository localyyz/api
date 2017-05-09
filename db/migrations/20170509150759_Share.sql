
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE shares (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,

    network varchar(64) NOT NULL DEFAULT '',
    network_share_id text NOT NULL DEFAULT '',
    reach integer NOT NULL DEFAULT 0,

    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shares;
