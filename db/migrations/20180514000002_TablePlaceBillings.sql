
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE place_billings (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    plan_id bigint NOT NULL REFERENCES billing_plans (id),
    status smallint DEFAULT 0 NOT NULL,

    -- shopify external reference id
    external_id bigint NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    accepted_at timestamp
);

-- weird postgresql does not have partial unique constraint
-- fake it by create a partial unique index
CREATE UNIQUE INDEX unique_place_active_billing ON place_billings (place_id, plan_id, status) WHERE (status = 3);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS place_billings;
