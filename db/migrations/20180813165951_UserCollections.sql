
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE user_collections (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users (id) ON DELETE CASCADE NOT NULL,
  title text DEFAULT '' NOT NULL,
  created_at TIMESTAMP without time zone DEFAULT now() NOT NULL,
  updated_at TIMESTAMP without time zone DEFAULT now() NOT NULL,
  deleted_at TIMESTAMP
);


CREATE TABLE user_collection_products (
   collection_id INT REFERENCES user_collections (id) ON DELETE CASCADE NOT NULL,
   product_id INT REFERENCES products (id) ON DELETE CASCADE NOT NULL,
   created_at TIMESTAMP without time zone DEFAULT now() NOT NULL,
   deleted_at TIMESTAMP without time zone,
   UNIQUE(collection_id, product_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS user_collection_products;
DROP TABLE IF EXISTS user_collections;