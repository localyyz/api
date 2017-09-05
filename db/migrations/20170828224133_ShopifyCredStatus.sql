
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shopify_creds ADD COLUMN status smallint DEFAULT 0 NOT NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE shopify_creds DROP COLUMN status;

