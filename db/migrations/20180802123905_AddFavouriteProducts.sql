
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE favourite_products(
  user_id INT NOT NULL,
  product_id INT NOT NULL,
  created_at timestamp without time zone DEFAULT now() NOT NULL,
  UNIQUE(user_id, product_id)
);

ALTER TABLE ONLY public.favourite_products
    ADD CONSTRAINT favourite_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.favourite_products
    ADD CONSTRAINT favourite_products_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS favourite_products;
