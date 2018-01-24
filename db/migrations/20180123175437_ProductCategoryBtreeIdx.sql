
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX products_category_type_idx ON products((category->>'type'));
CREATE INDEX products_category_value_idx ON products((category->>'value'));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

