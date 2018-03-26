
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE merchant_approvals (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,

    collection smallint NOT NULL DEFAULT 0,
    category smallint NOT NULL DEFAULT 0,
    price_range smallint NOT NULL DEFAULT 0,

    rejection_reason text,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp,

    approved_at timestamp,
    rejected_at timestamp
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS merchant_approvals;
