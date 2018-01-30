
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- NOTE: for future reference. storing as 15,6 precision numeric is future
-- proofing any kind of currency conversion.
ALTER TABLE product_variants ADD COLUMN price numeric(15,6) DEFAULT 0 NOT NULL;
ALTER TABLE product_variants ADD COLUMN prev_price numeric(15,6);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_variants DROP COLUMN price;
ALTER TABLE product_variants DROP COLUMN prev_price;
