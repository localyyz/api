
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE blacklist (
    word varchar(30) NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS blacklist;
