
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE cells (
    id bigint PRIMARY KEY,
    locale_id bigint REFERENCES promos (id) ON DELETE CASCADE
);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE cells;
