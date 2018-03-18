
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE MATERIALIZED VIEW search_words AS
SELECT word, ndoc FROM
    ts_stat('SELECT to_tsvector(''simple'', title) FROM products')
    WHERE ndoc > 1 AND to_tsvector(word) <> '';

-- PERIODICALLY RUN
-- `REFRESH MATERIALIZED VIEW search_words;`

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP MATERIALIZED VIEW IF EXISTS search_words;
