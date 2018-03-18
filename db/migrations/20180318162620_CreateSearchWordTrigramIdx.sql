-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- NOTE: MUST RUN `CREATE EXTENSION pg_trgm;` 

CREATE INDEX words_idx ON search_words USING GIN (word gin_trgm_ops);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
