
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE products ADD COLUMN price numeric(15,6) DEFAULT 0.0 NOT NULL;
ALTER TABLE products ADD COLUMN discount_pct numeric DEFAULT 0.0 NOT NULL;

CREATE INDEX product_price_idx ON products USING btree (price);
CREATE INDEX product_discount_pct_idx ON products USING btree (discount_pct);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE products DROP COLUMN price;
ALTER TABLE products DROP COLUMN discount_pct;

DROP INDEX product_price_idx;
DROP INDEX product_discount_pct_idx;
