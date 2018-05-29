-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE MATERIALIZED VIEW search_words AS
SELECT word, ndoc FROM
    ts_stat('SELECT to_tsvector(''simple'', title) FROM products')
    WHERE ndoc > 1 AND to_tsvector(word) <> '';

-- NOTE: MUST RUN `CREATE EXTENSION pg_trgm;` 

CREATE INDEX words_idx ON search_words USING GIN (word gin_trgm_ops);

-- PERIODICALLY RUN
-- `REFRESH MATERIALIZED VIEW search_words;`

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP MATERIALIZED VIEW IF EXISTS search_words;
