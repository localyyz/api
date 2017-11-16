
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX product_variants_offer_id_idx ON product_variants (offer_id);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

