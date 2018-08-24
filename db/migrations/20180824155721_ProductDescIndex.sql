
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX products_score_desc_index ON products (score DESC nulls last);
CREATE INDEX products_created_desc_index ON products (created_at DESC nulls last);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

