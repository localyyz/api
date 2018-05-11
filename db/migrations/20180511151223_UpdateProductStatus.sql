
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE products set status = 3 where category != '{}' and status = 0;
UPDATE products set status = 1 where category = '{}' and status = 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
