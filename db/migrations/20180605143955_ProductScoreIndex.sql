
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX product_score_idx ON products USING btree (score);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX product_score_idx;
