
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD COLUMN gender smallint DEFAULT 0 NOT NULL;
CREATE INDEX product_gender_idx ON products (gender);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN IF EXISTS gender;
