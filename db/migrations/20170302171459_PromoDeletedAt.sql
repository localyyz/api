
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos ADD COLUMN deleted_at timestamp;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE promos DELETE column deleted_at;
