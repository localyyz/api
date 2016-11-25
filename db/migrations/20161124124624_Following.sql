
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE followings (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,

    created_at timestamp DEFAULT now() NOT NULL,
    CONSTRAINT unique_user_following UNIQUE (user_id, place_id)
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS followings;
