
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE user_access (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,

    admin boolean NOT NULL DEFAULT false,
    promoter boolean NOT NULL DEFAULT false,
    member boolean NOT NULL DEFAULT false,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp
);



-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS user_access;
