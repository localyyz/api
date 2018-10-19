-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE deal_metrics (
    id SERIAL PRIMARY KEY,
    deal_id bigint REFERENCES deals (id) ON DELETE CASCADE,
    product_id bigint REFERENCES products (id) ON DELETE CASCADE,
    product_name text DEFAULT '' NOT NULL,
    clicks smallInt DEFAULT 0.0 NOT NULL,
    views smallint DEFAULT 0.0 NOT NULL,
    copies smallInt DEFAULT 0.0 NOT NULL,
    add_to_cart smallInt DEFAULT 0.0 NOT NULL,
    claimed smallInt DEFAULT 0.0 NOT NULL,
    week_day text DEFAULT '' NOT NULL,
    time timestamp,
    gender text DEFAULT '' NOT NULL,
    category text DEFAULT '' NOT NULL,
    value smallInt DEFAULT '' NOT NULL,
    brand text DEFAULT '' NOT NULL,
);

CREATE INDEX metric_deal_idx on deal_metrics (deal_id);
CREATE INDEX metric_product_idx on deal_metrics (product_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS deal_metrics;
