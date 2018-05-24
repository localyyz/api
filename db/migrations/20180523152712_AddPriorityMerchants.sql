
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE priority_merchants (
    id INT NOT NULL
);
CREATE INDEX priority_merchant_index ON priority_merchants (id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX IF EXISTS priority_merchants.priority_merchant_index;
DROP TABLE IF EXISTS priority_merchants;
