
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
DROP TABLE IF EXISTS user_deals;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

