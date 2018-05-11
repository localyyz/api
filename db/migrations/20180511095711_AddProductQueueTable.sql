
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE product_queue (
    id INT PRIMARY KEY NOT NULL,
    image_url text NOT NULL,
    tags text NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS product_queue;
