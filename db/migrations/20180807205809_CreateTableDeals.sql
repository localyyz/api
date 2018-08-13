
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE deals (
    id SERIAL PRIMARY KEY,
    external_id bigint NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    merchant_id bigint REFERENCES places (id) ON DELETE CASCADE NOT NULL,
    parent_id bigint REFERENCES deals (id) ON DELETE CASCADE,
    user_id bigint REFERENCES users (id) ON DELETE CASCADE,
    code text DEFAULT '' NOT NULL,
    value numeric(15,6) DEFAULT 0.0 NOT NULL,
    use_limit smallint DEFAULT 0.0 NOT NULL,
    use_once bool DEFAULT false NOT NULL,
    start_at timestamp,
    end_at timestamp 
);

ALTER TABLE ONLY deals
    ADD CONSTRAINT unique_deal_external UNIQUE (external_id);
ALTER TABLE ONLY deals
    ADD CONSTRAINT unique_user_deal UNIQUE (user_id, code);
CREATE INDEX deal_user_idx on deals (user_id) WHERE user_id IS NOT NULL; 
CREATE INDEX deal_parent_idx on deals (parent_id) WHERE parent_id IS NOT NULL; 
CREATE INDEX deal_merchant_idx on deals (merchant_id);

CREATE TABLE deal_products (
    deal_id bigint REFERENCES deals (id) ON DELETE CASCADE NOT NULL,
    product_id bigint REFERENCES products (id) ON DELETE CASCADE NOT NULL
);

CREATE INDEX deal_products_product_idx on deal_products (product_id); 
CREATE INDEX deal_products_deal_idx on deal_products (deal_id); 

ALTER TABLE ONLY deal_products
    ADD CONSTRAINT unique_deal_product UNIQUE (deal_id, product_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS deal_products;
DROP TABLE IF EXISTS deals;
