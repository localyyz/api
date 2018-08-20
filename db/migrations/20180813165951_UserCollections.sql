
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
   UNIQUE(product_id, collection_id)
);


CREATE INDEX user_collections_user_id_idx ON user_collections (user_id);
CREATE INDEX user_collections_user_deleted_idx ON user_collections (deleted_at);
CREATE INDEX user_collections_products_product_id_idx ON user_collection_products (product_id);
CREATE INDEX user_collections_products_collection_id_idx ON user_collection_products (collection_id);
CREATE INDEX user_collections_products_products_collections_id_idx ON user_collection_products (product_id, collection_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX IF EXISTS user_collections_user_id_idx;
DROP INDEX IF EXISTS user_collections_user_deleted_idx;
DROP INDEX IF EXISTS user_collections_products_product_id_idx;
DROP INDEX IF EXISTS user_collections_products_collection_id_idx;
DROP INDEX IF EXISTS user_collections_products_products_collections_id_idx;

DROP TABLE IF EXISTS user_collection_products;
DROP TABLE IF EXISTS user_collections;