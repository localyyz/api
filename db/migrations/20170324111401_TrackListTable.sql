
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE track_list (
    id serial PRIMARY KEY,

    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    sales_url text DEFAULT '' NOT NULL,

    last_tracked_at timestamp,
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE track_list;
