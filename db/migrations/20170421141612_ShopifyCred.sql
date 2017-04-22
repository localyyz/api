
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE shopify_creds (
    id serial PRIMARY KEY,

    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    auth_access_token text NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE shopify_creds;
