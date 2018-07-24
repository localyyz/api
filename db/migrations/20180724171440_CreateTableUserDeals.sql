
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE user_deals (
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    deal_id bigint REFERENCES collections (id) ON DELETE CASCADE,
    status smallint DEFAULT 0 NOT NULL,
    start_at timestamp DEFAULT NOW() NOT NULL,
    end_at timestamp NOT NULL
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS user_deals;
