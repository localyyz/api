
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_tags ADD COLUMN type smallint default 0 NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_tags DROP COLUMN type;
