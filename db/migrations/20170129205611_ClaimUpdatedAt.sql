
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE claims ADD column updated_at timestamp;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE claims DROP column updated_at;

