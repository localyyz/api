
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos ADD column user_id smallint DEFAULT 0 NOT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE promos DELETE column user_id;