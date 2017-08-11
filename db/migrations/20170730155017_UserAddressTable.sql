
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE user_addresses (
    id serial PRIMARY KEY,

    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    is_shipping boolean NOT NULL DEFAULT FALSE,
    is_billing boolean NOT NULL DEFAULT FALSE,

    first_name varchar(128) NOT NULL DEFAULT '',
    last_name varchar(128) NOT NULL DEFAULT '',
    address varchar(256) NOT NULL DEFAULT '',
    address_opt varchar(128) NOT NULL DEFAULT '',
    city varchar(128) NOT NULL DEFAULT '',
    country varchar(128) NOT NULL DEFAULT '',
    province varchar(123) NOT NULL DEFAULT '',
    zip varchar(64) NOT NULL DEFAULT '',

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS user_addresses;

