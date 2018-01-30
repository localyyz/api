
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE products SET tsv = to_tsvector(title);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
