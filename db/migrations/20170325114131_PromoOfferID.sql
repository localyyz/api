
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos ADD COLUMN offer_id bigint DEFAULT 0 NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE promos DELETE COLUMN offer_id;
