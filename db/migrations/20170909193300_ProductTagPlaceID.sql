
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_tags ADD COLUMN place_id bigint REFERENCES places (id) ON DELETE CASCADE;

UPDATE product_tags pt
SET place_id = i.place_id
FROM (
	SELECT id, place_id
	FROM products
) i
WHERE i.id = pt.product_id;

ALTER TABLE product_tags ALTER COLUMN place_id SET NOT NULL;
CREATE INDEX ON product_tags (place_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE product_tags DROP COLUMN place_id;
