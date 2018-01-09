
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_categories ADD COLUMN gender smallint DEFAULT 0 NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_categories DROP COLUMN IF EXISTS gender;
