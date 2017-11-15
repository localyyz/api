
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX CONCURRENTLY products_created_at_idx ON products (created_at);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

