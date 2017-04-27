
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shopify_creds ADD COLUMN api_url text DEFAULT '' NOT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE shopify_creds DROP COLUMN api_url;
