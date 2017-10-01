
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE webhook_calls (
    id serial PRIMARY KEY,
    place_id bigint REFERENCES places (id) ON DELETE SET NULL,
    data jsonb DEFAULT '{}' NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL
);
CREATE INDEX webhook_calls_idx_place ON webhook_calls (place_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE webhook_calls;
