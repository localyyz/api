
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE places ALTER COLUMN tos_ip TYPE varchar(64);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

