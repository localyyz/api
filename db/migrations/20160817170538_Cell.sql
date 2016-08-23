
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE cells (
    id serial PRIMARY KEY,
    cell_id bigint NOT NULL,
    locale_id bigint REFERENCES locales (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX cells_cellid_localeid_unique_idx ON cells USING btree (cell_id, locale_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE cells;
