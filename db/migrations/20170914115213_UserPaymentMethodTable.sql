
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE user_payment_methods (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    brand varchar(16) DEFAULT '' NOT NULL,
    exp_month smallint DEFAULT 0 NOT NULL,
    exp_year smallint DEFAULT 0 NOT NULL,
    last_four varchar(4) DEFAULT '' NOT NULL,
    country varchar(4) DEFAULT '' NOT NULL,
    stripe_card_id varchar(64) DEFAULT '' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS user_payment_methods;
