
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE favourite_products(
  user_id INT NOT NULL,
  product_id INT NOT NULL,
  created_at timestamp without time zone DEFAULT now() NOT NULL,
  UNIQUE(user_id, product_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS favourite_products;
